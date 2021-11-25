package services

import (
	"fmt"
	"sort"
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

// func (c *CrossService) getSingleCrossAmount(btask *bridge.Task) (uint64, error) {
// 	estimateResult, err := c.bridgeCli.EstimateTask(btask)
// 	if err != nil {
// 		return 0, err
// 	}
// 	single := estimateResult.SingleQuota
// 	return single, nil
// }

func (c *CrossService) estimateCrossTask(addrFrom, addrTo, currencyFrom, currencyTo string, amount uint64) (total, single uint64, err error) {
	fromAccountId := c.bridgeCli.GetAccountId(addrFrom)
	toAccountId := c.bridgeCli.GetAccountId(addrTo)
	fromCurrencyId := c.bridgeCli.GetCurrencyID(currencyFrom)
	toCurrencyId := c.bridgeCli.GetCurrencyID(currencyTo)
	btask := &bridge.Task{
		FromAccountId:  fromAccountId,
		ToAccountId:    toAccountId,
		FromCurrencyId: fromCurrencyId,
		ToCurrencyId:   toCurrencyId,
		Amount:         amount,
	}
	estimateResult, err := c.bridgeCli.EstimateTask(btask)
	if err != nil {
		return 0, 0, err
	}
	return estimateResult.TotalQuota, estimateResult.SingleQuota, nil
}

func (c *CrossService) addCrossSubTasks(t *types.CrossTask) (finished bool, err error) {
	if t.Amount == "" {
		return true, nil
	}
	amount, _ := strconv.ParseUint(t.Amount, 10, 64) //TODO amount type
	if amount == 0 {                                 //create sub task finish
		return true, nil
	}
	subTasks, _ := c.db.GetCrossSubTasks(t.ID)
	if len(subTasks) > 0 {
		sort.Slice(subTasks, func(i, j int) bool {
			return subTasks[i].TaskNo < subTasks[j].TaskNo
		})
		latestSub := subTasks[len(subTasks)-1]
		switch crossSubState(latestSub.State) {
		case crossing:
			fallthrough
		case crossed:
			var totalAmount uint64
			for _, sub := range subTasks {
				a, _ := strconv.ParseUint(sub.Amount, 10, 64)
				totalAmount += a
			}

			if totalAmount < amount {
				amountLeft := amount - totalAmount
				_, single, err := c.estimateCrossTask(t.ChainFromAddr, t.ChainToAddr, t.CurrencyFrom, t.CurrencyTo, amountLeft)
				if err != nil {
					return false, err
				}
				amountLeft = min(amountLeft, single)
				subTask := &types.CrossSubTask{
					ParentTaskId: t.ID,
					TaskNo:       t.TaskNo,
					ChainFrom:    t.ChainFrom,
					ChainTo:      t.ChainTo,
					CurrencyFrom: t.CurrencyFrom,
					CurrencyTo:   t.CurrencyTo,
					Amount:       fmt.Sprintf("%d", amountLeft),
				}
				err = c.db.SaveCrossSubTask(subTask)
				if err != nil {
					return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
				}
				if amountLeft <= single { //剩余amount可一次提交完成
					// c.db.UpdateCrossTaskState(t.ID, int(subTaskCreated))
					return true, nil
				}
			} else if totalAmount == amount {
				// c.db.UpdateCrossTaskState(t.ID, int(subTaskCreated))
				return true, nil
			} else {
				logrus.Fatalf("unexpected amount taskID:%d,task:%v", t.ID, t)
			}
		case toCross:
			return false, nil
		}
	} else { //
		total, single, err := c.estimateCrossTask(t.ChainFromAddr, t.ChainToAddr, t.CurrencyFrom, t.CurrencyTo, amount)
		if err != nil {
			return false, err
		}
		if amount <= total {
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
			err = c.db.SaveCrossSubTask(subTask)
			if err != nil {
				return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
			}
			if amount <= single {
				// c.db.UpdateCrossTaskState(t.ID, int(subTaskCreated))
				return true, nil
			}
		}
	}
	return false, nil
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
			ok, err := c.addCrossSubTasks(task)
			if err != nil {
				logrus.Errorf("add subtasks err:v,task:%v", err, task)
				continue
			} else if ok {
				err := c.transferTaskState(task.ID, subTaskCreated)
				if err != nil {
					logrus.Errorf("update cross task state err:%v,task:%v", err, task)
				}
			}
		case subTaskCreated:
			subTasks, err := c.db.GetOpenedCrossSubTasks(task.ID)
			if err != nil {
				continue
			} else {
				var sucCnt int
				for _, subT := range subTasks {
					if subT.State == int(crossed) {
						sucCnt++
					}
				}
				if sucCnt == len(subTasks) {
					err = c.transferTaskState(task.ID, taskSuc)
					if err != nil {
						continue
					}
				}
			}

		default:
			return fmt.Errorf("state not define taskId:%d", task.ID)
		}
	}
	return nil
}
