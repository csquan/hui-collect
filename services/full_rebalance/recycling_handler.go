package full_rebalance

import (
	"encoding/json"
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

func (r *recyclingHandler) Do(task *types.FullReBalanceTask) (err error) {
	vaultList, err := getVaultList(r.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	var sendToBridgeParams []*types.SendToBridgeParam
	var receiveFromBridgeParams []*types.ReceiveFromBridgeParam
	partRebalanceParam := &types.Params{ReceiveFromBridgeParams: receiveFromBridgeParams, SendToBridgeParams: sendToBridgeParams}
	for _, vault := range vaultList {
		bscSend := r.buildSendToBridgeParam(vault, chainBsc)
		polySend := r.buildSendToBridgeParam(vault, chainPoly)
		receiveBsc := r.buildReceiveFromBridgeParam(vault, chainHeco, chainBsc)
		receivePoly := r.buildReceiveFromBridgeParam(vault, chainHeco, chainPoly)
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

func (r *recyclingHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {

	return true, types.FullReBalanceParamsCalc, nil
}

func getVaultList(url string) (lpList []*types.VaultInfo, err error) {
	data, err := utils.DoGet(url, nil)
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
	lpList = lpResponse.Data.VaultInfoList
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
func (r *recyclingHandler) buildReceiveFromBridgeParam(vault *types.VaultInfo, chainName string, fromChain string) (sendParam *types.ReceiveFromBridgeParam) {
	sendParam = &types.ReceiveFromBridgeParam{}
	chain, ok := r.conf.Chains[chainName]
	if !ok {
		logrus.Fatalf("can not find chain config for %s", chainName)
	}
	sendParam.ChainId = chain.ID
	sendParam.ChainName = chainName
	sendParam.From = chain.BridgeAddress
	sendParam.To = vault.ActiveAmount.Heco.ControllerAddress

	sendParam.Erc20ContractAddr = common.HexToAddress("TODO") //TODO

	switch fromChain {
	case chainBsc:
		sendParam.Amount = vault.ActiveAmount.BSC.Amount
	case chainPoly:
		sendParam.Amount = vault.ActiveAmount.Polygon.Amount
	default:
		logrus.Fatalf("buildSendToBridgeParam err chainName:%s", chainName)
	}
	sendParam.Amount = vault.ActiveAmount.Polygon.Amount
	sendParam.TaskID = "1" //TODO
	return
}
