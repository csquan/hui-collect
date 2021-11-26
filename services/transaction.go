package services

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	signer "github.com/starslabhq/hermes-rebalance/sign"
	"github.com/starslabhq/hermes-rebalance/utils"
	"github.com/go-xorm/xorm"
)

type TransactionState int

const (
	TxInit TransactionState = iota
	SignState
	AuditState
	ValidatorState
	TxSigned
	TxCheckReceipt
	TxSuccess
	TxFailed
)

type Transaction struct {
	db     types.IDB
	config *config.Config
	client ethclient.Client
}

func NewTransactionService(db types.IDB, conf *config.Config) (p *AssetTransfer, err error) {
	p = &AssetTransfer{
		db:     db,
		config: conf,
	}
	return
}

func (t *Transaction) Name() string {
	return "transaction"
}
func (t *Transaction) Run() (err error) {
	task, err := t.db.GetOpenedTransactionTask()
	if err != nil {
		return
	}
	if task == nil {
		logrus.Infof("no available Transaction task.")
	}
	switch TransactionState(task.State) {
	case SignState:
		return t.handleSign(task)
	case AuditState:
		return t.handleAudit(task)
	case ValidatorState:
		return t.handleValidator(task)
	case TxSigned:
		return t.handleTransactionSigned(task)
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", task.State, task.ID)
	}

	return
}

func (t *Transaction) handleSign(task *types.TransactionTask) (err error) {
	input := task.Input_data
	decimal := task.Decimal
	nonce := task.Nonce
	from := task.From  //这个是签名机固定的地址？？
	to := task.To
	GasLimit :=""
	GasPrice :=""
	Amount := "0"
	quantity:= string(task.Value)
	receiver:= task.To  //和to一致

	signRet,err := signer.SignTx(input, decimal, nonce, from, to, GasLimit, GasPrice, Amount, quantity, receiver)
	if err != nil {
		return err
	}else {
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(AuditState)
			task.Cipher = signRet.Data.Extra.Cipher
			task.EncryptData = signRet.Data.EncryptData
			task.Hash = signRet.Data.Extra.TxHash

			execErr = t.db.UpdateTxTask(session, task)
			if execErr != nil {
				logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}
	return nil
}

func (t *Transaction) handleAudit(task *types.TransactionTask) (err error) {
	input := task.Input_data
	quantity := string(task.Value)
	receiver := task.To
	orderID,_ := t.db.GetOrderID()

	defer utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		t.db.UpdateOrderID(session,orderID+1)
		return
	})

	_, err = signer.AuditTx(input,receiver,quantity,orderID)
	if err != nil {
		return err
	}else{
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(ValidatorState)
			execErr = t.db.UpdateTxTask(session, task)
			if execErr != nil {
				logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}
	return nil
}

func (t *Transaction) handleValidator(task *types.TransactionTask) (err error) {
	input := task.Input_data
	quantity := string(task.Value)
	orderID,_ := t.db.GetOrderID()
	to := task.To

	defer utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
		t.db.UpdateOrderID(session,orderID+1)
		return
	})

	vRet,err := signer.ValidatorTx(input, to, quantity,orderID)
	if err != nil  {
		return err
	}else{
		err = utils.CommitWithSession(t.db, func(session *xorm.Session) (execErr error) {
			task.State = int(TxSigned)
			task.Input_data = vRet.RawTx
			execErr = t.db.UpdateTxTask(session, task)
			if execErr != nil {
				logrus.Errorf("update part audit task error:%v task:[%v]", err, task)
				return
			}
			return
		})
	}
	return nil
}


func (t *Transaction) handleTransactionSigned(task *types.TransactionTask) error {
	transaction := &etypes.Transaction{}
	if err := json.Unmarshal(task.SignData, transaction); err != nil {
		return err
	}
	if err := t.client.SendTransaction(context.Background(), transaction); err != nil {
		return err
	}
	task.State = int(TxCheckReceipt)
	return t.db.UpdateTransactionTask(task)
}

func (t *Transaction) handleTransactionCheck(task *types.TransactionTask) error {
	receipt, err := t.client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		return err
	}
	// TODO 如何判断交易已经被记录到链上，如果判断成功或失败。
	if receipt == nil {
		return nil
	}
	if receipt.Status == 1 {
		task.State = int(TxSuccess)
	} else if receipt.Status == 0 {
		task.State = int(TxFailed)
	}
	return t.db.UpdateTransactionTask(task)
}

func broadcast(task *types.TransactionTask) error {
	return nil
}
func txCheck(task *types.TransactionTask) bool {
	return true
}
