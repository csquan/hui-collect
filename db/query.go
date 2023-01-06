package db

import (
	"fmt"
	"github.com/ethereum/HuiCollect/types"
)

func (m *Mysql) GetMonitorCountInfo(Addr string) (int, error) {
	count := 0
	sql := fmt.Sprintf("select count(*) from t_monitor where addr = \"%s\";", Addr)
	ok, err := m.engine.SQL(sql).Limit(1).Get(&count)
	if err != nil {
		return count, err
	}
	if !ok {
		return count, nil
	}

	return count, err
}

func (m *Mysql) GetMonitorHeightInfo(Addr string) (int, error) {
	height := 0
	sql := fmt.Sprintf("select height from t_monitor where addr = \"%s\";", Addr)
	ok, err := m.engine.SQL(sql).Limit(1).Get(&height)
	if err != nil {
		return height, err
	}
	if !ok {
		return height, nil
	}

	return height, err
}

func (m *Mysql) GetMonitorCollectTask(addr string, height int) ([]*types.TxErc20, error) {
	tasks := make([]*types.TxErc20, 0)
	err := m.engine.Table("tx_erc20").Where("receiver = ? and block_num > ?", addr, height).OrderBy("block_num").Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedCollectTask() ([]*types.CollectTxDB, error) {
	tasks := make([]*types.CollectTxDB, 0)
	err := m.engine.Table("t_src_tx").Where("collect_state = ?", types.TxReadyCollectState).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetCollectTask(id uint64) (*types.CollectTxDB, error) {
	task := &types.CollectTxDB{}
	ok, err := m.engine.Table("t_src_tx").Where("id = ?", id).Limit(1).Get(task)
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
