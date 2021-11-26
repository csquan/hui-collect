package db

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

func (m *Mysql) GetOpenedPartReBalanceTasks() (tasks []*types.PartReBalanceTask, err error) {
	tasks = make([]*types.PartReBalanceTask, 0)
	_, err = m.engine.Where("state != ? and state != ?", types.PartReBalanceSuccess, types.PartReBalanceFailed).Desc("state").Get(&tasks)
	return
}

func (*Mysql) GetOpenedAssetTransferTasks() ([]*types.AssetTransferTask, error) {
	return nil, nil
}

func (m *Mysql) GetAssetTransferTasksWithReBalanceId(reBalanceId uint64, transferType int) (tasks []*types.AssetTransferTask, err error) {
	tasks = make([]*types.AssetTransferTask, 0)
	_, err = m.engine.Where("rebalance_id = ? and transfer_type = ?", reBalanceId, transferType).Get(&tasks)
	return
}

func (m *Mysql) GetOpenedTransactionTask() (tasks []*types.TransactionTask, err error) {
	//在交易表中找到所有状态不为SignState的交易任务
	tasks = make([]*types.TransactionTask, 0)
	err = m.engine.Table("transaction_task").Where("state < ?", 4).Find(&tasks)
	return
}

func (*Mysql) GetTxTasks(uint64) ([]*types.TransactionTask, error) {
	return nil, nil
}

func (m *Mysql) GetOpenedCrossTasks() ([]*types.CrossTask, error) {
	tasks := make([]*types.CrossTask, 0) //state:0等待创建子任务
	err := m.engine.Where("state = ?", 0).Find(&tasks)
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

func (*Mysql) GetCrossSubTasks(parentTaskId uint64) ([]*types.CrossSubTask, error) {
	return nil, nil
}

func (m *Mysql) GetOpenedCrossSubTasks(parentTaskId uint64) ([]*types.CrossSubTask, error) {
	tasks := make([]*types.CrossSubTask, 0)
	err := m.engine.Where("parent_id = ? and state != ?", parentTaskId, 2).Find(&tasks) //state:2 跨链任务完成
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (m *Mysql) GetOrderID() (int, error) {
	OrderId := 0
	_, err := m.engine.Get(&OrderId)
	if err != nil {
		return OrderId, err
	}
	return OrderId, nil

}
