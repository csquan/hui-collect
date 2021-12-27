package part_rebalance

import (
	"fmt"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type initHandler struct {
	db        types.IDB
	nextState types.PartReBalanceState
}

func newInitHandler(db types.IDB) *initHandler {
	return &initHandler{
		db:        db,
		nextState: types.PartReBalanceTransferOut, //default
	}
}

func (i *initHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	return true, i.nextState, nil
}

func (i *initHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceTransferOut {
		params, err1 := task.ReadTransactionParams(types.SendToBridge)
		if err1 != nil {
			return fmt.Errorf("sendToBridge params err:%v", err1)
		}
		for _, param := range params {
			sendParam, ok := param.(*types.SendToBridgeParam)
			if !ok {
				logrus.Fatalf("unexpected sendtobridge param:%v", param)
			}
			err = i.checkSendToBridgeParam(sendParam)
			if err != nil {
				i.nextState = types.PartReBalanceFailed //check err change part_rebalance to fail
				return err
			}
		}
		tasks, err = CreateTransactionTask(task, types.SendToBridge, params)
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

func (i *initHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

func (i *initHandler) checkSendToBridgeParam(param *types.SendToBridgeParam) error {
	return nil
}
