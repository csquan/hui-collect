package full_rebalance
import (
	"github.com/starslabhq/hermes-rebalance/types"
)

type claimLPCheckHandler struct {
	db types.IDB
}

func (i *claimLPCheckHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	return true, types.FullReBalanceMarginBalanceTransferOut, nil
}

func (i *claimLPCheckHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	task.State = nextState
	err = i.db.UpdateFullReBalanceTask(i.db.GetSession(), task)
	return
}