package db

import (
	"github.com/ethereum/HuiCollect/types"
)

func (m *Mysql) GetOpenedCollectTask() ([]*types.CollectTxDB, error) {
	tasks := make([]*types.CollectTxDB, 0)
	err := m.engine.Table("t_src_tx").Where("f_collect_state = ?", types.TxReadyCollectState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetCollectTask(id uint64) (*types.CollectTxDB, error) {
	task := &types.CollectTxDB{}
	ok, err := m.engine.Table("t_src_tx").Where("f_id = ?", id).Limit(1).Get(task)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return task, nil
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

func (m *Mysql) GetOpenedCheckTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\"  and f_state in (?)", types.TxBroadcastState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedUpdateAccountTasks() ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_error = \"\" and f_state in (?)", types.TxCheckState).Find(&tasks)
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

func (m *Mysql) GetAccountBalance(accountAddr string, contratAddr string) (*types.Account, error) {
	task := &types.Account{}
	ok, err := m.engine.Table("t_account").Where("f_addr = ? and f_contractAddr = ?", accountAddr, contratAddr).Limit(1).Get(task)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (m *Mysql) GetTokenInfo(contratAddr string, chain string) (*types.Token, error) {
	token := &types.Token{}
	ok, err := m.engine.Table("t_token").Where("f_address = ? and f_chain = ?", contratAddr, chain).Limit(1).Get(token)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return token, nil
}
