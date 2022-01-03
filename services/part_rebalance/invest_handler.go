package part_rebalance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type investHandler struct {
	db       types.IDB
	eChecker EventChecker
}

func newInvestHandler(db types.IDB, conf *config.Config) *investHandler {
	eChecker := &eventCheckHandler{
		url: "", //TODO
		c: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	return &investHandler{
		db:       db,
		eChecker: eChecker,
	}
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
	txTasks, err1 := i.db.GetTransactionTasksWithPartRebalanceId(task.ID, types.Invest)
	if err1 != nil {
		return fmt.Errorf("get part_rebalance tasks err:%v", err1)
	}
	var params []*checkEventParam
	for _, txTask := range txTasks {
		param := &checkEventParam{
			ChainID: txTask.ChainId,
			Hash:    txTask.Hash,
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
		logrus.Warnf("event not handled hashs:%s,err:%v", b, err1)
	}

	return
}

func (i *investHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

func (i *investHandler) checkEventsHandled(params []*checkEventParam) (bool, error) {
	if len(params) == 0 {
		return true, nil
	}
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
	Hash    string
	ChainID int
}

//go:generate mockgen -source=$GOFILE -destination=./mock_invest_handler.go -package=part_rebalance
type EventChecker interface {
	checkEventHandled(*checkEventParam) (bool, error)
}

type eventCheckHandler struct {
	url string
	c   *http.Client
}

func (e *eventCheckHandler) checkEventHandled(*checkEventParam) (bool, error) {
	return true, nil
}
