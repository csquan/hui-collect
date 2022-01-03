package part_rebalance

import (
	"encoding/json"
	"net/http"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type investHandler struct {
	db       types.IDB
	eChecker eventChecker
}

func (i *investHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(i.db, task, types.Invest)
	if err != nil {
		return
	}
	switch state {
	case types.StateSuccess:
		finished = true
		nextState = types.PartReBalanceSuccess
	case types.StateFailed:
		finished = true
		nextState = types.PartReBalanceFailed
	case types.StateOngoing:
		finished = false
	default:
		logrus.Errorf("invest checkFinished unrecognized state %v", state)
	}
	return
}

func (i *investHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	var txTasks []*types.TransactionTask
	var params []*checkEventParam
	for _, txTask := range txTasks {
		param := &checkEventParam{
			chainID: txTask.ChainId,
			hash:    txTask.Hash,
		}
		params = append(params, param)
	}
	if ok, err1 := i.checkEventsHandled(params); ok && err1 == nil {
		err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
			task.State = nextState
			execErr = i.db.UpdatePartReBalanceTask(session, task)
			if execErr != nil {
				logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
				return
			}
			return
		})
	} else {
		b, _ := json.Marshal(params)
		logrus.Warnf("event not handled hashs:%s", b)
	}

	return
}

func (i *investHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

func (i *investHandler) checkEventsHandled(params []*checkEventParam) (bool, error) {
	for _, p := range params {
		ok, err := i.eChecker.checkEventHandled(p)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

type checkEventParam struct {
	hash    string
	chainID int
}

type eventChecker interface {
	checkEventHandled(*checkEventParam) (bool, error)
}

type eventCheckHandler struct {
	url string
	c   *http.Client
}
