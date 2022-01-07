package full_rebalance

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-xorm/xorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type recyclingHandler struct {
	db   types.IDB
	conf *config.Config
}

func (r *recyclingHandler) Name() string {
	return "recycling_handler"
}

func (r *recyclingHandler) Do(task *types.FullReBalanceTask) (err error) {
	res, err := getLpData(r.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	//确保拆LP已完成
	if res.LiquidityProviderList != nil && len(res.LiquidityProviderList) > 0 {
		logrus.Infof("LiquidityProviderList is not nil, cannot do recycling")
		return errors.New("LiquidityProviderList is not nil, cannot do recycling") //返回err，避免重复发送钉钉通知
	}
	tokens, err1 := r.db.GetTokens()
	if err1 != nil {
		return fmt.Errorf("get tokens err:%v", err1)
	}
	currencyList, err2 := r.db.GetCurrency()
	if err2 != nil {
		return fmt.Errorf("get currency err:%v", err2)
	}
	partRebalanceParam := &types.Params{
		SendToBridgeParams:      make([]*types.SendToBridgeParam, 0),
		CrossBalances:           make([]*types.CrossBalanceItem, 0),
		ReceiveFromBridgeParams: make([]*types.ReceiveFromBridgeParam, 0),
		InvestParams:            make([]*types.InvestParam, 0),
	}
	for _, vault := range res.VaultInfoList {
		if err = r.appendParam(vault, partRebalanceParam, tokens, currencyList); err != nil {
			logrus.Warnf("recyclingHandler appendParam err:%v, res:%+v, task:%+v", err, res, task)
			return
		}
	}
	var b0, b1 []byte
	if res != nil {
		b0, _ = json.Marshal(res)
		b1, _ = json.Marshal(partRebalanceParam)
	}
	logrus.Infof("recyclingHandler appendParam res:%s, task:%+v, param:%s", b0, task, b1)
	data, _ := json.Marshal(partRebalanceParam)
	if len(data) > 65535 {
		return fmt.Errorf("part rebalance size is over 65535")
	}
	partTask := &types.PartReBalanceTask{
		Params:          string(data),
		FullRebalanceID: task.ID,
	}
	err = utils.CommitWithSession(r.db, func(session *xorm.Session) (execErr error) {
		execErr = r.db.SaveRebalanceTask(session, partTask)
		if execErr != nil {
			return
		}
		task.State = types.FullReBalanceRecycling
		execErr = r.db.UpdateFullReBalanceTask(session, task)
		if execErr != nil {
			return
		}
		return
	})
	return
}

func (r *recyclingHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	partTask, err := r.db.GetPartReBalanceTaskByFullRebalanceID(task.ID)
	if err != nil {
		logrus.Errorf("GetPartReBalanceTaskByFullRebalanceID err:%v", err)
		return
	}
	if partTask == nil {
		return true, types.FullReBalanceParamsCalc, nil
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

func (r *recyclingHandler) appendParam(vault *types.VaultInfo, partRebalanceParam *types.Params,
	tokens []*types.Token, currencyList []*types.Currency) (err error) {
	hecoChainName := "heco"
	hecoChain := mustGetChainInfo(hecoChainName, r.conf)
	hecoController, err := hecoController(vault, hecoChainName)
	if err != nil {
		return
	}
	for fromChainName, info := range vault.ActiveAmount {
		if strings.ToLower(fromChainName) == hecoChainName {
			continue
		}
		if info == nil || info.ControllerAddress == "" {
			continue
		}
		//根据chainName，从配置中获取bridgeAddress信息
		fromChain := mustGetChainInfo(fromChainName, r.conf)
		//Currency的f_min为null或者0则不参与跨回
		currency := mustGetCurrency(currencyList, vault.Currency)
		if currency.Min.Cmp(decimal.Zero) <= 0 {
			logrus.Infof("currency.min not config, currency:%v", vault.Currency)
			continue
		}
		//判断amount是否大于最小值
		var amount, soloAmount decimal.Decimal
		if info.VaultAmount == "" {
			return fmt.Errorf("vaultAmount empty valut addr:%s", info.ControllerAddress)
		}
		if amount, err = decimal.NewFromString(info.VaultAmount); err != nil {
			logrus.Errorf("convert vaultAmount to decimal err:%v", err)
			return
		}
		if info.SoloAmount == "" {
			return fmt.Errorf("soloAmount empty valut addr:%s", info.ControllerAddress)
		}
		if soloAmount, err = decimal.NewFromString(info.SoloAmount); err != nil {
			logrus.Errorf("convert rewart to dceimal err:%v,reward:%s", err, info.SoloAmount)
			return
		}
		amount = amount.Add(soloAmount)
		//amount = amount.Truncate(currency.CrossScale) 由跨链桥处理

		if amount.Cmp(currency.Min) == -1 {
			logrus.Infof("amount less than currency.min amount:%s, min:%s,vaultAddr:%s", amount.String(), currency.Min.String(), info.ControllerAddress)
			return
		}
		var fromToken, hecoToken *types.Token
		fromToken = mustGetToken(tokens, vault.Currency, fromChainName)
		hecoToken = mustGetToken(tokens, vault.Currency, hecoChainName)
		amountStr := powN(amount, fromToken.Decimal).String()
		taskID := fmt.Sprintf("%d", time.Now().UnixNano()/1000)
		sendParam := &types.SendToBridgeParam{
			ChainId:       fromChain.ID,
			ChainName:     fromChainName,
			From:          fromChain.BridgeAddress,
			To:            info.ControllerAddress,
			BridgeAddress: common.HexToAddress(fromChain.BridgeAddress),
			Amount:        amountStr,
			TaskID:        taskID,
		}
		crossParam := &types.CrossBalanceItem{
			FromChain:    fromChainName,
			ToChain:      hecoChainName,
			FromAddr:     fromChain.BridgeAddress,
			ToAddr:       hecoChain.BridgeAddress,
			FromCurrency: fromToken.CrossSymbol,
			ToCurrency:   hecoToken.CrossSymbol,
			Amount:       amount.String(),
		}
		receiveParam := &types.ReceiveFromBridgeParam{
			ChainId:           hecoChain.ID,
			ChainName:         hecoChainName,
			From:              hecoChain.BridgeAddress,
			To:                hecoController.ControllerAddress,
			Erc20ContractAddr: common.HexToAddress(hecoToken.Address),
			Amount:            amountStr,
			TaskID:            taskID,
		}
		partRebalanceParam.SendToBridgeParams = append(partRebalanceParam.SendToBridgeParams, sendParam)
		partRebalanceParam.CrossBalances = append(partRebalanceParam.CrossBalances, crossParam)
		partRebalanceParam.ReceiveFromBridgeParams = append(partRebalanceParam.ReceiveFromBridgeParams, receiveParam)
	}
	return
}

func mustGetToken(tokens []*types.Token, currency, chain string) (token *types.Token) {
	for _, token = range tokens {
		if token.Currency == strings.ToLower(currency) && token.Chain == strings.ToLower(chain) {
			return token
		}
	}
	logrus.Fatalf("can not find token from db, currency:%s, chain:%s", currency, chain)
	return
}

func mustGetCurrency(currencyList []*types.Currency, name string) (currency *types.Currency) {
	for _, currency = range currencyList {
		if currency.Name == strings.ToLower(name) {
			return currency
		}
	}
	logrus.Fatalf("can not find currency from db, name:%s", name)
	return
}

func hecoController(vault *types.VaultInfo, hecoChain string) (controller *types.ControllerInfo, err error) {
	for chainName, info := range vault.ActiveAmount {
		if strings.ToLower(chainName) == hecoChain {
			controller = info
			break
		}
	}
	if controller == nil {
		err = fmt.Errorf("heco controller not found, vault:%v", vault)
		return
	}
	return
}

func mustGetChainInfo(chainName string, conf *config.Config) *config.ChainInfo {
	chain, ok := conf.Chains[strings.ToLower(chainName)]
	if !ok {
		logrus.Fatalf("can not find chain from config, chain:%s", chainName)
	}
	return chain
}

func (r *recyclingHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# recycling
	- taskID: %d
	`, taskId)
}
