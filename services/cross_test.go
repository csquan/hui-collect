package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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
func TestCrossRun(t *testing.T) {
	log.Init("cross_test", config.Log{
		Stdout: config.DefaultLogConfig.Stdout,
	})
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
