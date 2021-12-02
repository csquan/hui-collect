package services

import (
	"context"
	"github.com/starslabhq/hermes-rebalance/clients"
	"time"

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
	GasLimit := t.config.SendConf.GasLimit
	GasPrice := task.GasPrice
	Amount := t.config.SendConf.Amount
	quantity := t.config.SendConf.Quantity
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
	quantity := t.config.SendConf.Quantity
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

func (t *Transaction) handleTransactionCheck(task *types.TransactionTask) error {
	client, ok := t.clientMap[task.ChainName]
	if !ok {
		logrus.Warnf("not find chain client, task:%v", task)
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		logrus.Warnf("hash not found, task:%v", task)
		err = nil
	}
	if receipt == nil {
		transaction, err := types.DecodeTransaction(task.SignData)
		if err != nil {
			logrus.Errorf("DecodeTransaction err:%v task:%v", err, task)
			return err
		}
		if err := client.SendTransaction(context.Background(), transaction); err != nil {
			logrus.Errorf("SendTransaction err:%v task:%v", err, task)
			return err
		}
		return nil
	}
	if receipt.Status == 1 {
		task.State = int(types.TxSuccessState)
	} else if receipt.Status == 0 {
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
	return err
}
