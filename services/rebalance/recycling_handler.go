package rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type recyclingHandler struct {
	db types.IDB
}

func (r *recyclingHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.ReBalanceParamsCalc, nil
}

func (r *recyclingHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	return
}
