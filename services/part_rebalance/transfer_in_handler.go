package part_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type transferInHandler struct {
	db types.IDB
}

func (t *transferInHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(t.db, task, types.ReceiveFromBridge)
	if err != nil {
		return
	}

	if state != types.StateSuccess && state != types.StateFailed {
		return
	}

	finished = true

	if state == types.StateSuccess {
		nextState = types.PartReBalanceInvest
	} else {
		nextState = types.PartReBalanceFailed
	}

	return
}

func (t *transferInHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceInvest {
		tasks, err = CreateTransactionTask(task, types.Invest)
		if err != nil {
			logrus.Errorf("InvestTask error:%v task:[%v]", err, task)
			return
		}
	}

	err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		if nextState == types.PartReBalanceInvest {
			if err = t.db.SaveTxTasks(session, tasks); err != nil {
				logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
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

func CreateTransactionTask(task *types.PartReBalanceTask, transactionType types.TransactionType) (tasks []*types.TransactionTask, err error) {
	params, err := task.ReadTransactionParams(transactionType)
	if err != nil {
		return
	}
	for _, param := range params {
		if transactionType == types.ReceiveFromBridge {
			var approveTask *types.TransactionTask
			approveTask, err = CreateApproveTask(task.ID, param.(*types.ReceiveFromBridgeParam))
			if err != nil {
				logrus.Errorf("create approve task from param err:%v task:%v", err, task)
				return
			}
			tasks = append(tasks, approveTask)
		}
		t, err := param.CreateTask(task.ID)
		if err != nil {
			logrus.Errorf("create task from param err:%v task:%v", err, task)
		}
		tasks = append(tasks, t)
	}
	if tasks, err = SetNonceAndGasPrice(tasks); err != nil {
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	return
}
