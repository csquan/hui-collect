package full_rebalance

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type claimLPHandler struct {
	db types.IDB
}

func (w *claimLPHandler) Name() string {
	return "full_rebalance_claim"
}

func (w *claimLPHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	//TODO 检查所有txTask状态
	return true, types.FullReBalanceMarginBalanceTransferOut, nil
}

func (w *claimLPHandler) Do(task *types.FullReBalanceTask) error {
	return nil
}
