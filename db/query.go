package db

import (
	"github.com/ethereum/fat-tx/types"
)

func (m *Mysql) GetOpenedAssemblyTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_state in (?)", types.TxInitState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedSignTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_state in (?)", types.TxAssmblyState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedBroadcastTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_state in (?)", types.TxSignState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedCheckReceiptTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_state in (?)", types.TxBroadcastState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetTaskNonce(from string) (*types.TransactionTask, error) {
	task := &types.TransactionTask{}
	ok, err := m.engine.Table("t_transaction_task").Where("f_from = ?", from).Desc("f_nonce").Limit(1).Get(task)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return task, nil
}
