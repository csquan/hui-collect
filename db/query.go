package db

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

func (m *Mysql) GetOpenedPartReBalanceTasks() (tasks []*types.PartReBalanceTask, err error) {
	tasks = make([]*types.PartReBalanceTask, 0)
	err = m.engine.Where("f_state != ? and f_state != ?",
		types.PartReBalanceSuccess,
		types.PartReBalanceFailed).
		Desc("f_state").
		Find(&tasks)
	return
}

func (m *Mysql) GetTransactionTasksWithReBalanceId(reBalanceId uint64, transactionType types.TransactionType) (tasks []*types.TransactionTask, err error) {
	tasks = make([]*types.TransactionTask, 0)
	_, err = m.engine.Where("f_rebalance_id = ? and f_type= ?", reBalanceId, transactionType).Get(&tasks)
	return
}

func (m *Mysql) GetOpenedTransactionTask() (tasks []*types.TransactionTask, err error) {
	tasks = make([]*types.TransactionTask, 0)
	err = m.engine.Where("f_state != ? and f_state != ?",
		types.TxSuccessState,
		types.TxFailedState).
		Desc("f_state"). //根据state倒叙可以确保授权task先执行
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
	err := m.engine.Where("rebalance_id = ?", reBalanceId).Find(&tasks)
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

func (m *Mysql) GetApprove(token, spender string) (*types.ApproveRecord, error) {
	var approve *types.ApproveRecord
	_, err := m.engine.Table("t_approve").Where("f_token = ? and f_spender = ?", token, spender).Get(approve)
	if err != nil {
		return nil, err
	}
	return approve, nil
}
