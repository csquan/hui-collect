package full_rebalance

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type marginOutHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *marginOutHandler) Name() string {
	return "full_rebalance_marginOut"
}

func (i *marginOutHandler) Do(task *types.FullReBalanceTask) (err error) {
	if err = createMarginOutJob(i.conf.ApiConf.MarginOutUrl, fmt.Sprintf("%d", task.ID)); err != nil {
		return
	}
	task.State = types.FullReBalanceMarginBalanceTransferOut
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}

func (i *marginOutHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	finished, err = checkMarginOutJobStatus(i.conf.ApiConf.MarginOutUrl+"status/query", fmt.Sprintf("%d", task.ID), i.conf)
	if err != nil {
		return
	}
	return true, types.FullReBalanceRecycling, nil
}

func createMarginOutJob(url string, params string) (err error) {
	req := &types.ImpermanectLostReq{}
	if err = json.Unmarshal([]byte(params), req); err != nil {
		logrus.Errorf("createMarginOutJob unmarshal params err:%v", err)
		return
	}
	data, err := utils.DoRequest(url, "POST", req)
	if err != nil {
		logrus.Errorf("margin job query status err:%v", err)
	}
	resp := &types.NormalResponse{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v", err)
		return
	}
	if resp.Code != 200 {
		logrus.Errorf("callImpermanentLoss code not 200, msg:%s", resp.Msg)
	}
	return
}

func checkMarginOutJobStatus(url string, bizNo string, conf *config.Config) (finished bool, err error) {
	req := struct {
		BizNo string `json:"bizNo"`
	}{BizNo: bizNo}
	resp, err := callMarginApi(url, conf, req)
	if err != nil {
		logrus.Errorf("margin out query status err:%v", err)
	}
	if resp.Code != 200 {
		logrus.Errorf("callMarginApi code not 200, msg:%s", resp.Msg)
		return
	}
	if v, ok := resp.Data["status"]; ok {
		return v.(string) == "SUCCESS", nil
	}
	return
}
