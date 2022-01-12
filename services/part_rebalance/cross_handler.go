package part_rebalance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"math/big"
	"time"

	"github.com/starslabhq/hermes-rebalance/clients"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type crossHandler struct {
	db        types.IDB
	clientMap map[string]*ethclient.Client
	start int64
}

func (c *crossHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	if c.start == 0 {
		c.start = time.Now().Unix()
	}
	crossTasks, err := c.db.GetCrossTasksByReBalanceId(task.ID)
	if err != nil {
		logrus.Errorf("get cross task for rebalance [%v] failed, err:%v", task, err)
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
		utils.CostLog(c.start, task.ID, "跨链耗时")
		c.start = 0
		nextState = types.PartReBalanceTransferIn
	} else {
		c.start = 0
		nextState = types.PartReBalanceFailed
	}

	return true, nextState, nil
}

func (c *crossHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceTransferIn {
		tasks, err = CreateTransactionTask(task, types.ReceiveFromBridge)
		if err != nil {
			logrus.Errorf("InvestTask error:%v task:[%v]", err, task)
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

func SetNonceAndGasPrice(tasks []*types.TransactionTask) (result []*types.TransactionTask, err error) {
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
				if nonce, err = types.GetNonce(t.From, t.ChainName); err != nil {
					return
				}
				if gasPrice, err = types.GetGasPrice(t.ChainName); err != nil {
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

func CreateApproveTask(taskID uint64, param *types.ReceiveFromBridgeParam) (task *types.TransactionTask, err error) {
	client, ok := clients.ClientMap[param.ChainName]
	if !ok {
		err = fmt.Errorf("rpc client for chain:[%v] not found", param.ChainName)
		return
	}

	data, err := types.AllowanceInput(param.From, param.To)
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

	out, err := types.AllowanceOutput(outBytes)
	if err != nil {
		err = fmt.Errorf("decode allowance error:%v", err)
		return
	}
	var amount *big.Int
	if amount, ok = new(big.Int).SetString(param.Amount, 10); !ok {
		err = fmt.Errorf("CreateApproveTask param error")
		return
	}
	// do not need to approve
	if out[0].(*big.Int).Cmp(amount) >= 0 {
		return nil, nil
	}

	inputData, err := types.ApproveInput(param.To)
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

const openedTemp = `
# cross_run_timeout
# task
{{ range . }}
- id: {{.ID}}
- rebalance_id: {{.RebalanceId}}
- chain: {{.ChainFrom}} -> {{.ChainTo}}
- addr: {{.ChainFromAddr}} -> {{.ChainToAddr}}
- currency: {{.CurrencyFrom}} -> {{.CurrencyTo}}
- amount: {{.Amount}}
- state: {{.State}}
{{- end}}
`

func getOpenedTaskMsg(tasks []*types.CrossTask) string {
	t := template.New("cross_timeout_msg")
	temp := template.Must(t.Parse(openedTemp))
	buf := &bytes.Buffer{}
	err := temp.Execute(buf, &tasks)
	if err != nil {
		logrus.Errorf("exec cross timeout msg err:%v", err)
	}
	return buf.String()
}

func (c *crossHandler) GetOpenedTaskMsg(taskId uint64) string {
	tasks, err := c.db.GetCrossTasksByReBalanceId(taskId)
	if err != nil {
		logrus.Errorf("get cross task err:%v", err)
	}
	var opened []*types.CrossTask
	for _, t := range tasks {
		if t.State != types.TaskSuc {
			opened = append(opened, t)
		}
	}
	return getOpenedTaskMsg(opened)
}
