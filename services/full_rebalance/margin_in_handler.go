package full_rebalance

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type impermanenceLostHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *impermanenceLostHandler) Name() string {
	return "full_rebalance_margin_in"
}

func (i *impermanenceLostHandler) Do(task *types.FullReBalanceTask) (err error) {
	lpData, err := getLpData(i.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	lpReq, err := lp2Req(lpData.LiquidityProviderList)
	if err != nil{
		logrus.Errorf("build margin_in params err:%v", err)
		return
	}
	if _, err = callMarginApi(i.conf.ApiConf.MarginUrl + "submit", i.conf,
		&types.ImpermanectLostReq{BizNo: fmt.Sprintf("%d", task.ID), LpList: lpReq}); err != nil {
		return
	}
	task.State = types.FullReBalanceMarginIn
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}

func (i *impermanenceLostHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	bizNo := fmt.Sprintf("%d", task.ID)
	res, err := callMarginApi(i.conf.ApiConf.MarginUrl + "status/query", i.conf, struct {
		BizNo string `json:"bizNo"`
	}{bizNo})
	if err != nil {
		return
	}
	if v, ok := res.Data["status"]; ok {
		if v.(string) != "SUCCESS" {
			return
		}
	}
	return true, types.FullReBalanceClaimLP, nil
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

