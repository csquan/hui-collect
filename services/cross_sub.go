package services

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type crossSubState int

const (
	toCross crossSubState = iota
	crossing
	crossed
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
		subTasks, err := c.db.GetOpenedCrossSubTasks(pt.ID)
		if err != nil {
			continue
		}
		if len(subTasks) == 0 {
			continue
		}
		for _, subT := range subTasks {
			switch crossSubState(subT.State) {
			case toCross:
				//do cross
				t := &bridge.Task{}
				taskId, err := c.bridgeCli.AddTask(t)
				if err != nil {
					s := c.db.GetSession()
					//update bridge taskId
					err = c.db.UpdateCrossSubTaskBridgeIDAndState(s, subT.ID, taskId, int(crossing))
					if err != nil {
						s.Rollback()
						continue
					}
					err = c.db.UpdateCrossTaskNo(s, pt.ID, pt.TaskNo+1)
					if err != nil {
						s.Rollback()
						continue
					}
					s.Commit()
				}
			case crossing:
				//watch
				bridgeTask, err := c.bridgeCli.GetTaskDetail(subT.BridgeTaskId)
				if err != nil {
					continue
				}
				switch bridgeTask.Status {
				case 0:
					logrus.Infof("bridge task wait to start bridgeTaskId:%d", subT.BridgeTaskId)
				case 1:
					logrus.Infof("bridge task crossing bridgeTaskId:%d", subT.BridgeTaskId)
				case 2:
					logrus.Infof("bridge task crossed bridgeTaskId:%d", subT.BridgeTaskId)
					err = c.db.UpdateCrossSubTaskState(subT.ID, int(crossed))
					if err != nil {
						continue
					}
				}
			}
		}
	}
	return nil
}
