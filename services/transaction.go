package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/starslabhq/hermes-rebalance/alert"

	"github.com/starslabhq/hermes-rebalance/clients"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	signer "github.com/starslabhq/hermes-rebalance/sign"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type Transaction struct {
	db        types.IDB
	config    *config.Config
	clientMap map[string]*ethclient.Client
}

func NewTransactionService(db types.IDB, conf *config.Config) (p *Transaction, err error) {
	p = &Transaction{
		db:        db,
		config:    conf,
		clientMap: clients.ClientMap,
	}
	return
}

func (t *Transaction) Name() string {
	return "transaction"
}
func (t *Transaction) Run() (err error) {
	tasks, err := t.db.GetOpenedTransactionTask()
	if err != nil {
		return
	}
	if len(tasks) == 0 {
		logrus.Infof("no available part Transaction task.")
		return
	}

	for _, task := range tasks {
		task.Message = utils.GenTxMessage(task.State, "")
		err = func(task *types.TransactionTask) error {
			switch types.TransactionState(task.State) {
			case types.TxUnInitState:
				return t.handleSign(task)
			case types.TxAuditState:
				return t.handleAudit(task)
			case types.TxValidatorState:
				return t.handleValidator(task)
			case types.TxCheckReceiptState:
				return t.handleTransactionCheck(task)
			default:
				logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
				return nil
			}
		}(task)

		if err != nil {
			t.db.UpdateTransactionTaskMessage(task.ID, utils.GenTxMessage(task.State, fmt.Sprintf("%v", err)))
			return
		}
	}
	return
}

func (t *Transaction) handleSign(task *types.TransactionTask) (err error) {
	nonce := task.Nonce
	input := task.InputData
	decimal := 18
	from := task.From
	to := task.To
	GasLimit := task.GasLimit
	if GasLimit == "" {
		GasLimit = "2000000"
	}
	GasPrice := task.GasPrice
	Amount := task.Amount
	if Amount == "" {
		Amount = "0"
	}
	quantity := task.Quantity
	if quantity == "" {
		quantity = "0"
	}
	receiver := task.To //和to一致

	signRet, err := signer.SignTx(input, decimal, int(nonce), from, to, GasLimit, GasPrice, Amount, quantity, receiver, task.ChainName)

	if err == nil && signRet.Result == true {
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(types.TxAuditState)
			task.Cipher = signRet.Data.Extra.Cipher
			task.EncryptData = signRet.Data.EncryptData
			task.Hash = signRet.Data.Extra.TxHash
			execErr = t.db.UpdateTransactionTask(session, task)
			if execErr != nil {
				logrus.Errorf("update sign task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}

	return err
}

func (t *Transaction) handleAudit(task *types.TransactionTask) (err error) {
	input := task.InputData
	quantity := task.Quantity
	if quantity == "" {
		quantity = "0"
	}
	receiver := task.To
	orderID := time.Now().UnixNano() / 1e6 //毫秒

	auditRet, err := signer.AuditTx(input, receiver, quantity, orderID, task.ChainName)

	if err == nil && auditRet.Success == true {
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(types.TxValidatorState)
			task.OrderId = orderID
			execErr = t.db.UpdateTransactionTask(session, task)
			if execErr != nil {
				logrus.Errorf("update  audit task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}
	return err
}

func (t *Transaction) handleValidator(task *types.TransactionTask) (err error) {
	vRet, err := signer.ValidatorTx(task)

	if err == nil && vRet.OK == true {
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(types.TxCheckReceiptState)
			task.SignData = vRet.RawTx
			execErr = t.db.UpdateTransactionTask(session, task)
			if execErr != nil {
				logrus.Errorf("update  validator task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	} else {
		_ = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(types.TxAuditState) //失败了则退回安审状态，下次重新安审

			execErr = t.db.UpdateTransactionTask(session, task)
			if execErr != nil {
				logrus.Errorf("update  validator task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}

	return err
}

func (t *Transaction) handleTransactionSigned(task *types.TransactionTask) error {
	client, ok := t.clientMap[task.ChainName]
	if !ok {
		logrus.Errorf("not find chain client, task:%v", task)
	}
	transaction, err := types.DecodeTransaction(task.SignData)
	if err != nil {
		logrus.Errorf("DecodeTransaction err:%v task:%v", err, task)
		return err
	}
	if err = client.SendTransaction(context.Background(), transaction); err != nil {
		//TODO nonce too low, 重新走签名流程？
		logrus.Errorf("SendTransaction err:%v task:%v", err, task)
		return err
	}

	err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		task.State = int(types.TxCheckReceiptState)
		execErr = t.db.UpdateTransactionTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part audit task error:%v task:[%v]", execErr, task)
			return execErr
		}
		return nil
	})
	return err
}

var txErrFormat = `
chain:%s
hash:%s
`

func (t *Transaction) isTxBlockSafe(chain string, txHeight, curHeight uint64) bool {
	chaiInfo := t.config.MustGetChainInfo(chain)
	blockSafe := chaiInfo.BlockSafe
	return curHeight-txHeight > uint64(blockSafe)
}

func (t *Transaction) handleTransactionCheck(task *types.TransactionTask) error {
	clikey := strings.ToLower(task.ChainName)
	client, ok := t.clientMap[clikey]
	if !ok {
		logrus.Warnf("not find chain client, task:%v", task)
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		logrus.Warnf("hash not found, task:%v,err:%v,hash:%s", task, err, task.Hash)
		err = nil
	}
	if receipt == nil {
		transaction, err := types.DecodeTransaction(task.SignData)
		if err != nil {
			logrus.Errorf("DecodeTransaction err:%v task:%v", err, task)
			return err
		}
		if err := client.SendTransaction(context.Background(), transaction); err != nil {
			logrus.Warnf("SendTransaction err:%v task:%v", err, task)
			return nil
		}
		return nil
	}
	curh, err := client.BlockNumber(context.Background())
	if err != nil {
		logrus.Warnf("get block num err:%v", err)
		return nil
	}
	txh := receipt.BlockNumber.Uint64()
	if t.isTxBlockSafe(task.ChainName, txh, curh) {
		if receipt.Status == 1 {
			task.State = int(types.TxSuccessState)
		} else if receipt.Status == 0 {
			alert.Dingding.SendAlert("transaction failed",
				alert.TaskFailedContent("transaction", task.ID, "CheckReceipt", fmt.Errorf(txErrFormat, task.ChainName, task.Hash)), nil)
			task.State = int(types.TxFailedState)
		}
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			execErr = t.db.UpdateTransactionTask(session, task)
			if execErr != nil {
				logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
				return execErr
			}
			return nil
		})
	} else {
		logrus.Debugf("tx not safe txh:%d,curh:%d,hash:%s,chain:%s", txh, curh, task.Hash, task.ChainName)
	}
	return err
}
