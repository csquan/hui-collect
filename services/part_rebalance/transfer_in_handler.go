package part_rebalance

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	params, err := task.ReadParams()
	if err != nil {
		return
	}
	var data, inputData []byte
	for _, param := range params.InvestParams {
		data, err = json.Marshal(param)
		if err != nil {
			logrus.Errorf("CreateTransactionTask param marshal err:%v", err)
			return
		}
		inputData, err = utils.InvestInput(param)
		if err != nil {
			logrus.Errorf("InvestInput err:%v", err)
			return
		}
		task := &types.TransactionTask{
			BaseTask:        &types.BaseTask{State: int(types.TxUnInitState)},
			RebalanceId:     task.ID,
			TransactionType: int(types.Invest),
			ChainId:         param.ChainId,
			ChainName:       param.ChainName,
			From:            param.From,
			To:              param.To,
			Params:          string(data),
			InputData:       hexutil.Encode(inputData),
		}
		tasks = append(tasks, task)
	}
	return
}
