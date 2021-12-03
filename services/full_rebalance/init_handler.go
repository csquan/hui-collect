package full_rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type initHandler struct {
	db types.IDB
}

func (i *initHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.FullReBalanceImpermanenceLoss, nil
}

func (i *initHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	task.State = nextState
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}