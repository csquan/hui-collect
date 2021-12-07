package full_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type recyclingHandler struct {
	db   types.IDB
	conf *config.Config
}

type info struct {
	symbol    string //HBTC
	chainID   string
	chainName string
}

const (
	chainBsc  = "bsc" //TODO 此处确保与配置中的chainName一致
	chainPoly = "polygon"
	chainHeco = "heco"
)

func (r *recyclingHandler) Name() string{
	return "recycling_handler"
}

func (r *recyclingHandler) Do(task *types.FullReBalanceTask) (err error) {
	res, err := getLpData(r.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	//根据接口返回的threshold构建(chainId,tokenSymbol)->tokenAddress 的mapping，后面获取tokenAddr读此map
	m := make(map[string]string)
	for _, threshold := range res.Thresholds {
		m[fmt.Sprintf("%d,%s", threshold.ChainId, threshold.TokenSymbol)] = threshold.TokenAddress
	}
	var sendToBridgeParams []*types.SendToBridgeParam
	var receiveFromBridgeParams []*types.ReceiveFromBridgeParam
	partRebalanceParam := &types.Params{ReceiveFromBridgeParams: receiveFromBridgeParams, SendToBridgeParams: sendToBridgeParams}
	for _, vault := range res.VaultInfoList {
		bscSend := r.buildSendToBridgeParam(vault, chainBsc)
		polySend := r.buildSendToBridgeParam(vault, chainPoly)
		var receiveBsc, receivePoly *types.ReceiveFromBridgeParam
		receiveBsc, err = r.buildReceiveFromBridgeParam(vault, chainHeco, chainBsc, m, vault.TokenSymbol)
		if err != nil{
			logrus.Errorf("buildReceiveFromBridgeParam err:%v", err)
			return
		}
		receivePoly, err = r.buildReceiveFromBridgeParam(vault, chainHeco, chainPoly, m, vault.TokenSymbol)
		if err != nil{
			logrus.Errorf("buildReceiveFromBridgeParam err:%v", err)
			return
		}
		sendToBridgeParams = append(sendToBridgeParams, bscSend, polySend)
		receiveFromBridgeParams = append(receiveFromBridgeParams, receiveBsc, receivePoly)
	}
	data, _ := json.Marshal(partRebalanceParam)
	partTask := &types.PartReBalanceTask{
		Base:   &types.Base{},
		Params: string(data),
	}
	err = utils.CommitWithSession(r.db, func(session *xorm.Session) (execErr error) {
		execErr = r.db.SaveRebalanceTask(session, partTask)
		if execErr != nil {
			task.State = types.FullReBalanceRecycling
			execErr = r.db.UpdateFullReBalanceTask(session, task)
		}
		return
	})
	return
}

func (r *recyclingHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	partTask, err := r.db.GetPartReBalanceTaskByFullRebalanceID(task.ID)
	if err != nil{
		logrus.Errorf("GetPartReBalanceTaskByFullRebalanceID err:%v", err)
		return
	}
	if partTask == nil{
		err = fmt.Errorf("GetPartReBalanceTaskByFullRebalanceID err:%v", err)
		logrus.Error(err)
		return
	}
	switch partTask.State {
	case types.PartReBalanceSuccess:
		return true, types.FullReBalanceParamsCalc, nil
	case types.PartReBalanceFailed:
		return true, types.FullReBalanceFailed, nil
	default:
		finished = false
		return
	}
}

func getLpData(url string) (lpList *types.Data, err error) {
	data, err := utils.DoRequest(url, "GET", nil)
	if err != nil {
		logrus.Errorf("request lp err:%v", err)
		return
	}
	lpResponse := &types.LPResponse{}
	if err = json.Unmarshal(data, lpResponse); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v", err)
		return
	}
	if lpResponse.Code != 200 {
		logrus.Errorf("lpResponse code not 200, msg:%s", lpResponse.Msg)
		return
	}
	lpList = lpResponse.Data
	return
}

func (r *recyclingHandler) buildSendToBridgeParam(vault *types.VaultInfo, chainName string) (sendParam *types.SendToBridgeParam) {
	sendParam = &types.SendToBridgeParam{}
	chain, ok := r.conf.Chains[chainName]
	if !ok {
		logrus.Fatalf("can not find chain config for %s", chainName)
	}
	sendParam.ChainId = chain.ID
	sendParam.ChainName = chainName
	sendParam.From = chain.BridgeAddress
	sendParam.To = vault.ActiveAmount.BSC.ControllerAddress

	sendParam.BridgeAddress = common.HexToAddress(chain.BridgeAddress)
	switch chainName {
	case chainBsc:
		sendParam.Amount = vault.ActiveAmount.BSC.Amount
	case chainPoly:
		sendParam.Amount = vault.ActiveAmount.Polygon.Amount
	default:
		logrus.Fatalf("buildSendToBridgeParam err chainName:%s", chainName)
	}
	sendParam.TaskID = "1" //TODO
	return
}
func (r *recyclingHandler) buildReceiveFromBridgeParam(vault *types.VaultInfo, chainName string,
	fromChain string, m map[string]string, tokenSymbol string) (sendParam *types.ReceiveFromBridgeParam, err error) {
	sendParam = &types.ReceiveFromBridgeParam{}
	chain, ok := r.conf.Chains[chainName]
	if !ok {
		logrus.Fatalf("can not find chain config for %s", chainName)
	}
	sendParam.ChainId = chain.ID
	sendParam.ChainName = chainName
	sendParam.From = chain.BridgeAddress
	sendParam.To = vault.ActiveAmount.Heco.ControllerAddress
	tokenAddr, ok := m[fmt.Sprintf("%d, %s", chain.ID, tokenSymbol)]
	if !ok {
		err = fmt.Errorf("not found tokenAddr chain.ID:%d,tokenSymbol:%s", chain.ID, tokenSymbol)
		return
	}
	sendParam.Erc20ContractAddr = common.HexToAddress(tokenAddr)

	switch fromChain {
	case chainBsc:
		sendParam.Amount = vault.ActiveAmount.BSC.Amount
	case chainPoly:
		sendParam.Amount = vault.ActiveAmount.Polygon.Amount
	default:
		logrus.Fatalf("buildSendToBridgeParam err chainName:%s", chainName)
	}
	sendParam.TaskID = "1" //TODO
	return
}
