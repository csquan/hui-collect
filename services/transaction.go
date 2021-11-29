package services

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	signer "github.com/starslabhq/hermes-rebalance/sign"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"time"
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
		clientMap: conf.ClientMap,
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
		case types.TxSignedState:
			return t.handleTransactionSigned(task)
		default:
			logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
		}
	}
	return
}

func (t *Transaction) handleSign(task *types.TransactionTask) (err error) {
	if err = t.approval(); err != nil{
		logrus.Errorf("handleSign approval err:%v", err)
		return
	}
	nonce, err := t.getNonce(task)
	if err != nil {
		logrus.Errorf("handleSign get nonce err:%v", err)
		return
	}
	input := task.InputData
	decimal := 18
	from := task.From
	to := task.To
	GasLimit := "2000000"
	GasPrice := "15000000000"
	Amount := "0"
	quantity := "0"
	receiver := task.To //和to一致

	signRet, err := signer.SignTx(input, decimal, int(nonce), from, to, GasLimit, GasPrice, Amount, quantity, receiver)
	if err != nil {
		return err
	} else {
		if  signRet.Result == true{
			err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
				task.State = int(types.TxAuditState)
				task.Cipher = signRet.Data.Extra.Cipher
				task.EncryptData = signRet.Data.EncryptData
				task.Hash = signRet.Data.Extra.TxHash
				execErr = t.db.UpdateTransactionTask(session, task)
				if execErr != nil {
					logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
					return
				}
				return
			})
		}
	}
	return nil
}

func (t *Transaction) handleAudit(task *types.TransactionTask) (err error) {
	input := task.InputData
	quantity := "0"
	receiver := task.To
	orderID := time.Now().UnixNano() / 1e6    //毫秒

	auditRet, err := signer.AuditTx(input, receiver, quantity, orderID)
	if err != nil {
		return err
	} else {
		if auditRet.Success == true{
			err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
				task.State = int(types.TxValidatorState)
				task.OrderId = orderID
				execErr = t.db.UpdateTransactionTask(session, task)
				if execErr != nil {
					logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
					return
				}
				return
			})
		}
	}
	return nil
}

func (t *Transaction) handleValidator(task *types.TransactionTask) (err error) {
	vRet, err := signer.ValidatorTx(task)
	if err != nil {
		return err

	} else {
		if vRet.OK == true{
			err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
				task.State = int(types.TxSignedState)
				task.SignData = vRet.RawTx
				execErr = t.db.UpdateTransactionTask(session, task)
				if execErr != nil {
					logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
					return
				}
				return
			})
		}
	}
	return nil
}

func (t *Transaction) handleTransactionSigned(task *types.TransactionTask) error {
	client, ok := t.clientMap[task.ChainName]
	if !ok {
		logrus.Fatalf("not find chain client, task:%v", task)
	}
	transaction := &etypes.Transaction{}
	if err := json.Unmarshal([]byte(task.SignData), transaction); err != nil {
		return err
	}
	if err := client.SendTransaction(context.Background(), transaction); err != nil {
		return err
	}

	err := utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
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
		logrus.Fatalf("not find chain client, task:%v", task)
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		return err
	}
	// TODO 如何判断交易已经被记录到链上，如果判断成功或失败。
	if receipt == nil {
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

func (t *Transaction) getNonce(task *types.TransactionTask) (uint64, error) {
	client, ok := t.clientMap[task.ChainName]
	if !ok {
		logrus.Fatalf("not find chain client, task:%v", task)
	}
	return client.NonceAt(context.Background(), common.HexToAddress(task.From), nil)
}

func (t *Transaction) approval() error {
	//TODO
	return nil
}
