package full_rebalance

import (
	"encoding/json"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type impermanenceLostCheckHandler struct {
	db   types.IDB
	conf *config.Config
}

func (i *impermanenceLostCheckHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.ReBalanceState, err error) {
	finished, err = checkMarginJobStatus(i.conf.ApiConf.MarginUrl, string(task.ID))
	if err != nil {
		return
	}
	return finished, types.FullReBalanceClaimLP, nil
}

func (i *impermanenceLostCheckHandler) MoveToNextState(task *types.FullReBalanceTask, nextState types.ReBalanceState) (err error) {
	// TODO
	// 1.获取 LP
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

func checkMarginJobStatus(url string, bizNo string) (finished bool, err error) {
	req := struct {
		BizNo string `json:"bizNo"`
	}{bizNo}
	data, err := utils.DoPost(url+"status/query", req)
	if err != nil {
		logrus.Errorf("request ImpermanentLoss api err:%v", err)
		return
	}
	resp := &types.NormalResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v", err)
		return
	}
	if resp.Code != 200 {
		logrus.Errorf("callImpermanentLoss code not 200, msg:%s", resp.Msg)
		return
	}
	if v, ok := resp.Data["status"]; ok {
		return v.(string) == "SUCCESS", nil
	}
	return
}
