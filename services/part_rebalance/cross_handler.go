package part_rebalance

import (
	"encoding/json"
	"fmt"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type crossHandler struct {
	db types.IDB
}

func (c *crossHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	crossTasks, err := c.db.GetCrossTasksByReBalanceId(task.ID)
	if err != nil {
		logrus.Errorf("get cross task for rebalance [%v] failed", task)
		return
	}

	if len(crossTasks) == 0 {
		err = fmt.Errorf("part rebalance task [%v] has no cross task", task)
		return
	}

	success := true
	for _, crossTask := range crossTasks {
		if crossTask.State != types.TaskSuc {
			logrus.Debugf("cross task [%v] is not finished", crossTask)

			return
		}

		success = success && crossTask.State == types.TaskSuc
	}

	if success {
		nextState = types.PartReBalanceTransferIn
	} else {
		nextState = types.PartReBalanceFailed
	}

	return true, nextState, nil
}

func (c *crossHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	err = utils.CommitWithSession(c.db, func(session *xorm.Session) (execErr error) {

		//create next state task
		if nextState == types.PartReBalanceTransferIn {
			var assetTransfer *types.AssetTransferTask
			assetTransfer, execErr = c.createTransferInTask(task)
			if execErr != nil {
				return
			}

			execErr = c.db.SaveAssetTransferTask(session, assetTransfer)
			if execErr != nil {
				logrus.Errorf("save assetTransfer task error:%v task:[%v]", execErr, task)
				return
			}
		}

		//move to next state
		task.State = nextState
		execErr = c.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}

func (c *crossHandler) createTransferInTask(task *types.PartReBalanceTask) (assetTransfer *types.AssetTransferTask, err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}

	assetTransferParams, err := json.Marshal(params.AssetTransferIn)
	if err != nil {
		logrus.Errorf("marshal AssetTransferInParams params error:%v task:[%v]", err, task)
		return
	}

	assetTransfer = &types.AssetTransferTask{
		BaseTask:     &types.BaseTask{State: types.AssetTransferInit},
		RebalanceId:  task.ID,
		TransferType: types.AssetTransferIn,
		Params:       string(assetTransferParams),
	}

	return
}
