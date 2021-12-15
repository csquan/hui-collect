package db

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

func (m *Mysql) GetPartReBalanceTaskByFullRebalanceID(fullRebalanceID uint64) (task *types.PartReBalanceTask, err error) {
	task = &types.PartReBalanceTask{}
	ok, err := m.engine.Where("f_full_rebalance_id = ?", fullRebalanceID).Desc("f_id").Limit(1).Get(task)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (m *Mysql) GetOpenedPartReBalanceTasks() (tasks []*types.PartReBalanceTask, err error) {
	tasks = make([]*types.PartReBalanceTask, 0)
	err = m.engine.Where("f_state != ? and f_state != ?",
		types.PartReBalanceSuccess,
		types.PartReBalanceFailed).
		Desc("f_state").
		Find(&tasks)
	return
}

func (m *Mysql) GetOpenedFullReBalanceTasks() (tasks []*types.FullReBalanceTask, err error) {
	tasks = make([]*types.FullReBalanceTask, 0)
	err = m.engine.Where("f_state != ? and f_state != ? and f_state != ?",
		types.FullReBalanceSuccess,
		types.FullReBalanceFailed,
		types.FullReBalanceParamsCalc).
		Desc("f_state").
		Find(&tasks)
	return
}

func (m *Mysql) GetTransactionTasksWithReBalanceId(reBalanceId uint64, transactionType types.TransactionType) (tasks []*types.TransactionTask, err error) {
	tasks = make([]*types.TransactionTask, 0)
	err = m.engine.Table("t_transaction_task").Where("f_rebalance_id = ? and f_type= ?", reBalanceId, transactionType).Find(&tasks)
	return
}

func (m *Mysql) GetOpenedTransactionTask() (tasks []*types.TransactionTask, err error) {
	tasks = make([]*types.TransactionTask, 0)
	err = m.engine.Where("f_state != ? and f_state != ?",
		types.TxSuccessState,
		types.TxFailedState).
		Asc("f_from", "f_nonce").
		Find(&tasks)
	return
}

func (m *Mysql) GetOpenedCrossTasks() ([]*types.CrossTask, error) {
	tasks := make([]*types.CrossTask, 0) //state:0等待创建子任务
	err := m.engine.Table("t_cross_task").Where("f_state in (?,?)", types.ToCreateSubTask, types.SubTaskCreated).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*types.CrossTask, error) {
	tasks := make([]*types.CrossTask, 0)
	err := m.engine.Where("f_rebalance_id = ?", reBalanceId).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetCrossSubTasks(parentTaskId uint64) ([]*types.CrossSubTask, error) {
	tasks := make([]*types.CrossSubTask, 0)
	err := m.engine.Table("t_cross_sub_task").Where("f_parent_id = ?", parentTaskId).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (m *Mysql) GetOpenedCrossSubTasks(parentTaskId uint64) ([]*types.CrossSubTask, error) {
	tasks := make([]*types.CrossSubTask, 0)
	err := m.engine.Table("t_cross_sub_task").Where("f_parent_id = ? and f_state != ?", parentTaskId, types.Crossed).Find(&tasks) //state:2 跨链任务完成
	if err != nil {
		return nil, err
	}
	return tasks, nil
}


func (m *Mysql) GetTransactionTasksWithFullRebalanceId(fullReBalanceId uint64, transactionType types.TransactionType) ([]*types.TransactionTask, error) {
	tasks := make([]*types.TransactionTask, 0)
	err := m.engine.Table("t_transaction_task").Where("f_full_rebalance_id = ? and f_type = ?", fullReBalanceId, transactionType).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (m *Mysql) GetTokens() ([]*types.Token, error) {
	tokens := make([]*types.Token, 0)
	err := m.engine.Find(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
func (m *Mysql) GetCurrency() ([]*types.Currency, error) {
	currency := make([]*types.Currency, 0)
	err := m.engine.Find(&currency)
	if err != nil {
		return nil, err
	}
	return currency, nil
}



