package part_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type initHandler struct {
	db types.IDB
}

func (i *initHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	return true, types.PartReBalanceCross, nil
}

func (i *initHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	params, err := task.ReadParams()
	if err != nil {
		return
	}
	crossBalances := params.CrossBalances

	if len(crossBalances) == 0 {
		logrus.Errorf("no cross balance is found for rebalance task: [%v]", task)
		return
	}

	crossTasks := make([]*types.CrossTask, 0, len(crossBalances))
	for _, param := range crossBalances {
		crossTasks = append(crossTasks, &types.CrossTask{
			//BaseTask:      &types.BaseTask{State: types.ToCreateSubTask},
			State: types.ToCreateSubTask,
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

	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
		execErr = i.db.SaveCrossTasks(session, crossTasks)
		if execErr != nil {
			logrus.Errorf("save cross task error:%v task:[%v]", execErr, task)
			return
		}

		task.State = nextState
		execErr = i.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}

		return
	})

	return
}
