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
	return true, types.PartReBalanceTransferOut, nil
}

func (i *initHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceTransferOut {
		tasks, err = CreateTransactionTask(task, types.SendToBridge)
		if err != nil {
			logrus.Errorf("InvestTask error:%v task:[%v]", err, task)
			return
		}
	}
	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
		//create next state task
		if nextState == types.PartReBalanceTransferOut {
			if err = i.db.SaveTxTasks(session, tasks); err != nil {
				logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
				return
			}
		}
		//move to next state
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
