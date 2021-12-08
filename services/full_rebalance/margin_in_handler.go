package full_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type impermanenceLostHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *impermanenceLostHandler) Name() string {
	return "full_rebalance_impermanenceLost"
}

func (i *impermanenceLostHandler) Do(task *types.FullReBalanceTask) (err error) {
	lpList, err := getLp(i.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	lpReq, err := lp2Req(lpList)
	if err != nil{
		logrus.Errorf("build margin_in params err:%v", err)
		return
	}
	if err = callImpermanentLoss(i.conf.ApiConf.MarginUrl,
		&types.ImpermanectLostReq{BizNo: fmt.Sprintf("%d", task.ID), LpList: lpReq}); err != nil {
		return
	}
	task.State = types.FullReBalanceMarginIn
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}

func (i *impermanenceLostHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	finished, err = checkMarginJobStatus(i.conf.ApiConf.MarginUrl, fmt.Sprintf("%d", task.ID))
	if err != nil {
		return
	}
	return true, types.FullReBalanceClaimLP, nil
}

func checkMarginJobStatus(url string, bizNo string) (finished bool, err error) {
	req := struct {
		BizNo string `json:"bizNo"`
	}{bizNo}
	data, err := utils.DoRequest(url+"status/query", "POST", req)
	if err != nil {
		logrus.Errorf("request ImpermanentLoss api err:%v", err)
		return
	}
	resp := &types.NormalResponse{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v", err)
		return
	}
	if resp.Code != 200 {
		logrus.Errorf("callImpermanentLoss code not 200, msg:%s", resp.Msg)
		return
	}
	if v, ok := resp.Data["status"]; ok {
		return v.(string) == "SUCCESS", nil
	}
	return
}

func getLp(url string) (lpList []*types.LiquidityProvider, err error) {
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
	lpList = lpResponse.Data.LiquidityProviderList
	return
}
func callImpermanentLoss(url string, req *types.ImpermanectLostReq) (err error) {
	data, err := utils.DoRequest(url+"submit","POST", req)
	if err != nil {
		logrus.Errorf("request ImpermanentLoss api err:%v", err)
		return
	}
	resp := &types.NormalResponse{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v", err)
		return
	}
	if resp.Code != 200 {
		logrus.Errorf("callImpermanentLoss code not 200, msg:%s", resp.Msg)
		return
	}
	return
}

func lp2Req(lpList []*types.LiquidityProvider) (req []*types.LpReq, err error) {
	for _, lp := range lpList {
		totalBaseAmount := decimal.Zero
		totalQuoteAmount := decimal.Zero
		for _, lpinfo := range lp.LpInfoList {
			var baseAmount, quoteAmount decimal.Decimal
			baseAmount, err = decimal.NewFromString(lpinfo.BaseTokenAmount)
			if err != nil {
				logrus.Errorf("BaseTokenAmount to decimal err:%v", err)
				return nil, err
			}
			quoteAmount, err = decimal.NewFromString(lpinfo.QuoteTokenAmount)
			if err != nil {
				logrus.Errorf("QuoteTokenAmount to decimal err:%v", err)
				return nil, err
			}
			totalBaseAmount = totalBaseAmount.Add(baseAmount)
			totalQuoteAmount = totalQuoteAmount.Add(quoteAmount)
		}
		r := &types.LpReq{
			Chain:              lp.Chain,
			LpTokenAddress:     lp.LpTokenAddress,
			LpAmount:           lp.LpAmount,
			Token0OriginAmount: totalBaseAmount.String(),
			Token1OriginAmount: totalQuoteAmount.String(),
		}
		req = append(req, r)
	}
	return
}

