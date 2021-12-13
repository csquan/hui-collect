package full_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"net/url"
	"path"
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
	if err != nil {
		logrus.Errorf("build margin_in params err:%v", err)
		return
	}
	bizNo := fmt.Sprintf("%d", task.ID)
	req := &types.ImpermanectLostReq{BizNo: bizNo, LpList: lpReq}
	urlStr, err := joinUrl(i.conf.ApiConf.MarginUrl, "submit")
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}

	if _, err = callMarginApi(urlStr, i.conf, req); err != nil {
		return
	}
	var params []byte
	if params, err = json.Marshal(req); err != nil {
		logrus.Errorf("marshal margin out params err:%v", err)
		return
	}
	task.Params = string(params) //save params for margin out
	task.State = types.FullReBalanceMarginIn
	err = i.db.UpdateFullReBalanceTask(i.db.GetEngine(), task)

	return
}

func (i *impermanenceLostHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	bizNo := fmt.Sprintf("%d", task.ID)
	urlStr, err := joinUrl(i.conf.ApiConf.MarginUrl, "status/query")
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}

	res, err := callMarginApi(urlStr, i.conf, struct {
		BizNo string `json:"bizNo"`
	}{bizNo})
	if err != nil {
		return
	}

	status, ok := res.Data["status"]
	if !ok || status.(string) != "SUCCESS" {
		return
	}
	return true, types.FullReBalanceClaimLP, nil
}

func joinUrl(urlInput string, pathInput string) (string, error) {
	u, err := url.Parse(urlInput)
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return "", err
	}

	u.Path = path.Join(pathInput, "submit")
	return u.String(), nil
}

func lp2Req(lpList []*types.LiquidityProvider) (req []*types.LpReq, err error) {
	for _, lp := range lpList {
		totalBaseAmount := decimal.Zero
		totalQuoteAmount := decimal.Zero
		totalLpAmount := decimal.Zero
		for _, lpinfo := range lp.LpInfoList {
			var baseAmount, quoteAmount, lpAmount decimal.Decimal
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
			lpAmount, err = decimal.NewFromString(lpinfo.LpAmount)
			if err != nil {
				logrus.Errorf("QuoteTokenAmount to decimal err:%v", err)
				return nil, err
			}
			totalBaseAmount = totalBaseAmount.Add(baseAmount)
			totalQuoteAmount = totalQuoteAmount.Add(quoteAmount)
			totalLpAmount = totalLpAmount.Add(lpAmount)
		}
		r := &types.LpReq{
			Chain:              lp.Chain,
			LpTokenAddress:     lp.LpTokenAddress,
			LpAmount:           totalLpAmount.String(),
			Token0OriginAmount: totalBaseAmount.String(),
			Token1OriginAmount: totalQuoteAmount.String(),
		}
		req = append(req, r)
	}
	return
}
