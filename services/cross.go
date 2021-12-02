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

var zeroD = decimal.NewFromFloat(0)

type CrossService struct {
	db        types.IDB
	bridgeCli bridge.IBridge
	config    *config.Config
}

func NewCrossService(db types.IDB, bCli bridge.IBridge, c *config.Config) *CrossService {
	return &CrossService{
		db:        db,
		bridgeCli: bCli,
		config:    c,
	}
}

func (c *CrossService) estimateCrossTask(fromAccountId, toAccountId uint64,
	fromCurrencyId, toCurrencyId int,
	amount string) (total, single, minAmount string, err error) {
	btask := &bridge.Task{
		FromAccountId:  fromAccountId,
		ToAccountId:    toAccountId,
		FromCurrencyId: fromCurrencyId,
		ToCurrencyId:   toCurrencyId,
		Amount:         amount,
	}
	estimateResult, err := c.bridgeCli.EstimateTask(btask)
	if err != nil {
		return "0", "0", "", err
	}
	return estimateResult.TotalQuota, estimateResult.SingleQuota, estimateResult.MinAmount, nil
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

func getBridgeID(bridgeCli bridge.IBridge, task *types.CrossTask) (*bridgeId, error) {
	fromChainId, ok := bridgeCli.GetChainId(task.ChainFrom)
	if !ok {
		return nil, fmt.Errorf("fromChainId not found")
	}
	toChainId, ok := bridgeCli.GetChainId(task.ChainTo)
	if !ok {
		return nil, fmt.Errorf("toChainId not found")
	}
	fromAccountId, ok := bridgeCli.GetAccountId(task.ChainFromAddr, fromChainId)
	if !ok {
		return nil, fmt.Errorf("fromAccountId not found")
	}
	toAccountId, ok := bridgeCli.GetAccountId(task.ChainToAddr, toChainId)
	if !ok {
		return nil, fmt.Errorf("toAccountId not found")
	}
	fromCurrencyId, ok := bridgeCli.GetCurrencyID(task.CurrencyFrom)
	if !ok {
		return nil, fmt.Errorf("fromCurrencyId not found")
	}
	toCurrencyId, ok := bridgeCli.GetCurrencyID(task.CurrencyTo)
	if !ok {
		return nil, fmt.Errorf("toCurrencyId not found")
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
	logrus.Infof("get cross sub tasks size:%d,parent:%d", len(subTasks), parent.ID)
	if len(subTasks) > 0 {
		sort.Slice(subTasks, func(i, j int) bool {
			return subTasks[i].TaskNo < subTasks[j].TaskNo
		})
		latestSub := subTasks[len(subTasks)-1]
		switch types.CrossSubState(latestSub.State) {
		case types.Crossing, types.Crossed:
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
				totalStr, singleStr, minStr, err := c.estimateCrossTask(bridgeId.fromAccountId, bridgeId.toAccountId,
					bridgeId.fromCurrencyId, bridgeId.toCurrencyId, amountLeft.String())
				if err != nil {
					return false, fmt.Errorf("estimate task err:%v,parent:%d", err, parent.ID)
				}
				if singleStr == "" || singleStr == "0" {
					return false, fmt.Errorf("singleQuota 0")
				}
				total := mustStrToDecimal(totalStr)
				single := mustStrToDecimal(singleStr)
				minAmount := mustStrToDecimal(minStr)

				if total.LessThan(amountLeft) {
					logrus.Fatalf("unexpectd esimate total parentId:%d,total:%d,amount:%d", parent.ID, total, amountLeft)
				}

				// amountLeft = decimal.Min(amountLeft, single)
				amountCur, err := getAmountCur(minAmount, single, total, amountLeft)
				if err != nil {
					return false, fmt.Errorf("amount err:%v,min:%s,single:%s,total:%s,amount:%s", err,
						minAmount.String(), single.String(), total.String(), amountLeft.String())
				}
				subTask := &types.CrossSubTask{
					ParentTaskId: parent.ID,
					TaskNo:       latestSub.TaskNo + 1,
					Amount:       amountCur.String(),
					State:        int(types.ToCross),
				}
				err = c.db.SaveCrossSubTask(subTask)
				if err != nil {
					return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
				}
				if amountCur.Equal(amountLeft) { //剩余amount可一次提交完成
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
		totalStr, singleStr, minAmountStr, err := c.estimateCrossTask(bridgeId.fromAccountId, bridgeId.toAccountId,
			bridgeId.fromCurrencyId, bridgeId.toCurrencyId, amount.String())
		if err != nil {
			return false, err
		}
		if singleStr == "" || singleStr == "0" {
			return false, fmt.Errorf("singleQuota 0")
		}
		total := mustStrToDecimal(totalStr)
		single := mustStrToDecimal(singleStr)
		minAmount := mustStrToDecimal(minAmountStr)

		amountCur, err := getAmountCur(minAmount, single, total, amount)
		logrus.Infof("amount cur:%s,err:%v", amountCur.String(), err)
		if err != nil {
			return false, fmt.Errorf("amount err:%v,min:%s,single:%s,total:%s,amount:%s", err,
				minAmount.String(), single.String(), total.String(), amount.String())
		}
		firstTaskNo := parent.ID << 10
		subTask := &types.CrossSubTask{
			ParentTaskId: parent.ID,
			TaskNo:       firstTaskNo,
			Amount:       amountCur.String(),
		}
		err = c.db.SaveCrossSubTask(subTask)
		if err != nil {
			return false, fmt.Errorf("add sub task err:%v,task:%v", err, subTask)
		}
		if amountCur.Equal(amount) {
			return true, nil
		}
	}
	return false, nil
}

func (c *CrossService) transferTaskState(taskId uint64, nextState types.CrossState) error {
	return c.db.UpdateCrossTaskState(taskId, nextState)
}

func (c *CrossService) Run() error {
	tasks, err := c.db.GetOpenedCrossTasks()
	if err != nil {
		return fmt.Errorf("get cross tasks err:%v", err)
	}

	if len(tasks) == 0 {
		logrus.Infof("no cross tasks")
		return nil
	}

	for _, task := range tasks {
		switch task.State {
		case types.ToCreateSubTask:
			ok, err := c.addCrossSubTasks(task)
			if err != nil {
				logrus.Errorf("add subtasks err:%v,task:%v", err, task)
				continue
			}

			if ok {
				err := c.transferTaskState(task.ID, types.SubTaskCreated)
				if err != nil {
					logrus.Errorf("update cross task state err:%v,task:%v", err, task)
				}
			}
		case types.SubTaskCreated:
			subTasks, err := c.db.GetCrossSubTasks(task.ID)
			if err != nil {
				logrus.Errorf("get cross sub tasks error: %v", err)
				continue
			}

			var sucCnt int
			for _, subT := range subTasks {
				if subT.State == int(types.Crossed) {
					sucCnt++
				}
			}

			logrus.Infof("cross task: %v progress:%v/%v", task, sucCnt, len(subTasks))
			if sucCnt == len(subTasks) {
				err = c.transferTaskState(task.ID, types.TaskSuc)
				if err != nil {
					continue
				}
			}
		default:
			return fmt.Errorf("state:[%v] not defined taskId:%d", task.State, task.ID)
		}
	}
	return nil
}

func (c CrossService) Name() string {
	return "cross"
}

func getAmountCur(minAmount, single, total, amount decimal.Decimal) (decimal.Decimal, error) {
	if minAmount.GreaterThan(zeroD) && amount.LessThan(minAmount) { //amount < min
		return zeroD, fmt.Errorf("amount less than minAmount")
	}
	if amount.GreaterThan(total) { //amount > total
		return zeroD, fmt.Errorf("amount greater than total")
	}
	if amount.LessThanOrEqual(single) { // min<= amount <= single
		return amount, nil
	}
	if minAmount.Equal(zeroD) { //无min限制
		return decimal.Min(single, amount), nil
	}
	twiceMin := minAmount.Add(minAmount)
	if amount.LessThan(twiceMin) { // single < amount < 2*minAmount
		return zeroD, fmt.Errorf("amount less than 2*minAmount")
	} else { // single < 2*minAmount< amount
		amountCur := amount
		for amountCur.GreaterThan(single) {
			amountCur = amountCur.Sub(minAmount)
		}
		return amountCur, nil
	}
}
