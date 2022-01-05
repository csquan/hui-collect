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
	task.State = types.FullReBalanceParamsCalc
	err = r.db.UpdateFullReBalanceTask(r.db.GetEngine(), task)
	return
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