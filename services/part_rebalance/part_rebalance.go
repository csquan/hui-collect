package part_rebalance

import (
	"fmt"
	"github.com/starslabhq/hermes-rebalance/clients"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type StateHandler interface {
	CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error)
	MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error)
}

type PartReBalance struct {
	db       types.IDB
	config   *config.Config
	handlers map[types.PartReBalanceState]StateHandler
}

func NewPartReBalanceService(db types.IDB, conf *config.Config) (p *PartReBalance, err error) {
	p = &PartReBalance{
		db:     db,
		config: conf,
		handlers: map[types.PartReBalanceState]StateHandler{
			types.PartReBalanceInit: &initHandler{
				db: db,
			},
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
			types.PartReBalanceInvest: &investHandler{
				db: db,
			},
		},
	}

	return
}

func (p *PartReBalance) Name() string {
	return "part_rebalance"
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

	handler, ok := p.handlers[tasks[0].State]
	if !ok {
		err = fmt.Errorf("unkonwn state for part rebalance task:%v", tasks[0])
		return
	}

	finished, next, err := handler.CheckFinished(tasks[0])
	if err != nil {
		return err
	}

	if !finished {
		return
	}

	logrus.Infof("part rebalance task move state, from:[%v], to:[%v]", tasks[0].State, next)
	err = handler.MoveToNextState(tasks[0], next)
	if err != nil {
		return err
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
