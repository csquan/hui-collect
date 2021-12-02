package rebalance

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type StateHandler interface {
	CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error)
	MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error)
}

type ReBalance struct {
	db       types.IDB
	config   *config.Config
	handlers map[types.ReBalanceState]StateHandler
}

func NewReBalanceService(db types.IDB, conf *config.Config) (p *ReBalance, err error) {
	p = &ReBalance{
		db:     db,
		config: conf,
		handlers: map[types.ReBalanceState]StateHandler{
			types.ReBalanceInit: &initHandler{
				db: db,
			},
			types.ReBalanceWithdrawLP: &withdrawLPHandler{
				db: db,
			},
			types.ReBalanceRecycling: &recyclingHandler{
				db: db,
			},
			types.ReBalanceParamsCalc: &paramsCalcHandler{
				db: db,
			},
			types.ReBalanceDoPartRebalance: &doPartRebalanceHandler{
				db: db,
			},
		},
	}

	return
}

func (p *ReBalance) Name() string {
	return "rebalance"
}

func (p *ReBalance) Run() (err error) {
	tasks, err := p.db.GetOpenedFullReBalanceTasks()
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
