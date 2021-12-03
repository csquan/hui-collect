package rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type paramsCalcHandler struct {
	db types.IDB
}

func (i *paramsCalcHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.ReBalanceDoPartRebalance, nil
}

func (i *paramsCalcHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	return
}
