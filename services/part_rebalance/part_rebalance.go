package part_rebalance

import (
	"fmt"

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
			types.PartReBalanceCross: &crossHandler{
				db: db,
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

	err = handler.MoveToNextState(tasks[0], next)
	if err != nil {
		return err
	}

	return
}

//getTransferSatae
func getTransferState(db types.IDB, task *types.PartReBalanceTask, transferType int) (state int, err error) {
	atTasks, err := db.GetAssetTransferTasksWithReBalanceId(task.ID, transferType)
	if err != nil {
		logrus.Errorf("get asset transfer task error:%v", err)
		return
	}

	if len(atTasks) == 0 {
		err = fmt.Errorf("part rebalance task [%v] has no transfer in task", task)
		return
	}
	success := true
	for _, at := range atTasks {
		if at.State == types.AssetTransferFailed {
			state = types.AssetTransferFailed
			return
		}
		success = success && (at.State == types.AssetTransferSuccess)
	}
	if success {
		state = types.AssetTransferSuccess
	} else {
		state = types.AssetTransferOngoing
	}
	return
}
