package full_rebalance

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type initHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *initHandler) Name() string {
	return "full_rebalance_init"
}

func (i *initHandler) Do(task *types.FullReBalanceTask) error {
	return nil
}

func (i *initHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	resp, err := utils.CallTaskManager(i.conf, fmt.Sprintf(`/v1/open/task/begin/Full_%d?taskType=rebalance`, task.ID), "POST")
	if err != nil || !resp.Data {
		logrus.Infof("call task manager begin resp:%v, err：%v", resp, err)
		return
	}
	return true, types.FullReBalanceMarginIn, nil
}

func (i *initHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

