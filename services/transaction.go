package services

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type TransactionState int

const (
	TxInit TransactionState = iota
	TxUnSigned

	TxSigned

	TxSuccess
	TxFailed
)

type Transaction struct {
	db     types.IDB
	config *config.Config
}

func NewTransactionService(db types.IDB, conf *config.Config) (p *Transfer, err error) {
	p = &Transfer{
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
	switch AssetTransferState(task.State) {
	case AssetTransferInit:
		return t.handleTransactionInit(task)
	case AssetTransferOngoing:
		return t.handleTransactionOngoing(task)
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", task.State, task.ID)
	}

	return
}

func (t *Transaction) handleTransactionInit(task *types.TransactionTask) error {
	return nil
}

func (t *Transaction) handleTransactionOngoing(task *types.TransactionTask) error {
	return nil
}
