package part_rebalance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type transferInHandler struct {
	db       types.IDB
	eChecker EventChecker
}

func costSince(start int64) int64 {
	return time.Now().Unix() - start
}

func (t *transferInHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(t.db, task, types.ReceiveFromBridge)
	if err != nil {
		return
	}
	switch state {
	case types.StateSuccess: //tx suc check event handled

		txTasks, err1 := t.db.GetTransactionTasksWithPartRebalanceId(task.ID, types.ReceiveFromBridge)
		if err != nil {
			err = fmt.Errorf("get tx_task err:%v,task_id:%d,tx_type:%d", err1, task.ID, types.ReceiveFromBridge)
			return
		}
		if len(txTasks) != 0 {
			var params []*checkEventParam
			for _, txTask := range txTasks {
				param := &checkEventParam{
					ChainID: txTask.ChainId,
					Hash:    txTask.Hash,
				}
				params = append(params, param)
			}
			b, _ := json.Marshal(params)
			ok, err1 := checkEventsHandled(t.eChecker, params)
			if err1 != nil {
				err = fmt.Errorf("check receiveFromBridge err:%v,params:%s", err1, b)
				finished = false
				return
			}
			if ok {
				utils.GetPartReCost(task.ID).AppendReport("资金从跨链桥到vault")
				finished = true
				nextState = types.PartReBalanceInvest
				logrus.Infof("receiveFromBridgeEvent handled hashs:%s,task_id:%d", b, task.ID)
			} else {
				logrus.Warnf("receiveFromBridge event not handled hashs:%s,task_id:%d", b, task.ID)
			}
		} else {
			finished = true
			utils.GetPartReCost(task.ID).AppendReport("资金从跨链桥到vault")
			nextState = types.PartReBalanceInvest
		}
	case types.StateFailed:
		finished = true
		nextState = types.PartReBalanceFailed
	case types.StateOngoing:
		finished = false
	default:
		logrus.Errorf("transferIn checkFinished unrecognized state %v", state)
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
			if approveTask != nil {
				tasks = append(tasks, approveTask)
			}
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

func (t *transferInHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# receiveFromBridge
	- taskID: %d
	`, taskId)
}
