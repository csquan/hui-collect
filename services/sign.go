package services

import (
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/go-xorm/xorm"
)

type SignService struct {
	db     types.IDB
	config *config.Config
}

func NewSignService(db types.IDB, c *config.Config) *SignService {
	return &SignService{
		db:     db,
		config: c,
	}
}

func (c *SignService) SignTx(task *types.TransactionTask) (finished bool, err error) {
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		//1.调用签名接口 2。依照结果更新task状态

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("add cross sub tasks err:%v", err)
	}
	return true, nil
}

func (c *SignService) tgalert(task *types.TransactionTask) {

}

func (c *SignService) Run() error {
	tasks, err := c.db.GetOpenedSignTasks()
	if err != nil {
		return fmt.Errorf("get tasks for sign err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.SignTx(task)
		if err == nil {
			c.tgalert(task)
		}
	}
	return nil
}

func (c SignService) Name() string {
	return "Sign"
}
