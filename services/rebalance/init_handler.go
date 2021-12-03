package rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type initHandler struct {
	db types.IDB
}

func (i *initHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	//检查是否有进行中的小R，如果有则停留在当前状态，等待小r结束
	//TODO 目前仅根据part_rebalance判断不够准确。part_rebalance如果是失败状态，子任务仍然有可能在运行。
	finished = false
	tasks, err := i.db.GetOpenedFullReBalanceTasks()
	if err != nil {
		logrus.Errorf("GetOpenedPartReBalanceTasks err:%v", err)
		return
	}
	if len(tasks) > 0 {
		logrus.Infof("have part rebalance task on running.")
		return
	}
	return true, types.ReBalanceWithdrawLP, nil
}

func (i *initHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	if nextState != types.ReBalanceWithdrawLP {
		return nil
	}
	//TODO 请求依赖服务，获取拆解LP所需参数，创建拆解LP的transactionTask的任务列表
	var params []*types.ClaimFromVaultParam
	var tasks []*types.TransactionTask
	for _, p := range params {
		t, err := p.CreateTask(task.ID)
		if err != nil {
			logrus.Errorf("create ClaimFromVault task from param err:%v task:%v", err, task)
		}
		tasks = append(tasks, t)
	}
	if tasks, err = part_rebalance.SetNonceAndGasPrice(tasks); err != nil {
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
		if err = i.db.SaveTxTasks(session, tasks); err != nil {
			logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
			return
		}
		//move to next state
		task.State = nextState
		execErr = i.db.UpdateReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})
	return
}
