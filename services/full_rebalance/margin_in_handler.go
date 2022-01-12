package full_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"net/url"
	"path"
)

type impermanenceLostHandler struct {
	db            types.IDB
	conf          *config.Config
	alertedTaskID uint64 //避免重复报警
}

func (i *impermanenceLostHandler) Name() string {
	return "full_rebalance_margin_in"
}

func (i *impermanenceLostHandler) Do(task *types.FullReBalanceTask) (err error) {
	lpData, err := getLpData(i.conf.ApiConf.LpUrl)
	if err != nil {
		return
	}
	if lpData.LiquidityProviderList == nil || len(lpData.LiquidityProviderList) == 0 {
		return moveState(i.db, task, types.FullReBalanceMarginIn, lpData)
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
	return moveState(i.db, task, types.FullReBalanceMarginIn, lpData)
}

func (i *impermanenceLostHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	if task.Params == "" {
		utils.GetFullReCost(task.ID).AppendReport("平无常")
		return true, types.FullReBalanceClaimLP, nil
	}
	bizNo := fmt.Sprintf("%d", task.ID)
	urlStr, err := joinUrl(i.conf.ApiConf.MarginUrl, "status/query")
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}
	res, err := GetMarginJobStatus(urlStr, i.conf, struct {
		BizNo string `json:"bizNo"`
	}{bizNo})
	if err != nil {
		return
	}
	status, ok := res.Data["status"]
	if !ok {
		return
	}
	if status.(string) == "SUCCESS" {
		utils.GetFullReCost(task.ID).AppendReport("平无常")
		return true, types.FullReBalanceClaimLP, nil
	}
	if status.(string) == "FAILED" {
		alert.Dingding.SendAlert("Full Rebalance Failed", alert.TaskFailedContent("大Re", task.ID, "marginIn", fmt.Errorf("magin in failed")), nil)
		return true, types.FullReBalanceFailed, nil
	}
	return
}

func joinUrl(urlInput string, pathInput string) (string, error) {
	u, err := url.Parse(urlInput)
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return "", err
	}

	u.Path = path.Join(u.Path, pathInput)
	return u.String(), nil
}

func lp2Req(lpList []*types.LiquidityProvider) (req []*types.LpReq, err error) {
	for _, lp := range lpList {
		totalBaseAmount := decimal.Zero
		totalQuoteAmount := decimal.Zero
		totalLpAmount := decimal.Zero
		baseTokenAdress := ""
		quoteTokenAddress := ""
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
			baseTokenAdress = lpinfo.BaseTokenAddress
			quoteTokenAddress = lpinfo.QuoteTokenAddress
		}
		var tokenList []*types.TokenInfo
		tokenList = append(tokenList,
			&types.TokenInfo{TokenAddress: baseTokenAdress, TokenOriginAmount: totalBaseAmount.String()},
			&types.TokenInfo{TokenAddress: quoteTokenAddress, TokenOriginAmount: totalQuoteAmount.String()})
		r := &types.LpReq{
			Chain:          lp.Chain,
			LpTokenAddress: lp.LpTokenAddress,
			LpAmount:       totalLpAmount.String(),
			TokenList:      tokenList,
		}
		req = append(req, r)
	}
	return
}

// GetOpenedTaskMsg full_rebalance task_id作为bizNo
func (i *impermanenceLostHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# full_margin_in_runtimeout
	- bizNo: %d
	`, taskId)
}
