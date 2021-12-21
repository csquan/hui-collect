package services

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/bridge/mock"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/types"
)

var dbtest *db.Mysql

func init() {
	var err error
	dbtest, err = db.NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		panic(fmt.Sprintf("c mysql cli err:%v", err))
	}
}
func NoTestCrossRun(t *testing.T) {
	log.Init("cross_test", config.Log{
		Stdout: config.DefaultLogConfig.Stdout,
	}, "dev")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bridgeCli := mock.NewMockIBridge(ctrl)
	var taskId1 uint64 = 1
	var taskId2 uint64 = 2
	bridgeCli.EXPECT().AddTask(gomock.Any()).Return(taskId1, nil).AnyTimes()
	bridgeCli.EXPECT().AddTask(gomock.Any()).Return(taskId2, nil).AnyTimes()

	//id
	bridgeCli.EXPECT().GetChainId(gomock.Any()).Return(1, true).AnyTimes()
	bridgeCli.EXPECT().GetCurrencyID(gomock.Any()).Return(10, true).AnyTimes()

	bridgeCli.EXPECT().GetAccountId(gomock.Any(), gomock.Any()).Return(uint64(100), true).AnyTimes()

	// estimate
	bridgeCli.EXPECT().EstimateTask(gomock.Any()).Return(&bridge.EstimateTaskResult{
		TotalQuota:  "1000",
		SingleQuota: "50",
		MinAmount:   "3",
	}, nil).AnyTimes()
	bridgeCli.EXPECT().EstimateTask(gomock.Any()).Return(&bridge.EstimateTaskResult{
		TotalQuota:  "1000",
		SingleQuota: "50",
		MinAmount:   "3",
	}, nil).AnyTimes()
	// query task status
	bridgeCli.EXPECT().GetTaskDetail(taskId1).Return(&bridge.TaskDetailResult{
		TaskId: taskId1,
		Status: 2,
	}, nil).AnyTimes()
	bridgeCli.EXPECT().GetTaskDetail(taskId2).Return(&bridge.TaskDetailResult{
		TaskId: taskId2,
		Status: 2,
	}, nil).AnyTimes()

	c := NewCrossService(dbtest, bridgeCli, nil)

	sub := NewCrossSubTaskService(dbtest, bridgeCli, nil)
	services := []types.IAsyncService{
		c,
		sub,
	}
	for {
		for _, s := range services {
			err := s.Run()
			if err != nil {
				t.Fatalf("service run err:%v,name:%s", err, s.Name())
			}
			time.Sleep(time.Second)
		}
	}
}

func TestGetAmounts(t *testing.T) {
	min := decimal.NewFromFloat(1)
	max := decimal.NewFromFloat(10)
	remain := decimal.NewFromFloat(11)
	amount := decimal.NewFromFloat(11)
	tests := []struct {
		min     decimal.Decimal
		max     decimal.Decimal
		remain  decimal.Decimal
		amount  decimal.Decimal
		amounts []decimal.Decimal
		errStr  string
	}{

		{min, max, remain, amount, []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(1)}, ""},                                                         //  max< amount &&amount > 2*min
		{decimal.NewFromFloat(9), decimal.NewFromFloat(10), decimal.NewFromFloat(11), decimal.NewFromFloat(11), []decimal.Decimal{}, "amount less than 2*minAmount"}, // max<amount<2*min
		{decimal.NewFromFloat(1), decimal.NewFromFloat(10), decimal.NewFromFloat(10), decimal.NewFromFloat(10), []decimal.Decimal{decimal.NewFromFloat(10)}, ""},     //  min <=amount <=max
		{decimal.NewFromFloat(1), decimal.NewFromFloat(10), decimal.NewFromFloat(10), decimal.NewFromFloat(0.1), []decimal.Decimal{}, "amount less than min"},        // amount< min
	}
	for i, input := range tests {
		amounts, err := getAmounts(input.min, input.max, input.remain, input.amount)

		if input.errStr != "" {
			if !reflect.DeepEqual(err.Error(), input.errStr) {
				t.Errorf("err not equal index:%d", i)
			}
		} else {
			size1 := len(amounts)
			size2 := len(input.amounts)
			if size1 != size2 {
				t.Errorf("size not equal index:%d", i)
			}
			for i := 0; i < size1; i++ {
				if !amounts[i].Equal(input.amounts[i]) {
					t.Errorf("amout not equal index:%d", i)
				}
			}
		}
	}
}

func TestAddCrossSubTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bridgeCli := mock.NewMockIBridge(ctrl)
	bridgeCli.EXPECT().EstimateTask(gomock.Any()).Return(&bridge.EstimateTaskResult{
		MinAmount:    "1",
		MaxAmount:    "10",
		RemainAmount: "100",
	}, nil).AnyTimes()

	//id
	bridgeCli.EXPECT().GetChainId(gomock.Any()).Return(1, true).AnyTimes()
	bridgeCli.EXPECT().GetCurrencyID(gomock.Any()).Return(10, true).AnyTimes()

	bridgeCli.EXPECT().GetAccountId(gomock.Any(), gomock.Any()).Return(uint64(100), true).AnyTimes()
	c := NewCrossService(dbtest, bridgeCli, nil)

	crossTask, err := dbtest.GetOpenedCrossTasks()
	if err != nil {
		t.Fatalf("get opened cross tasks err:%v", err)
	}
	if len(crossTask) == 0 {
		t.Logf("cross task not found")
		return
	}
	ok, err := c.addCrossSubTasksV2(crossTask[0])
	if !ok || err != nil {
		t.Errorf("add cross sub tasks err:%v,ok:%v", err, ok)
	}
}

func TestCrossMsg(t *testing.T) {
	c, err := createCrossMesg("created", &types.CrossTask{
		ChainFrom:     "heco",
		ChainTo:       "bsc",
		ChainFromAddr: "addr_from",
		ChainToAddr:   "addr_to",
		CurrencyFrom:  "c_from",
		CurrencyTo:    "c_to",
		Amount:        "10",
	}, []*types.CrossSubTask{
		&types.CrossSubTask{
			BridgeTaskId: 1,
			TaskNo:       1024,
			Amount:       "5",
			State:        0,
		},
		&types.CrossSubTask{
			BridgeTaskId: 2,
			TaskNo:       1024,
			Amount:       "5",
			State:        0,
		},
	})
	if err != nil {
		t.Fatalf("cross msg err:%v", err)
	}
	t.Logf("cross msg:%s", c)
}

func TestCrossSubMsg(t *testing.T) {
	c, err := createCrossSubMsg("crosing", &CrossSubInfo{
		Parent: &types.CrossTask{
			ChainFrom:     "heco",
			ChainTo:       "bsc",
			ChainFromAddr: "addr_from",
			ChainToAddr:   "addr_to",
			CurrencyFrom:  "c_from",
			CurrencyTo:    "c_to",
			Amount:        "10",
		},
		Sub: &types.CrossSubTask{
			BridgeTaskId: 0,
			TaskNo:       1024,
			Amount:       "5",
			State:        0,
		},
		FromAccountId:  1,
		ToAccountId:    2,
		FromCurrencyId: 11,
		ToCurrencyId:   22,
	})
	if err != nil {
		t.Fatalf("create cross sub msg err:%v", err)
	}
	t.Logf("cross sub msg:%s", c)
}
