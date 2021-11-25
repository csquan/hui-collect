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
)

type TransactionState int

const (
	TxInit TransactionState = iota
	TxUnSigned

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
	case TxUnSigned:
		return t.handleTransactionUnSigned(task)
	case TxSigned:
		return t.handleTransactionSigned(task)
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", task.State, task.ID)
	}

	return
}

func (t *Transaction) handleTransactionUnSigned(task *types.TransactionTask) error {
	txRow, err := signTx(task)
	if err != nil {
		return err
	}
	data, err := json.Marshal(txRow)
	if err != nil {
		return err
	}
	task.SignData = data
	task.State = int(TxSigned)
	return t.db.UpdateTransactionTask(task)
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

func signTx(task *types.TransactionTask) (*etypes.Transaction, error) {
	return nil, nil
}
func broadcast(task *types.TransactionTask) error {
	return nil
}
func txCheck(task *types.TransactionTask) bool {
	return true
}
