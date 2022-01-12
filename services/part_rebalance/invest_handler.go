package part_rebalance

import (
	"encoding/json"
	"fmt"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type investHandler struct {
	db       types.IDB
	eChecker EventChecker
}

func newInvestHandler(db types.IDB, eChecker EventChecker, conf *config.Config) *investHandler {
	return &investHandler{
		db:       db,
		eChecker: eChecker,
	}
}

func (i *investHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(i.db, task, types.Invest)
	if err != nil {
		return
	}
	switch state {
	case types.StateSuccess:
		finished = true
		nextState = types.PartReBalanceSuccess
	case types.StateFailed:
		finished = true
		nextState = types.PartReBalanceFailed
	case types.StateOngoing:
		finished = false
	default:
		logrus.Errorf("invest checkFinished unrecognized state %v", state)
	}
	//检查账本是否更新
	if finished && nextState == types.PartReBalanceSuccess {
		txTasks, err1 := i.db.GetTransactionTasksWithPartRebalanceId(task.ID, types.Invest)
		if err1 != nil {
			finished = false
			return
		}
		var params []*checkEventParam
		for _, txTask := range txTasks {
			investParam := &types.InvestParam{}
			err1 := json.Unmarshal([]byte(txTask.Params), investParam)
			if err1 != nil {
				err = fmt.Errorf("investParamDecodeErr err:%v,data:%s", err, txTask.Params)
				return
			}
			var sAddrs []string
			for _, addr := range investParam.StrategyAddresses {
				sAddrs = append(sAddrs, addr.Hex())
			}
			param := &checkEventParam{
				ChainID:       txTask.ChainId,
				Hash:          txTask.Hash,
				StrategyAddrs: sAddrs,
			}
			params = append(params, param)
		}
		ok, err1 := checkEventsHandled(i.eChecker, params)
		if err1 != nil || !ok {
			logrus.Warnf("event not handled params:%+v,err:%v", params, err1)
			finished = false
			return
		}
		utils.GetPartReCost(task.ID).AppendReport("invest")
	}
	return
}

func (i *investHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
		task.State = nextState
		execErr = i.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})
	return
}

func (i *investHandler) GetOpenedTaskMsg(taskId uint64) string {
	return fmt.Sprintf(`
	# invest
	- taskID: %d
	`, taskId)
}
