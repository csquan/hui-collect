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

func (*Mysql) GetAssetTransferTasksWithReBalanceId(reBalanceId uint64) ([]*types.AssetTransferTask, error) {
	return nil, nil
}

func (*Mysql) GetOpenedTransactionTask() (*types.TransactionTask, error) {
	return nil, nil
}
func (*Mysql) GetTxTasks(uint64) ([]*types.TransactionTask, error) {
	return nil, nil
}
