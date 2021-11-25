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

func (*Mysql) GetOpenedSignTasks() ([]*types.SignTask, error) {
	return nil, nil
}

func (m *Mysql) GetAssetTransferTasksWithReBalanceId(reBalanceId uint64) (tasks []*types.AssetTransferTask, err error) {
	tasks = make([]*types.AssetTransferTask, 0)
	_, err = m.engine.Where("rebalance_id", reBalanceId).Get(&tasks)
	return
}

func (*Mysql) GetOpenedTransactionTask() (*types.TransactionTask, error) {
	return nil, nil
}

func (*Mysql) GetTxTasks(uint64) ([]*types.TransactionTask, error) {
	return nil, nil
}

func (*Mysql) GetOpenedCrossTasks() ([]*types.CrossTask, error) {
	return nil, nil
}

func (*Mysql) GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*types.CrossTask, error) {
	return nil, nil
}

func (*Mysql) GetCrossSubTasks(crossTaskId uint64) ([]*types.CrossSubTask, error) {
	return nil, nil
}

func (m *Mysql) GetOpenedCrossSubTasks(uint64) ([]*types.CrossSubTask, error) {
	return nil, nil
}
