package full_rebalance

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

var states []types.ReBalanceState = []types.ReBalanceState{
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
	CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error)
	//movetoCurstate
	Do(task *types.FullReBalanceTask) error
	Name() string
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
			types.FullReBalanceInit: &initHandler{
				db: db,
			},
			// types.FullReBalanceMarginIn: &impermanenceLostHandler{
			// 	db: db,
			// },
			// types.FullReBalanceClaimLP: &claimLPHandler{
			// 	db: db,
			// },
			// types.FullReBalanceMarginBalanceTransferOut: &marginOutHandler{
			// 	db: db,
			// },
			// types.FullReBalanceRecycling: &recyclingHandler{
			// 	db: db,
			// },
			// // 计算状态由python处理
			// //types.FullReBalanceParamsCalc: &paramsCalcHandler{
			// //	db: db,
			// //},
			// types.FullReBalanceOndoing: &doPartRebalanceHandler{
			// 	db: db,
			// },
		},
	}

	return
}

func (p *ReBalance) getHandler(state types.ReBalanceState) StateHandler {
	return p.handlers[state]
}

func (p *ReBalance) Name() string {
	return "full_rebalance"
}

func checkState(state types.ReBalanceState) error {
	for _, v := range states {
		if v == state {
			return nil
		}
	}
	return fmt.Errorf("state:%d err", state)
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

	handler := p.getHandler(tasks[0].State)
	finished, next, err := handler.CheckFinished(tasks[0])
	if err != nil {
		return err
	}

	if !finished {
		return
	}
	if err := checkState(next); err != nil {
		return err
	}
	if next == types.FullReBalanceSuccess || next == types.FullReBalanceFailed {
		//update state
		tasks[0].State = next
		return p.db.UpdateFullReBalanceTask(p.db.GetEngine(), tasks[0])
	} else {
		nextHandler := p.getHandler(next)
		if nextHandler == nil {
			b, _ := json.Marshal(tasks[0])
			logrus.Fatalf("unexpectd state:%d,task:%s", next, b)
		}
		if err := nextHandler.Do(tasks[0]); err != nil {
			logrus.Errorf("handler do err:%v,name:%s", err, nextHandler.Name())
			return err
		}
	}
	return
}
