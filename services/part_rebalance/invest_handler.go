package part_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"net/http"
)

type investHandler struct {
	db types.IDB
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
	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
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

func (i *investHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

type checkEventParam struct {
	hash string
	chainID int
}

type eventChecker interface {
	checkEventHandled([]*checkEventParam) (bool, error)
}

type eventCheckHandler struct {
	url string
	c *http.Client
}

func(e *eventCheckHandler) checkEventHandled([]*checkEventParam) (bool, error){
	return true, nil
}