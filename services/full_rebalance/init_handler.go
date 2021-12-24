package full_rebalance

import (
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type initHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *initHandler) Name() string {
	return "full_rebalance_init"
}

func (i *initHandler) Do(task *types.FullReBalanceTask) error {
	return nil
}

func (i *initHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	return true, types.FullReBalanceMarginIn, nil
}

func (i *initHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}
