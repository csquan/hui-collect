package part_rebalance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
		url: conf.ApiConf.TaskManager,
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
	//检查账本是否更新
	if finished && nextState == types.PartReBalanceSuccess {
		txTasks, err1 := i.db.GetTransactionTasksWithPartRebalanceId(task.ID, types.Invest)
		if err1 != nil {
			finished = false
			return
		}
		var params []*checkEventParam
		for _, txTask := range txTasks {
			param := &checkEventParam{
				ChainID: txTask.ChainId,
				Hash:    txTask.Hash,
			}
			params = append(params, param)
		}
		ok, err1 := i.checkEventsHandled(params)
		if err1 != nil || !ok {
			logrus.Warnf("event not handled params:%+v,err:%v", params, err1)
			finished = false
			return
		}
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
	return fmt.Sprintf(`
	# invest
	- taskID: %d
	`, taskId)
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

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
	Data bool   `json:"data"`
}

func (e *eventCheckHandler) checkEventHandled(p *checkEventParam) (result bool, err error) {
	path := fmt.Sprintf("/v1/open/hash?hash=%s&chainId=%d", p.Hash, p.ChainID)
	urlStr, err := utils.JoinUrl(e.url, path)
	if err != nil {
		logrus.Warnf("parse url error:%v", err)
		return
	}
	urlStr, err = url.QueryUnescape(urlStr)
	if err != nil {
		logrus.Warnf("checkEventHandled QueryUnescape error:%v", err)
		return
	}
	data, err := utils.DoRequest(urlStr, "GET", nil)
	if err != nil {
		return
	}
	resp := &response{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Warnf("unmarshar resp err:%v,body:%s", err, data)
		return
	}
	if resp.Code != 200 {
		logrus.Infof("checkEvent response %v", resp)
		return false, errors.New("response code not 200")
	}
	return resp.Data, nil
}
