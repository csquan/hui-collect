package db

import (
	"github.com/starslabhq/hermes-rebalance/types"
)

func (m *Mysql) GetOpenedPartReBalanceTasks() (tasks []*types.PartReBalanceTask, err error) {
	tasks = make([]*types.PartReBalanceTask, 0)
	_, err = m.engine.Where("state != ? and state != ?", types.Success, types.Failed).Desc("state").Get(&tasks)
	return
}
