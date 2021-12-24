package part_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type transferOutHandler struct {
	db types.IDB
}

func (t *transferOutHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(t.db, task, types.SendToBridge)
	if err != nil {
		return
	}
	switch state {
	case types.StateSuccess:
		finished = true
		nextState = types.PartReBalanceCross
	case types.StateFailed:
		finished = true
		nextState = types.PartReBalanceFailed
	case types.StateOngoing:
		finished = false
	default:
		logrus.Errorf("transferOut checkFinished unrecognized state %v", state)
	}
	return
}

func (t *transferOutHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}
	crossBalances := params.CrossBalances

	if len(crossBalances) == 0 {
		//logrus.Errorf("no cross balance is found for rebalance task: [%v]", task)
		task.State = nextState
		err = t.db.UpdatePartReBalanceTask(t.db.GetEngine(), task)
		if err != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", err, task)
			return
		}
		return
	}

	crossTasks := make([]*types.CrossTask, 0, len(crossBalances))
	for _, param := range crossBalances {
		crossTasks = append(crossTasks, &types.CrossTask{
			//BaseTask:      &types.BaseTask{State: types.ToCreateSubTask},
			State:         types.ToCreateSubTask,
			RebalanceId:   task.ID,
			ChainFrom:     param.FromChain,
			ChainTo:       param.ToChain,
			ChainFromAddr: param.FromAddr,
			ChainToAddr:   param.ToAddr,
			CurrencyFrom:  param.FromCurrency,
			CurrencyTo:    param.ToCurrency,
			Amount:        param.Amount,
		})
	}

	err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		if nextState == types.PartReBalanceCross {
			execErr = t.db.SaveCrossTasks(session, crossTasks)
			if execErr != nil {
				logrus.Errorf("save cross task error:%v task:[%v]", execErr, task)
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

func (t *transferOutHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}
