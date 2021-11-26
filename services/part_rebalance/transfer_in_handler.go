package part_rebalance

import (
	"encoding/json"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type transferInHandler struct {
	db types.IDB
}

func (t *transferInHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransferState(t.db, task, types.AssetTransferIn)
	if err != nil {
		return
	}

	if state != types.AssetTransferSuccess && state != types.AssetTransferFailed {
		return
	}

	finished = true

	if state == types.AssetTransferSuccess {
		nextState = types.PartReBalanceInvest
	} else {
		nextState = types.PartReBalanceFailed
	}

	return
}

func (t *transferInHandler) createInvestTask(task *types.PartReBalanceTask) (assetTransfer *types.AssetTransferTask, err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}

	investParams, err := json.Marshal(params.Invest)
	if err != nil {
		logrus.Errorf("marshal AssetTransferInParams params error:%v task:[%v]", err, task)
		return
	}

	assetTransfer = &types.AssetTransferTask{
		BaseTask:     &types.BaseTask{State: types.AssetTransferInit},
		RebalanceId:  task.ID,
		TransferType: types.Invest,
		Params:       string(investParams),
	}

	return
}

func (t *transferInHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		if nextState == types.PartReBalanceInvest {
			var invest *types.AssetTransferTask

			invest, execErr = t.createInvestTask(task)
			if execErr != nil {
				return
			}

			execErr = t.db.SaveAssetTransferTask(session, invest)
			if execErr != nil {
				logrus.Errorf("save investHandler task error:%v task:[%v]", execErr, task)
				return
			}
		}

		task.State = nextState
		execErr = t.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}
