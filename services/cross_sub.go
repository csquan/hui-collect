package services

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type CrossSubTaskService struct {
	db        types.IDB
	bridgeCli *bridge.Bridge
	config    *config.Config
}

func (c *CrossSubTaskService) Run() error {
	parentTasks, err := c.db.GetOpenedCrossTasks()
	if err != nil {
		return err
	}

	for _, pt := range parentTasks {
		childTasks, err := c.db.GetOpenedCrossSubTasks(pt.ID)
		if err != nil {
			continue
		}
		if len(childTasks) == 0 {
			continue
		}
		for _, subTask := range childTasks {
			switch types.CrossSubState(subTask.State) {
			case types.ToCross:
				//do cross
				t := &bridge.Task{
					TaskNo:         subTask.TaskNo,
					FromAccountId:  0,
					ToAccountId:    0,
					FromCurrencyId: 0,
					ToCurrencyId:   0,
					Amount:         subTask.TaskNo,
				}
				taskId, err := c.bridgeCli.AddTask(t)
				if err != nil && taskId != 0 {
					//update bridge taskId
					err1 := c.db.UpdateCrossSubTaskBridgeIDAndState(subTask.ID, taskId, int(types.Crossing))
					if err != nil {
						logrus.Warnf("update cross sub task err:%v,subTaskId:%d", err1, subTask.ID)
					}
				}
			case types.Crossing:
				//watch
				bridgeTask, err := c.bridgeCli.GetTaskDetail(subTask.BridgeTaskId)
				if err != nil {
					continue
				}
				switch bridgeTask.Status {
				case 0:
					logrus.Infof("bridge task wait to start bridgeTaskId:%d", subTask.BridgeTaskId)
				case 1:
					logrus.Infof("bridge task crossing bridgeTaskId:%d", subTask.BridgeTaskId)
				case 2:
					logrus.Infof("bridge task crossed bridgeTaskId:%d", subTask.BridgeTaskId)
					err = c.db.UpdateCrossSubTaskState(subTask.ID, int(types.Crossed))
					if err != nil {
						continue
					}
				default:
					logrus.Fatalf("unexpected bridge task state subtaskId:%d,bridge:%d,state:%d", subTask.ID, subTask.BridgeTaskId, subTask.State)
				}
			default:
				logrus.Fatalf("unexpected sub task state sub_task id:%d,state:%d", subTask.ID, subTask.State)
			}
		}
	}
	return nil
}
