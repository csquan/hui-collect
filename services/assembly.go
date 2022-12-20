package services

import (
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

//type CrossMsg struct {
//	Stage    string
//	Task     *types.CrossTask
//	SubTasks []*types.CrossSubTask
//}

type AssemblyService struct {
	db     types.IDB
	config *config.Config
}

func NewAssemblyService(db types.IDB, c *config.Config) *AssemblyService {
	return &AssemblyService{
		db:     db,
		config: c,
	}
}

func (c *AssemblyService) AssemblyTx() (finished bool, err error) {
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {

		c.stateChanged(types.TxAssmblyState)

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("add cross sub tasks err:%v", err)
	}
	return true, nil
}

func (c *AssemblyService) stateChanged(next types.TransactionState) {
	//var (
	//	msg string
	//	err error
	//)
	//msg, err = createCrossMesg("cross_finished", task, subTasks)
	//if err != nil {
	//	logrus.Errorf("create subtask_finished msg err:%v,state:%d,tid:%d", err, next, task.ID)
	//}
	//
	//err = alert.Dingding.SendMessage("cross", msg)
	//if err != nil {
	//	logrus.Errorf("send message err:%v,msg:%s", err, msg)
	//}

}

func (c *AssemblyService) Run() error {
	tasks, err := c.db.GetOpenedAssemblyTasks()
	if err != nil {
		return fmt.Errorf("get tasks for assembly err:%v", err)
	}

	if len(tasks) == 0 {
		logrus.Infof("no tasks for assembly")
		return nil
	}

	for _, task := range tasks {
		switch task.State {
		case int(types.TxInitState):
			_, err := c.AssemblyTx()
			if err == nil {
				c.stateChanged(types.TxAssmblyState)
			}
		default:
			return fmt.Errorf("state:[%v] unknow:%d", task.State)
		}
	}
	return nil
}

func (c AssemblyService) Name() string {
	return "Assembly"
}
