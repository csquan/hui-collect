package full_rebalance

import (
	"fmt"
	"github.com/starslabhq/hermes-rebalance/utils"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
)

type doPartRebalanceHandler struct {
	db types.IDB
}

func (d *doPartRebalanceHandler) Name() string {
	return "doPartRebalanceHandler"
}

func (d *doPartRebalanceHandler) Do(task *types.FullReBalanceTask) (err error) {
	return nil
}

func (d *doPartRebalanceHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	partTask, err := d.db.GetPartReBalanceTaskByFullRebalanceID(task.ID)
	if err != nil {
		logrus.Errorf("GetPartReBalanceTaskByFullRebalanceID err:%v", err)
		return
	}
	if partTask == nil {
		err = fmt.Errorf("GetPartReBalanceTaskByFullRebalanceID err:%v", err)
		logrus.Error(err)
		return
	}
	switch partTask.State {
	case types.PartReBalanceSuccess:
		utils.GetFullReCost(task.ID).AppendReport("正向小R")
		return true, types.FullReBalanceSuccess, nil
	case types.PartReBalanceFailed:
		utils.GetFullReCost(task.ID).AppendReport("正向小R")
		return true, types.FullReBalanceFailed, nil
	default:
		finished = false
		return
	}
}

func (d *doPartRebalanceHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# fullRebalance doPartRebalance
	- taskID: %d
	`, taskId)
}
