package services

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type crossState = int

const (
	toCreateSubTask crossState = iota
	subTaskCreated
	taskSuc    //all sub task suc
	subTaskSuc //single sub task suc
)

type CrossService struct {
	db        types.IDB
	bridgeCli *bridge.Bridge
	config    *config.Config
}

func min(a, b uint64) uint64 {
	if a <= b {
		return a
	}
	return a
}

func (c *CrossService) getSingleCrossAmount(btask *bridge.Task) (uint64, error) {
	estimateResult, err := c.bridgeCli.EstimateTask(btask)
	if err != nil {
		return 0, err
	}
	single := estimateResult.SingleQuota
	return single, nil
}

func (c *CrossService) addCrossSubTasks(t *types.CrossTask) error {
	fromAccountId := c.bridgeCli.GetAccountId(t.ChainFromAddr)
	toAccountId := c.bridgeCli.GetAccountId(t.ChainToAddr)
	fromCurrencyId := c.bridgeCli.GetCurrencyID(t.CurrencyFrom)
	toCurrencyId := c.bridgeCli.GetCurrencyID(t.CurrencyTo)
	amount, _ := strconv.ParseUint(t.Amount, 10, 64) //TODO
	btask := &bridge.Task{
		TaskNo:         t.TaskNo,
		FromAccountId:  fromAccountId,
		ToAccountId:    toAccountId,
		FromCurrencyId: fromCurrencyId,
		ToCurrencyId:   toCurrencyId,
		Amount:         amount,
	}
	estimateResult, err := c.bridgeCli.EstimateTask(btask)
	if err != nil {
		return err
	}
	total := estimateResult.TotalQuota
	single := estimateResult.SingleQuota
	var taskNo = t.TaskNo
	if amount <= total {
		for {
			amountCur := min(amount, single)
			subTask := &types.CrossSubTask{
				ParentTaskId: t.ID,
				TaskNo:       t.TaskNo,
				ChainFrom:    t.ChainFrom,
				ChainTo:      t.ChainTo,
				CurrencyFrom: t.CurrencyFrom,
				CurrencyTo:   t.CurrencyTo,
				Amount:       fmt.Sprintf("%d", amountCur),
			}
			err := c.db.SaveCrossSubTask(subTask)
			if err != nil {
				return fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
			}

			//update bridge task amount
			btask.Amount = amountCur
			bridgeTaskId, err := c.bridgeCli.AddTask(btask)
			if err != nil {
				return fmt.Errorf("add bridge task err:%v,task:%v", err, btask)
			}

			//transaction start TODO
			s := c.db.GetSession()

			err = s.Begin()
			if err != nil {

			}

			err = c.db.UpdateCrossSubTaskBridgeID(s, subTask.ID, bridgeTaskId)
			if err != nil {
				s.Rollback()
				return err
			}
			amount -= amountCur
			taskNo++
			err = c.db.UpdateCrossTaskNoAndAmount(s, t.ID, taskNo, amount)
			if err != nil {
				s.Rollback()
				return fmt.Errorf("update taskNo err:%v,taskId:%d", err, t.ID)
			}
			err = s.Commit()
			if err != nil {
				s.Rollback()
			}
			//transaction end
			if amount > 0 {
				btask.Amount = amount
				single, err = c.getSingleCrossAmount(btask)
				if err != nil {
					return fmt.Errorf("estimate task err:%v,task:%v", err, btask)
				}
			} else {
				break
			}
		}

	} else {
		logrus.Warnf("task amount bigger than total task:%v,total:%d", t, total)
	}
	return nil
}

func (c *CrossService) transferTaskState(taskId uint64, nextState crossState) error {
	return c.db.UpdateCrossTaskState(taskId, int(nextState))
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
			err := c.addCrossSubTasks(task)
			if err != nil {
				logrus.Errorf("add subtasks err:v,task:%v", err, task)
				continue
			} else {
				err := c.transferTaskState(task.ID, subTaskCreated)
				if err != nil {
					logrus.Errorf("update cross task state err:%v,task:%v", err, task)
				}
			}
		case subTaskCreated:
			subTasks, err := c.db.GetOpenedCrossSubTasks(task.ID)
			if err != nil {

			} else {
				if len(subTasks) == 0 {
					err = c.transferTaskState(task.ID, taskSuc)
					if err != nil {
						logrus.Errorf("update cross task state err:%v,task:%v", err, task)
						continue
					}
				}
				for _, subTask := range subTasks {
					result, err := c.bridgeCli.GetTaskDetail(subTask.BridgeTaskId)
					if err != nil {
						logrus.Errorf("get bridge task detail err:%v,result:%s,taskId:%d", err, result, subTask.BridgeTaskId)
						continue
					}
					switch result.Status {
					case 0:
						logrus.Infof("bridge cross not start taskId:%d", subTask.BridgeTaskId)
					case 1:
						logrus.Infof("bridge crossing taskId:%d", subTask.BridgeTaskId)
					case 2:
						err = c.db.UpdateCrossSubTaskState(task.ID, int(subTaskSuc))
						if err != nil {
							logrus.Errorf("update cross sub task state err:%v,taskID:%d", err, task.ID)
						}
					default:
						logrus.Fatalf("unexpected bridge task status:%d,bridgeTaskId:%d", result.Status, subTask.BridgeTaskId)
					}
				}
			}

		default:
			return fmt.Errorf("state not define taskId:%d", task.ID)
		}
	}
	return nil
}
