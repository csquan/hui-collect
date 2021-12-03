package rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type withdrawLPHandler struct {
	db types.IDB
}

func (w *withdrawLPHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {

	return true, types.ReBalanceRecycling, nil
}

func (w *withdrawLPHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	//TODO 计算param，创建ReBalanceRecycling（partRebalance，不包含invest参数）
	return
}