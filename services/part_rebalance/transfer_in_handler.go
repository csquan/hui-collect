package part_rebalance

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type transferInHandler struct {
	db types.IDB
}

func (t *transferInHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	state, err := getTransactionState(t.db, task, types.ReceiveFromBridge)
	if err != nil {
		return
	}

	if state != types.StateSuccess && state != types.StateFailed {
		return
	}

	finished = true

	if state == types.StateSuccess {
		nextState = types.PartReBalanceInvest
	} else {
		nextState = types.PartReBalanceFailed
	}

	return
}

func (t *transferInHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceInvest {
		tasks, err = t.CreateInvestTask(task)
		if err != nil {
			logrus.Errorf("InvestTask error:%v task:[%v]", err, task)
			return
		}
		if tasks, err = SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
			logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
			return
		}
	}

	err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		if nextState == types.PartReBalanceInvest {
			if err = t.db.SaveTxTasks(session, tasks); err != nil {
				logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
				return
			}
		}

		task.State = nextState
		execErr = t.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}

func (t *transferInHandler) CreateInvestTask(task *types.PartReBalanceTask) (tasks []*types.TransactionTask, err error) {
	params, err := task.ReadTransactionParams(types.Invest)
	if err != nil {
		return
	}
	var data, inputData string
	for _, param := range params {
		data, err = param.EncodeParam()
		if err != nil {
			logrus.Errorf("CreateTransactionTask param marshal err:%v", err)
			return
		}
		inputData, err = param.EncodeInput()
		if err != nil {
			logrus.Errorf("InvestInput err:%v", err)
			return
		}
		chainId, chainName, from, to := param.GetBase()
		task := &types.TransactionTask{
			BaseTask:        &types.BaseTask{State: int(types.TxUnInitState)},
			RebalanceId:     task.ID,
			TransactionType: int(types.Invest),
			ChainId:         chainId,
			ChainName:       chainName,
			From:            from,
			To:              to,
			Params:          data,
			InputData:       inputData,
		}
		tasks = append(tasks, task)
	}
	return
}
