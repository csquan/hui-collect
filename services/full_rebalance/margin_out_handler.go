package full_rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type marginOutHandler struct {
	db types.IDB
}

func (i *marginOutHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.FullReBalanceRecycling, nil
}

func (i *marginOutHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	task.State = nextState
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}
