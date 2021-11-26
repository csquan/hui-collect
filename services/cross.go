package services

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
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

func (c *CrossService) estimateCrossTask(fromAccountId, toAccountId uint64, fromCurrencyId, toCurrencyId int, amount string) (total, single string, err error) {
	btask := &bridge.Task{
		FromAccountId:  fromAccountId,
		ToAccountId:    toAccountId,
		FromCurrencyId: fromCurrencyId,
		ToCurrencyId:   toCurrencyId,
		Amount:         amount,
	}
	estimateResult, err := c.bridgeCli.EstimateTask(btask)
	if err != nil {
		return "0", "0", err
	}
	return estimateResult.TotalQuota, estimateResult.SingleQuota, nil
}

func mustStrToDecimal(num string) decimal.Decimal {
	d, err := decimal.NewFromString(num)
	if err != nil {
		logrus.Fatalf("to decimal err:%v,num:%s", err, num)
	}
	return d
}

type bridgeId struct {
	fromChainId    int
	toChainId      int
	fromAccountId  uint64
	toAccountId    uint64
	fromCurrencyId int
	toCurrencyId   int
}

func getBridgeID(bridgeCli *bridge.Bridge, task *types.CrossTask) (*bridgeId, error) {
	fromChainId, ok := bridgeCli.GetChainId(task.ChainFrom)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	toChainId, ok := bridgeCli.GetChainId(task.ChainTo)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	fromAccountId, ok := bridgeCli.GetAccountId(task.ChainFromAddr, fromChainId)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	toAccountId, ok := bridgeCli.GetAccountId(task.ChainToAddr, toChainId)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	fromCurrencyId, ok := bridgeCli.GetCurrencyID(task.CurrencyFrom)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	toCurrencyId, ok := bridgeCli.GetCurrencyID(task.CurrencyTo)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	return &bridgeId{
		fromChainId:    fromChainId,
		toChainId:      toChainId,
		fromAccountId:  fromAccountId,
		toAccountId:    toAccountId,
		fromCurrencyId: fromCurrencyId,
		toCurrencyId:   toCurrencyId,
	}, nil
}

func (c *CrossService) addCrossSubTasks(parent *types.CrossTask) (finished bool, err error) {
	if parent.Amount == "" {
		return true, nil
	}
	if parent.Amount == "0" { //create sub task finish
		return true, nil
	}
	amount, err := decimal.NewFromString(parent.Amount)
	if err != nil {
		return false, err
	}
	bridgeId, err := getBridgeID(c.bridgeCli, parent)
	if err != nil {
		return false, err
	}

	subTasks, _ := c.db.GetCrossSubTasks(parent.ID)

	if len(subTasks) > 0 {
		sort.Slice(subTasks, func(i, j int) bool {
			return subTasks[i].TaskNo < subTasks[j].TaskNo
		})
		latestSub := subTasks[len(subTasks)-1]
		switch types.CrossSubState(latestSub.State) {
		case types.Crossing:
			fallthrough
		case types.Crossed:
			var totalAmount decimal.Decimal
			for _, sub := range subTasks {
				subAmount, err := decimal.NewFromString(sub.Amount)
				if err != nil {
					logrus.Fatalf("unexpectd sub amount err:%v,subTaskId:%d", err, sub.ID)
				}
				totalAmount = decimal.Sum(totalAmount, subAmount)
			}

			// if totalAmount < amount {
			if totalAmount.LessThan(amount) {
				amountLeft := amount.Sub(totalAmount)
				totalStr, singleStr, err := c.estimateCrossTask(bridgeId.fromAccountId, bridgeId.toAccountId,
					bridgeId.fromCurrencyId, bridgeId.toCurrencyId, amountLeft.String())
				total := mustStrToDecimal(totalStr)
				single := mustStrToDecimal(singleStr)
				if true {
					logrus.Fatalf("unexpectd esimate total parentId:%d,total:%d,amount:%d", parent.ID, total, amountLeft)
				}
				if err != nil {
					return false, err
				}
				amountLeft = decimal.Min(amountLeft, single)
				subTask := &types.CrossSubTask{
					ParentTaskId: parent.ID,
					TaskNo:       latestSub.TaskNo + 1,
					Amount:       fmt.Sprintf("%d", amountLeft),
					State:        int(types.ToCross),
				}
				err = c.db.SaveCrossSubTask(subTask)
				if err != nil {
					return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
				}
				if amountLeft.LessThanOrEqual(single) { //剩余amount可一次提交完成
					return true, nil
				}
			} else if totalAmount == amount {
				return true, nil
			} else {
				logrus.Fatalf("unexpected amount taskID:%d,task:%v", parent.ID, parent)
			}
		case types.ToCross:
			return false, nil
		default:
			logrus.Fatalf("unexpected task state:%d,sub_task id:%d", latestSub.State, latestSub.ID)
		}
	} else { // the first sub task
		totalStr, singleStr, err := c.estimateCrossTask(bridgeId.fromAccountId, bridgeId.toAccountId,
			bridgeId.fromCurrencyId, bridgeId.toCurrencyId, amount.String())
		if err != nil {
			return false, err
		}
		total := mustStrToDecimal(totalStr)
		signal := mustStrToDecimal(singleStr)
		if amount.LessThanOrEqual(total) {
			amountCur := decimal.Min(amount, signal)
			subTask := &types.CrossSubTask{
				ParentTaskId: parent.ID,
				TaskNo:       0,
				Amount:       amountCur.String(),
			}
			err = c.db.SaveCrossSubTask(subTask)
			if err != nil {
				return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
			}
			if amount.LessThanOrEqual(signal) {
				return true, nil
			}
		} else {
			//logrus.Warnf("cross task amount bigger than total taskId:%d,amount:%d,total:%s", parent.ID, parent.Amount) //TODO
		}
	}
	return false, nil
}

func (c *CrossService) transferTaskState(taskId uint64, nextState types.CrossState) error {
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
		switch types.CrossState(task.State) {
		case types.ToCreateSubTask:
			ok, err := c.addCrossSubTasks(task)
			if err != nil {
				logrus.Errorf("add subtasks err:v%,task:%v", err, task)
				continue
			} else if ok {
				err := c.transferTaskState(task.ID, types.SubTaskCreated)
				if err != nil {
					logrus.Errorf("update cross task state err:%v,task:%v", err, task)
				}
			}
		case types.SubTaskCreated:
			subTasks, err := c.db.GetOpenedCrossSubTasks(task.ID)
			if err != nil {
				continue
			} else {
				var sucCnt int
				for _, subT := range subTasks {
					if subT.State == int(types.Crossed) {
						sucCnt++
					}
				}
				if sucCnt == len(subTasks) {
					err = c.transferTaskState(task.ID, types.TaskSuc)
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
