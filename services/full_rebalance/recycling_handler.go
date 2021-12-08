package full_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-xorm/xorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"strings"
)

type recyclingHandler struct {
	db   types.IDB
	conf *config.Config
}

func (r *recyclingHandler) Name() string {
	return "recycling_handler"
}

func (r *recyclingHandler) Do(task *types.FullReBalanceTask) (err error) {
	res, err := lpData(r.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	tokensMap, err := tokensMap(r.db)
	currencyMap, err := r.currencyMap()
	if err != nil {
		return
	}
	partRebalanceParam := &types.Params{
		SendToBridgeParams:      make([]*types.SendToBridgeParam, 0),
		CrossBalances:           make([]*types.CrossBalanceItem, 0),
		ReceiveFromBridgeParams: make([]*types.ReceiveFromBridgeParam, 0),
	}
	for _, vault := range res.VaultInfoList {
		if err = r.appendParam(vault, partRebalanceParam, tokensMap, currencyMap); err != nil{
			return
		}
	}
	data, _ := json.Marshal(partRebalanceParam)
	partTask := &types.PartReBalanceTask{
		Base:   &types.Base{},
		Params: string(data),
	}
	err = utils.CommitWithSession(r.db, func(session *xorm.Session) (execErr error) {
		execErr = r.db.SaveRebalanceTask(session, partTask)
		if execErr != nil {
			return
		}
		task.State = types.FullReBalanceRecycling
		execErr = r.db.UpdateFullReBalanceTask(session, task)
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
		err = fmt.Errorf("not found part rebalance task")
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

func lpData(url string) (lpList *types.Data, err error) {
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

func (r *recyclingHandler) appendParam(vault *types.VaultInfo, partRebalanceParam *types.Params,
	tokensMap map[string]*types.Token, currencyMap map[string]*types.Currency) (err error) {
	heco, err := r.hecoController(vault)
	if err != nil {
		return
	}
	for fromChainName, info := range vault.ActiveAmount {
		if strings.ToLower(fromChainName) == "heco" {
			continue
		}
		//根据chainName，从配置中获取bridgeAddress信息
		fromChain, ok := r.conf.Chains[strings.ToLower(fromChainName)]
		if !ok {
			err = fmt.Errorf("can not find chain config for %s", fromChainName)
			return
		}
		//判断amount是否大于最小值
		currency, ok := currencyMap[vault.Currency]
		if !ok {
			logrus.Warnf("not found currency from db,currency:%s", vault.Currency)
			return
		}
		if currency.Min.Cmp(decimal.Zero) <= 0 {
			logrus.Infof("currency.min not config, currency:%v", vault.Currency)
			return
		}
		var amount decimal.Decimal
		if amount, err = decimal.NewFromString(info.Amount); err != nil {
			logrus.Errorf("convert amount to decimal err:%v", err)
			return
		}
		if amount.Cmp(currency.Min) == -1 {
			logrus.Infof("amount less than currency.min amount:%s, min:%s", amount.String(), currency.Min.String())
			return
		}
		var fromToken, hecoToken *types.Token
		fromToken, ok = getToken(tokensMap, vault.Currency, info.Chain)
		hecoToken, ok = getToken(tokensMap, vault.Currency, heco.Chain)
		amountStr := powN(strMustToDecimal(info.Amount), fromToken.Decimal).String()
		taskID := "1" // TODO
		sendParam := &types.SendToBridgeParam{
			ChainId:       fromChain.ID,
			ChainName:     fromChainName,
			From:          fromChain.BridgeAddress,
			To:            info.ControllerAddress,
			BridgeAddress: common.HexToAddress(fromChain.BridgeAddress),
			Amount:        amountStr,
			TaskID:        taskID,        
		}
		crossParam := &types.CrossBalanceItem{ //USDT
			FromChain:    fromChainName,
			ToChain:      heco.Chain,
			FromAddr:     fromChain.BridgeAddress,
			ToAddr:       heco.BridgeAddress,
			FromCurrency: fromToken.CrossSymbol,
			ToCurrency:   hecoToken.CrossSymbol,
			Amount:       amountStr,
		}
		receiveParam := &types.ReceiveFromBridgeParam{ //USDT
			ChainId:           heco.ChainID,
			ChainName:         heco.Chain,
			From:              heco.BridgeAddress,
			To:                heco.ControllerAddress,
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

func getToken(tokensMap map[string]*types.Token, currency, chain string) (*types.Token, bool) {
	key := fmt.Sprintf("%s,%s", currency, chain)
	token, ok := tokensMap[strings.ToLower(key)]
	return token, ok
}

func tokensMap(db types.IDB) (map[string]*types.Token, error) {
	tokens, err := db.GetTokens()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*types.Token)
	for _, token := range tokens {
		key := fmt.Sprintf("%s,%s", token.Currency, token.Chain)
		m[strings.ToLower(key)] = token
	}
	return m, nil
}

func (r *recyclingHandler) currencyMap() (map[string]*types.Currency, error) {
	list, err := r.db.GetCurrency()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*types.Currency)
	for _, currency := range list {
		m[currency.Name] = currency
	}
	return m, nil
}

func (r *recyclingHandler) hecoController(vault *types.VaultInfo) (controller *types.ControllerInfo, err error) {
	hecoChainName := "heco"
	for chainName, info := range vault.ActiveAmount {
		if strings.ToLower(chainName) == hecoChainName {
			controller = info
			break
		}
	}
	if controller == nil {
		err = fmt.Errorf("heco controller not found, vault:%v", vault)
		return
	}
	chain, ok := r.conf.Chains[hecoChainName]
	if !ok {
		err = fmt.Errorf("can not find chain config for %s", hecoChainName)
		return
	}
	controller.Chain = hecoChainName
	controller.ChainID = chain.ID
	controller.BridgeAddress = chain.BridgeAddress
	return
}
