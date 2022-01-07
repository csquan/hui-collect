package part_rebalance

import (
	"errors"
	"fmt"
	"time"

	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/clients"
	"github.com/starslabhq/hermes-rebalance/utils"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type StateHandler interface {
	CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error)
	MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error)
	GetOpenedTaskMsg(taskId uint64) string
}

type PartReBalance struct {
	db       types.IDB
	config   *config.Config
	handlers map[types.PartReBalanceState]StateHandler
	ticker   int64
}

func NewPartReBalanceService(db types.IDB, conf *config.Config) (p *PartReBalance, err error) {
	p = &PartReBalance{
		db:     db,
		config: conf,
		handlers: map[types.PartReBalanceState]StateHandler{
			types.PartReBalanceInit: newInitHandler(db, conf),
			types.PartReBalanceTransferOut: &transferOutHandler{
				db: db,
			},
			types.PartReBalanceCross: &crossHandler{
				db:        db,
				clientMap: clients.ClientMap,
			},
			types.PartReBalanceTransferIn: &transferInHandler{
				db: db,
			},
			types.PartReBalanceInvest: newInvestHandler(db, conf),
		},
	}

	return
}

func (p *PartReBalance) Name() string {
	return "part_rebalance"
}

func (p *PartReBalance) clearTick() {
	p.ticker = 0
}

func (p *PartReBalance) startTick() {
	p.ticker = time.Now().Unix()
}

func (p *PartReBalance) Run() (err error) {
	tasks, err := p.db.GetOpenedPartReBalanceTasks()
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		logrus.Infof("no available part rebalance task.")
		return
	}

	if len(tasks) > 1 {
		err = fmt.Errorf("more than one rebalance tasks are being processed. tasks:%v", tasks)
		return
	}
	if p.ticker == 0 {
		p.startTick()
	}

	handler, ok := p.handlers[tasks[0].State]
	if !ok {
		err = fmt.Errorf("unkonwn state for part rebalance task:%v", tasks[0])
		return
	}

	finished, next, err := handler.CheckFinished(tasks[0])
	if !finished {
		now := time.Now().Unix()
		if now-p.ticker > p.config.Alert.MaxWaitTime {
			// 把子状态拿出来
			msg := handler.GetOpenedTaskMsg(tasks[0].ID)
			if msg != "" {
				alert.Dingding.SendAlert("State 停滞提醒", msg, nil)
			}
			p.clearTick()
		}
		return
	}
	if err != nil {
		return err
	}
	if finished {
		p.clearTick()
	}
	if next == types.PartReBalanceFailed || next == types.PartReBalanceSuccess {
		if tasks[0].FullRebalanceID == 0 {
			var resp *types.TaskManagerResponse
			resp, err = utils.CallTaskManager(p.config, fmt.Sprintf("/v1/open/task/end/%s?taskType=rebalance", tasks[0].TaskID), "POST")
			if err != nil || !resp.Data {
				logrus.Infof("call task manager end resp:%v, err:%v, taskID:%s", resp, err, tasks[0].TaskID)
				return
			}
		}
	}
	var status string
	tasks[0].Message, status = utils.GenPartRebalanceMessage(next, "")
	logrus.Infof("part rebalance task move state, from:[%v], to:[%v]", types.PartReBalanceStateName[tasks[0].State], types.PartReBalanceStateName[next])
	err = handler.MoveToNextState(tasks[0], next)
	if err != nil {
		message, _ := utils.GenPartRebalanceMessage(next, fmt.Sprintf("%v", err))
		p.db.UpdatePartReBalanceTaskMessage(tasks[0].ID, message)
		return err
	}
	if next == types.PartReBalanceFailed {
		alert.Dingding.SendAlert("Part Rebalance State Change", alert.TaskFailedContent("小Re", tasks[0].ID, status, errors.New(tasks[0].Message)), nil)
	} else {
		alert.Dingding.SendMessage("Part Rebalance State Change", alert.TaskStateChangeContent("小Re", tasks[0].ID, status))
	}
	return
}

//getTransactionState
func getTransactionState(db types.IDB, task *types.PartReBalanceTask, transferType types.TransactionType) (state types.TaskState, err error) {
	txTasks, err := db.GetTransactionTasksWithReBalanceId(task.ID, transferType)
	if err != nil {
		logrus.Errorf("get transaction task error:%v", err)
		return
	}
	if len(txTasks) == 0 {
		state = types.StateSuccess
		return
	}
	success := true
	for _, tx := range txTasks {
		if tx.State == int(types.TxFailedState) {
			state = types.StateFailed
			return
		}
		success = success && (tx.State == int(types.TxSuccessState))
	}
	if success {
		state = types.StateSuccess
	} else {
		state = types.StateOngoing
	}
	return
}
