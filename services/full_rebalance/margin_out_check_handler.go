package full_rebalance

import "github.com/starslabhq/hermes-rebalance/types"

type marginOutCheckHandler struct {
	db types.IDB
}

func (i *marginOutCheckHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.FullReBalanceRecycling, nil
}

func (i *marginOutCheckHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	task.State = nextState
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}