package rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type doPartRebalanceHandler struct {
	db types.IDB
}

func (d *doPartRebalanceHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.ReBalanceWithdrawLP, nil
}

func (d *doPartRebalanceHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	return
}