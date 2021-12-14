package full_rebalance

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type marginOutHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *marginOutHandler) Name() string {
	return "full_rebalance_marginOut"
}

func (i *marginOutHandler) Do(task *types.FullReBalanceTask) (err error) {
	req := &types.ImpermanectLostReq{}
	if task.Params == "" {
		//params==""说明之前没有调用margin in，直接跳转到下一状态。
		task.State = types.FullReBalanceMarginBalanceTransferOut
		err = i.db.UpdateFullReBalanceTask(i.db.GetEngine(), task)
		return
	}
	if err = json.Unmarshal([]byte(task.Params), req); err != nil {
		logrus.Errorf("createMarginOutJob unmarshal params err:%v", err)
		return
	}

	urlStr, err := joinUrl(i.conf.ApiConf.MarginOutUrl, "submit")
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}

	_, err = callMarginApi(urlStr, i.conf, req)
	if err != nil {
		logrus.Errorf("margin job query status err:%v", err)
		return
	}
	task.State = types.FullReBalanceMarginBalanceTransferOut
	err = i.db.UpdateFullReBalanceTask(i.db.GetEngine(), task)
	return
}

func (i *marginOutHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	if task.Params == "" {
		return true, types.FullReBalanceRecycling, nil
	}
	urlStr, err := joinUrl(i.conf.ApiConf.MarginOutUrl, "status/query")
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}

	req := struct {
		BizNo string `json:"bizNo"`
	}{BizNo: fmt.Sprintf("%d", task.ID)}
	resp, err := GetMarginJobStatus(urlStr, i.conf, req)
	if err != nil {
		return
	}
	status, ok := resp.Data["status"]
	if !ok || status.(string) != "SUCCESS" {
		return
	}
	return true, types.FullReBalanceRecycling, nil
}
