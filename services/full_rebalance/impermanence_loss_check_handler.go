package full_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type impermanenceLostCheckHandler struct {
	db types.IDB
}

func (i *impermanenceLostCheckHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	// TODO 查询平无常状态
	//http://phabricator.huobidev.com/w/financial-product-center/hermes/tech-doc/auto-transfer-api-v2/
	return true, types.FullReBalanceClaimLP, nil
}

func (i *impermanenceLostCheckHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	// TODO
	// 1.获取 LP
	// http://phabricator.huobidev.com/w/financial-product-center/hermes/prd/机枪池二期需求文档/第三阶段后端接口数据需求/
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
		execErr = i.db.UpdateFullReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part full_rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})
	return
}
