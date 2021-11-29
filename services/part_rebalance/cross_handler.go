package part_rebalance

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type crossHandler struct {
	db types.IDB
}

func (c *crossHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	crossTasks, err := c.db.GetCrossTasksByReBalanceId(task.ID)
	if err != nil {
		logrus.Errorf("get cross task for rebalance [%v] failed, err:%v", task, err)
		return
	}

	if len(crossTasks) == 0 {
		err = fmt.Errorf("part rebalance task [%v] has no cross task", task)
		return
	}

	success := true
	for _, crossTask := range crossTasks {
		if crossTask.State != types.TaskSuc {
			logrus.Debugf("cross task [%v] is not finished", crossTask)

			return
		}

		success = success && crossTask.State == types.TaskSuc
	}

	if success {
		nextState = types.PartReBalanceTransferIn
	} else {
		nextState = types.PartReBalanceFailed
	}

	return true, nextState, nil
}

func (c *crossHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {

	err = utils.CommitWithSession(c.db, func(session *xorm.Session) (execErr error) {
		//create next state task
		if nextState == types.PartReBalanceTransferIn {
			execErr = CreateReceiveFromBridgeTask(task, c.db, session)
			if execErr != nil {
				logrus.Errorf("create transaction task error:%v task:[%v]", execErr, task)
				return
			}
		}

		//move to next state
		task.State = nextState
		execErr = c.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})
	return
}

func CreateReceiveFromBridgeTask(task *types.PartReBalanceTask, db types.IDB, session xorm.Interface) (err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}
	var tasks []*types.TransactionTask
	var paramData, inputData []byte
	for _, param := range params.ReceiveFromBridgeParams {
		var approve *types.ApproveRecord
		if approve, err = db.GetApprove(common.Address.String(param.Erc20ContractAddr), param.To); err != nil {
			logrus.Errorf("GetApprove err:%v", err)
			return
		}
		if approve == nil {
			var task *types.TransactionTask
			if task, err = CreateApproveTask(task.ID, param); err != nil{
				logrus.Errorf("CreateApproveTask err:%v", err)
				return
			}
			tasks = append(tasks, task)
		}
		paramData, err = json.Marshal(param)
		if err != nil {
			logrus.Errorf("CreateTransactionTask param marshal err:%v", err)
			return
		}
		inputData, err = utils.ReceiveFromBridgeInput(param)
		if err != nil {
			logrus.Errorf("ReceiveFromBridgeInput err:%v", err)
			return
		}
		task := &types.TransactionTask{
			BaseTask:        &types.BaseTask{State: int(types.TxUnInitState)},
			RebalanceId:     task.ID,
			TransactionType: int(types.ReceiveFromBridge),
			ChainId:         param.ChainId,
			ChainName:       param.ChainName,
			From:            param.From,
			To:              param.To,
			Params:          string(paramData),
			InputData:       string(inputData),
		}
		tasks = append(tasks, task)
	}
	err = db.SaveTxTasks(session, tasks)
	if err != nil {
		logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
		return
	}
	return
}

func CreateApproveTask(taskID uint64, param *types.ReceiveFromBridgeParam) (task *types.TransactionTask, err error) {
	inputData, err := utils.ApproveInput(param)
	if err != nil {
		logrus.Errorf("CreateApproveTask err:%v", err)
		return
	}
	paramData, err := json.Marshal(param)
	if err != nil {
		logrus.Errorf("CreateTransactionTask param marshal err:%v", err)
		return
	}
	task = &types.TransactionTask{
		BaseTask:        &types.BaseTask{State: int(types.TxUnInitState)},
		RebalanceId:     taskID,
		TransactionType: int(types.Approve),
		ChainId:         param.ChainId,
		ChainName:       param.ChainName,
		From:            param.From,
		To:              common.Address.String(param.Erc20ContractAddr),
		Params:          string(paramData),
		InputData:       string(inputData),
	}
	return
}
