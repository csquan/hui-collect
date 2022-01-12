package full_rebalance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/tokens"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

var states []types.FullReBalanceState = []types.FullReBalanceState{
	types.FullReBalanceInit,
	types.FullReBalanceMarginIn,
	types.FullReBalanceClaimLP,
	types.FullReBalanceMarginBalanceTransferOut,
	types.FullReBalanceRecycling,
	types.FullReBalanceParamsCalc,
	types.FullReBalanceOndoing,
	types.FullReBalanceSuccess,
	types.FullReBalanceFailed,
}

type StateHandler interface {
	CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error)
	//movetoCurstate
	Do(task *types.FullReBalanceTask) error
	Name() string
	GetOpenedTaskMsg(taskId uint64) string
}

type FullReBalance struct {
	db                 types.IDB
	config             *config.Config
	handlers           map[types.FullReBalanceState]StateHandler
	ticker             int64
	checkInfo          *checkInfo
}
type checkInfo struct {
	taskID   uint64
	curState int
	next     int
	finished bool
}

func NewReBalanceService(db types.IDB, conf *config.Config) (p *FullReBalance, err error) {
	token, err := tokens.NewTokens(db)
	if err != nil {
		return nil, fmt.Errorf("create tokens err:%v", err)
	}
	p = &FullReBalance{
		db:     db,
		config: conf,
		handlers: map[types.FullReBalanceState]StateHandler{
			types.FullReBalanceInit: &initHandler{
				db:   db,
				conf: conf,
			},
			types.FullReBalanceMarginIn: &impermanenceLostHandler{
				conf: conf,
				db:   db,
			},
			types.FullReBalanceClaimLP: newClaimLpHandler(conf, db, token),
			types.FullReBalanceMarginBalanceTransferOut: &marginOutHandler{
				conf: conf,
				db:   db,
			},
			types.FullReBalanceRecycling: &recyclingHandler{
				conf: conf,
				db:   db,
			},
			//计算状态由python处理
			types.FullReBalanceParamsCalc: &paramsCalcHandler{
				db: db,
			},
			types.FullReBalanceOndoing: &doPartRebalanceHandler{
				db: db,
			},
		},
	}

	return
}

func (p *FullReBalance) getHandler(state types.FullReBalanceState) StateHandler {
	return p.handlers[state]
}

func (p *FullReBalance) Name() string {
	return "full_rebalance"
}

func checkState(state types.FullReBalanceState) error {
	for _, v := range states {
		if v == state {
			return nil
		}
	}
	return fmt.Errorf("unexpected state")
}

func (p *FullReBalance) startTick() {
	p.ticker = time.Now().Unix()
}
func (p *FullReBalance) clearTick() {
	p.ticker = 0
}

func (p *FullReBalance) Run() (err error) {
	tasks, err := p.db.GetOpenedFullReBalanceTasks()
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		logrus.Infof("no available full_rebalance task.")
		return
	}

	if len(tasks) > 1 {
		err = fmt.Errorf("more than one full_rebalance tasks are being processed. tasks:%v", tasks)
		return
	}
	if p.ticker == 0 {
		p.startTick()
	}
	if p.checkInfo == nil {
		p.checkInfo = &checkInfo{}
	}
	//checkInfo的作用是避免重复checkFinished。
	handler := p.getHandler(tasks[0].State)
	if p.checkInfo.curState == tasks[0].State && p.checkInfo.taskID == tasks[0].ID && p.checkInfo.finished {
		logrus.Infof("task step has finished,taskID:%d state:%d", tasks[0].ID, tasks[0].State)
	} else {
		p.checkInfo.finished, p.checkInfo.next, err = handler.CheckFinished(tasks[0])
		if err != nil {
			return err
		}
		if !p.checkInfo.finished {
			now := time.Now().Unix()
			if now-p.ticker > p.config.Alert.MaxWaitTime {
				// 把子状态拿出来
				var msg = handler.GetOpenedTaskMsg(tasks[0].ID)
				if msg != "" {
					alert.Dingding.SendAlert("State 停滞提醒", msg, nil)
				}
				p.clearTick()
			}
			return
		} else {
			p.checkInfo.curState = tasks[0].State
			p.checkInfo.taskID = tasks[0].ID
			p.clearTick()
		}
	}
	if err := checkState(p.checkInfo.next); err != nil {
		return fmt.Errorf("state err:%v,state:%d,tid:%d,handler:%s", err, p.checkInfo.next, tasks[0].ID, handler.Name())
	}
	status := types.FullReBalanceStateName[p.checkInfo.next]
	if p.checkInfo.next == types.FullReBalanceSuccess || p.checkInfo.next == types.FullReBalanceFailed {
		var resp *types.TaskManagerResponse
		resp, err = utils.CallTaskManager(p.config, fmt.Sprintf(`/v1/open/task/end/Full_%d?taskType=rebalance`, tasks[0].ID), "POST")
		if err != nil || !resp.Data {
			logrus.Infof("call task manager func:end resp:%v, err:%v", resp, err)
			return
		}
		logrus.Info(utils.GetFullReCost(tasks[0].ID).Report)
		alert.Dingding.SendMessage("Full Rebalance State Change", alert.TaskStateChangeContent("大Re", tasks[0].ID, status))
		return moveState(p.db, tasks[0], p.checkInfo.next, nil)
	} else {
		nextHandler := p.getHandler(p.checkInfo.next)
		if nextHandler == nil {
			b, _ := json.Marshal(tasks[0])
			logrus.Fatalf("unexpectd state:%d,task:%s", p.checkInfo.next, b)
			return
		}
		if err := nextHandler.Do(tasks[0]); err != nil {
			logrus.Errorf("handler do err:%v,name:%s", err, nextHandler.Name())
			return err
		}
		alert.Dingding.SendMessage("Full Rebalance State Change", alert.TaskStateChangeContent("大Re", tasks[0].ID, status))
	}
	return
}

func moveState(db types.IDB, task *types.FullReBalanceTask, state types.FullReBalanceState, params interface{}) error {
	status := types.FullReBalanceStateName[state]
	task.AppendMessage(&types.FullReMsg{Status: status, Params: params})
	task.State = state
	return db.UpdateFullReBalanceTask(db.GetEngine(), task)
}
