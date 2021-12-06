package full_rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type claimLPHandler struct {
	db types.IDB
}

func (w *claimLPHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	//TODO 检查所有txTask状态
	return true, types.FullReBalanceMarginBalanceTransferOut, nil
}

func (w *claimLPHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {

	//TODO 创建check任务

	return
}
