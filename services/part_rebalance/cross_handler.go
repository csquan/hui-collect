package part_rebalance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"math/big"
)

type crossHandler struct {
	db        types.IDB
	clientMap map[string]*ethclient.Client
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
	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceTransferIn {
		tasks, err = c.CreateReceiveFromBridgeTask(task)
		if err != nil {
			logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
			return
		}
		if tasks, err = c.SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
			logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
			return
		}
	}
	err = utils.CommitWithSession(c.db, func(session *xorm.Session) (execErr error) {
		//create next state task
		if nextState == types.PartReBalanceTransferIn {
			if err = c.db.SaveTxTasks(session, tasks); err != nil {
				logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
				return
			}
			return
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

func (c *crossHandler) CreateReceiveFromBridgeTask(rbTask *types.PartReBalanceTask) (tasks []*types.TransactionTask, err error) {
	params, err := rbTask.ReadParams()
	if err != nil {
		return
	}
	var paramData, inputData []byte
	for _, param := range params.ReceiveFromBridgeParams {
		var approveTask *types.TransactionTask
		if approveTask, err = c.CreateApproveTask(rbTask.ID, param); err != nil {
			logrus.Errorf("CreateApproveTask err:%v", err)
			return
		}
		if approveTask != nil {
			tasks = append(tasks, approveTask)
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
			RebalanceId:     rbTask.ID,
			TransactionType: int(types.ReceiveFromBridge),
			ChainId:         param.ChainId,
			ChainName:       param.ChainName,
			From:            param.From,
			To:              param.To,
			Params:          string(paramData),
			InputData:       hexutil.Encode(inputData),
		}
		tasks = append(tasks, task)
	}
	return
}

func (c *crossHandler) SetNonceAndGasPrice(tasks []*types.TransactionTask) (result []*types.TransactionTask, err error) {
	//group by From
	m := make(map[string][]*types.TransactionTask)
	for _, task := range tasks {
		if l, ok := m[task.From]; ok {
			m[task.From] = append(l, task)
		} else {
			m[task.From] = []*types.TransactionTask{task}
		}
	}
	for _, l := range m {
		var nonce uint64
		var gasPrice *big.Int
		for i, t := range l {
			if i == 0 {
				if nonce, err = c.getNonce(t); err != nil {
					return
				}
				if gasPrice, err = c.getGasPrice(t); err != nil {
					return
				}
			} else {
				nonce++
			}
			t.Nonce = nonce
			t.GasPrice = gasPrice.String()
			result = append(result, t)
		}
	}
	return
}

func (c *crossHandler) CreateApproveTask(taskID uint64, param *types.ReceiveFromBridgeParam) (task *types.TransactionTask, err error) {
	client, ok := c.clientMap[param.ChainName]
	if !ok {
		err = fmt.Errorf("rpc client for chain:[%v] not found", param.ChainName)
		return
	}

	data, err := utils.AllowanceInput(param)
	if err != nil {
		err = fmt.Errorf("encode allowance input error:%v", err)
		return
	}

	outBytes, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &param.Erc20ContractAddr,
		Data: data,
	}, nil)
	if err != nil {
		err = fmt.Errorf("get allowance error:%v", err)
		return
	}

	out, err := utils.AllowanceOutput(outBytes)
	if err != nil {
		err = fmt.Errorf("decode allowance error:%v", err)
		return
	}

	// do not need to approve
	if out[0].(*big.Int).Cmp(param.Amount) >= 0 {
		return nil, nil
	}

	//if approve, err = c.db.GetApprove(common.Address.String(param.Erc20ContractAddr), param.To); err != nil {
	//	logrus.Errorf("GetApprove err:%v", err)
	//	return
	//}
	//if approve != nil {
	//	return nil, nil
	//}
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
		ContractAddress: param.To,
		Params:          string(paramData),
		InputData:       hexutil.Encode(inputData),
	}
	return
}

func (c *crossHandler) getNonce(task *types.TransactionTask) (uint64, error) {
	client, ok := c.clientMap[task.ChainName]
	if !ok {
		logrus.Fatalf("not find chain client, task:%v", task)
	}
	//TODO client.PendingNonceAt() ?
	return client.NonceAt(context.Background(), common.HexToAddress(task.From), nil)
}

func (c *crossHandler) getGasPrice(task *types.TransactionTask) (*big.Int, error) {
	client, ok := c.clientMap[task.ChainName]
	if !ok {
		logrus.Fatalf("not find chain client, task:%v", task)
	}
	return client.SuggestGasPrice(context.Background())
}
