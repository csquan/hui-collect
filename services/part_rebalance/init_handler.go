package part_rebalance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/clients"

	"math/big"
	"strings"

	"github.com/starslabhq/hermes-rebalance/config"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type initHandler struct {
	db   types.IDB
	conf *config.Config
}

func newInitHandler(db types.IDB, conf *config.Config) *initHandler {
	return &initHandler{
		db:   db,
		conf: conf,
	}
}

func (i *initHandler) CheckFinished(task *types.PartReBalanceTask) (finished bool, nextState types.PartReBalanceState, err error) {
	if !i.conf.IsCheckParams {
		return true, types.PartReBalanceTransferOut, nil
	}
	var params []types.TransactionParamInterface
	params, err = task.ReadTransactionParams(types.SendToBridge)
	if err != nil {
		return
	}
	for _, param := range params {
		sendParam, ok := param.(*types.SendToBridgeParam)
		if !ok {
			logrus.Fatalf("unexpected sendtobridge param:%v", param)
		}
		ok, err = i.checkSendToBridgeParam(sendParam, task)
		if err != nil {
			logrus.Warnf("check send bridge err:%v,tid:%d", err, task.ID)
			return false, 0, err
		}
		if !ok {
			b, _ := json.Marshal(param)
			logrus.Warnf("sendToBridge valut amount not enough task_id:%d,params:%s", task.ID, b)
			return true, types.PartReBalanceFailed, nil
		}
	}
	return true, types.PartReBalanceTransferOut, nil
}

func (i *initHandler) MoveToNextState(task *types.PartReBalanceTask, nextState types.PartReBalanceState) (err error) {
	var tasks []*types.TransactionTask
	if nextState == types.PartReBalanceTransferOut {
		tasks, err = CreateTransactionTask(task, types.SendToBridge)
		if err != nil {
			logrus.Errorf("InvestTask error:%v task:[%v]", err, task)
			return
		}
	}
	err = utils.CommitWithSession(i.db, func(session *xorm.Session) (execErr error) {
		//create next state task
		if nextState == types.PartReBalanceTransferOut {
			if err = i.db.SaveTxTasks(session, tasks); err != nil {
				logrus.Errorf("save transaction task error:%v tasks:[%v]", err, tasks)
				return
			}
		}
		//move to next state
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

func (i *initHandler) GetOpenedTaskMsg(taskId uint64) string {
	return ""
}

func (i *initHandler) checkSendToBridgeParam(param *types.SendToBridgeParam, task *types.PartReBalanceTask) (result bool, err error) {
	client, ok := clients.ClientMap[strings.ToLower(param.ChainName)]
	if !ok {
		logrus.Fatalf("not find chain client, chainName:%s", param.ChainName)
		return
	}
	to := common.HexToAddress(param.To)
	input, err := types.CapableAmountInput()
	if err != nil {
		return
	}
	msg := ethereum.CallMsg{To: &to, Data: input}
	ret, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		err = fmt.Errorf("eth_call getCapableAmount err:%v", err)
		return
	}
	output, err := types.CapableAmountOutput(ret)
	if err != nil {
		err = fmt.Errorf("ret unpack err:%v,ret:%s", err, ret)
		return
	}
	if len(output) == 0 {
		err = fmt.Errorf("getCapableAmount output size zero")
		return
	}
	var amount *big.Int
	if amount, ok = new(big.Int).SetString(param.Amount, 10); !ok {
		err = fmt.Errorf("sendToBridge param error amount:%s", param.Amount)
		return
	}
	valutAmout := output[0].(*big.Int)
	logrus.Infof("sendBridge amout:%s,valutAmout:%s", param.Amount, valutAmout.String())
	if valutAmout.Cmp(amount) >= 0 {
		return true, nil
	}
	errMsg := fmt.Errorf("sendToBridge amount not enough, chain:%s, vault:%s vaultAmount:%s less then param amount:%s",
		param.ChainName, param.To, output[0].(*big.Int).String(), param.Amount)
	alert.Dingding.SendAlert("Amount not Enough",
		alert.TaskFailedContent("小Re", task.ID, "failed", errMsg), nil)
	message, _ := utils.GenPartRebalanceMessage(types.PartReBalanceFailed, fmt.Sprintf("%v", errMsg))
	i.db.UpdatePartReBalanceTaskMessage(task.ID, message)
	return false, nil
}
