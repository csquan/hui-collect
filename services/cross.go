package services

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type crossState int

const (
	toCreateSubTask crossState = iota
	suTaskCreated
	taskSuc
	taskFail
	taskPartSuc
)

type CrossService struct {
	db     types.IDB
	config *config.Config
}

func getCrossSubTasks(t *types.CrossTask) []*types.CrossSubTask {
	return nil
}

func (c *CrossService) transferTaskState(taskId uint, nextState crossState) error {
	return nil
}

func (c *CrossService) Run() error {
	tasks, err := c.db.GetOpenedCrossTasks()
	if err != nil {
		return fmt.Errorf("get cross tasks err:%v", err)
	}
	if len(tasks) == 0 {
		logrus.Infof("no cross tasks")
	}

	for _, task := range tasks {
		switch crossState(task.State) {
		case toCreateSubTask:
			subTasks := getCrossSubTasks(task)
			//insert sub tasks
			err := c.db.SaveCrossSubTasks(subTasks)
			if err != nil {
				c.transferTaskState(task.ID, suTaskCreated)
			}
		case suTaskCreated:
			subTasks, err := c.db.GetCrossSubTasks(task.ID)
			if err != nil {
				cnt := len(subTasks)
				var (
					sucCnt  int
					failCnt int
				)
				for _, subTask := range subTasks {
					switch crossSubState(subTask.State) {
					case suc:
						sucCnt++
					case fail:
						failCnt++
					}
				}
				if sucCnt+failCnt == cnt {
					switch failCnt {
					case 0:
						c.transferTaskState(task.ID, taskSuc)
					case cnt:
						c.transferTaskState(task.ID, taskFail)
					default:
						c.transferTaskState(task.ID, taskPartSuc)

					}
				}
			}

		default:
			return fmt.Errorf("state not define taskId:%d", task.ID)
		}
	}
	return nil
}
