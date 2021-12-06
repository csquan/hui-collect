package full_rebalance

import (
	"container/list"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type StateHandler interface {
	CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error)
	MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error)
	GetState() types.ReBalanceState
	Do(task *types.FullReBalanceTask) error
}

type ReBalance struct {
	db     types.IDB
	config *config.Config
	// handlers map[types.ReBalanceState]StateHandler
	handlers *list.List
}

func NewReBalanceService(db types.IDB, conf *config.Config) (p *ReBalance, err error) {
	p = &ReBalance{
		db:       db,
		config:   conf,
		handlers: list.New(),
		// handlers: map[types.ReBalanceState]StateHandler{
		// 	types.FullReBalanceInit: &initHandler{
		// 		db: db,
		// 	},
		// 	types.FullReBalanceImpermanenceLoss: &impermanenceLostHandler{
		// 		db: db,
		// 	},
		// 	types.FullReBalanceClaimLP: &claimLPHandler{
		// 		db: db,
		// 	},
		// 	types.FullReBalanceMarginBalanceTransferOut: &marginOutHandler{
		// 		db: db,
		// 	},
		// 	types.FullReBalanceRecycling: &recyclingHandler{
		// 		db: db,
		// 	},
		// 	// 计算状态由python处理
		// 	//types.FullReBalanceParamsCalc: &paramsCalcHandler{
		// 	//	db: db,
		// 	//},
		// 	types.FullReBalanceOndoing: &doPartRebalanceHandler{
		// 		db: db,
		// 	},
		// },
	}
	handlers := []StateHandler{
		&initHandler{
			db: db,
		},
		// &impermanenceLostHandler{
		// 	db: db,
		// },
		// &claimLPHandler{
		// 	db: db,
		// },
		// &marginOutHandler{
		// 	db: db,
		// },
		// &recyclingHandler{
		// 	db: db,
		// },
		// &doPartRebalanceHandler{
		// 	db: db,
		// },
	}
	for _, h := range handlers {
		p.handlers.PushBack(h)
	}

	return
}

func (p *ReBalance) getHandler(state types.ReBalanceState) (cur, next StateHandler) {
	for e := p.handlers.Front(); e != nil; e = e.Next() {
		v := e.Value
		h, _ := v.(StateHandler)
		if h != nil && h.GetState() == state {
			nextElemetn := e.Next()
			if nextElemetn != nil {
				next = nextElemetn.Value.(StateHandler)
			}
			return h, next
		}
	}
	return nil, nil
}

func (p *ReBalance) Name() string {
	return "full_rebalance"
}

func (p *ReBalance) Run() (err error) {
	tasks, err := p.db.GetOpenedFullReBalanceTasks()
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		logrus.Infof("no available part full_rebalance task.")
		return
	}

	if len(tasks) > 1 {
		err = fmt.Errorf("more than one full_rebalance tasks are being processed. tasks:%v", tasks)
		return
	}

	// handler, ok := p.handlers[tasks[0].State]
	// if !ok {
	// 	err = fmt.Errorf("unkonwn state for part full_rebalance task:%v", tasks[0])
	// 	return
	// }
	handler, nextHandler := p.getHandler(tasks[0].State)
	finished, next, err := handler.CheckFinished(tasks[0])
	if err != nil {
		return err
	}

	if !finished {
		return
	}
	if err := nextHandler.Do(tasks[0]); err != nil {
		err = handler.MoveToNextState(tasks[0], next)
		if err != nil {
			return err
		}
	}
	return
}
