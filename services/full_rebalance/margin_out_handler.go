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
	req := &types.ImpermanectLostReq{}
	if err = json.Unmarshal([]byte(task.Params), req); err != nil {
		logrus.Errorf("createMarginOutJob unmarshal params err:%v", err)
		return
	}
	_, err = callMarginApi(i.conf.ApiConf.MarginOutUrl +"/submit", i.conf, req)
	if err != nil {
		logrus.Errorf("margin job query status err:%v", err)
		return
	}
	task.State = types.FullReBalanceMarginBalanceTransferOut
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}

func (i *marginOutHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	req := struct {
		BizNo string `json:"bizNo"`
	}{BizNo: fmt.Sprintf("%d", task.ID)}
	resp, err := callMarginApi(i.conf.ApiConf.MarginOutUrl+"status/query", i.conf, req)
	if err != nil {
		return
	}
	if v, ok := resp.Data["status"]; ok {
		if v.(string) != "SUCCESS" {
			return
		}
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

