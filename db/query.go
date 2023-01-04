package db

import (
	"github.com/ethereum/Hui-TxState/types"
)

func (m *Mysql) GetOpenedCollectTask() ([]*types.CollectTxDB, error) {
	tasks := make([]*types.CollectTxDB, 0)
	err := m.engine.Table("t_src_tx").Where("collect_state = ?", types.TxReadyCollectState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

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
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\"  and f_state in (?)", types.TxAssmblyState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedBroadcastTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\"  and f_state in (?)", types.TxSignState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedCheckReceiptTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\"  and f_state in (?)", types.TxBroadcastState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedOkCallBackTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\" and f_state in (?)", types.TxCheckState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedFailCallBackTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error != \"\" and f_state in (?,?)", types.TxSignState, types.TxBroadcastState).Find(&tasks)
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

func (m *Mysql) GetSpecifyTasks(input_task *types.TransactionTask) (*types.TransactionTask, error) {
	task := &types.TransactionTask{}
	ok, err := m.engine.Table("t_transaction_task").Where("f_from = ? and f_uuid = ? and f_request_id = ? and f_state >= 3", input_task.From, input_task.UUID, input_task.RequestId).Limit(1).Get(task)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return task, nil
}
