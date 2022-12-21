package services

import (
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/go-xorm/xorm"
)

type BroadcastService struct {
	db     types.IDB
	config *config.Config
}

func NewBoradcastService(db types.IDB, c *config.Config) *BroadcastService {
	return &BroadcastService{
		db:     db,
		config: c,
	}
}

func (c *BroadcastService) BroadcastTx(task *types.TransactionTask) (finished bool, err error) {
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		//1.广播交易 2.广播交易后需要根据hash查询receipt，确定查询到结果后再更新状态 3.产生叮叮状态转换消息

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("add cross sub tasks err:%v", err)
	}
	return true, nil
}

func (c *BroadcastService) tgalert(task *types.TransactionTask) {

}

func (c *BroadcastService) Run() error {
	tasks, err := c.db.GetOpenedBroadcastTasks()
	if err != nil {
		return fmt.Errorf("get tasks for broadcast err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.BroadcastTx(task)
		if err == nil {
			c.tgalert(task)
		}
	}
	return nil
}

func (c BroadcastService) Name() string {
	return "Broadcast"
}
