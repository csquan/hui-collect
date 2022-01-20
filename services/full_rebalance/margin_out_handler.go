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
	if task.Params == "" {
		//params==""说明之前没有调用margin in，直接跳转到下一状态。
		return moveState(i.db, task, types.FullReBalanceMarginBalanceTransferOut, nil)
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
	return moveState(i.db, task, types.FullReBalanceMarginBalanceTransferOut, nil)
}

func (i *marginOutHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	//if task.Params == "" {
	//	return true, types.FullReBalanceRecycling, nil
	//}
	//urlStr, err := joinUrl(i.conf.ApiConf.MarginOutUrl, "status/query")
	//if err != nil {
	//	logrus.Errorf("parse url error:%v", err)
	//	return
	//}
	//
	//req := struct {
	//	BizNo string `json:"bizNo"`
	//}{BizNo: fmt.Sprintf("%d", task.ID)}
	//resp, err := GetMarginJobStatus(urlStr, i.conf, req)
	//if err != nil {
	//	return
	//}
	//status, ok := resp.Data["status"]
	//if !ok || status.(string) != "SUCCESS" {
	//	return
	//}
	utils.GetFullReCost(task.ID).AppendReport("保证金转出，不等待结果")
	return true, types.FullReBalanceRecycling, nil
}

func (i *marginOutHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# fullMarginOut
	- taskID: %d
	`, taskId)
}
