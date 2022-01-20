package full_rebalance

import (
	"fmt"

	"github.com/starslabhq/hermes-rebalance/types"
)

type paramsCalcHandler struct {
	db types.IDB
}

func (r *paramsCalcHandler) Name() string {
	return "paramsCalcHandler"
}

func (r *paramsCalcHandler) Do(task *types.FullReBalanceTask) (err error) {
	return moveState(r.db, task, types.FullReBalanceParamsCalc, nil)
}

func (r *paramsCalcHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	return
}

func (r *paramsCalcHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# paramsCalcHandler
	- taskID: %d
	`, taskId)
}
