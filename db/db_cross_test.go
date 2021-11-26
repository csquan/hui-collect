package db

import (
	"fmt"
	"testing"

	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

var dbtest *Mysql

func init() {
	var err error
	dbtest, err = NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		panic(fmt.Sprintf("c mysql cli err:%v", err))
	}
}
func TestAddCrossTask(t *testing.T) {
	s := dbtest.GetSession()
	err := dbtest.SaveCrossTasks(s, []*types.CrossTask{
		&types.CrossTask{
			RebalanceId:   3,
			ChainFrom:     "HECO",
			ChainFromAddr: "0xee",
			ChainTo:       "BSC",
			ChainToAddr:   "0xff",
			CurrencyFrom:  "BTC",
			CurrencyTo:    "BTC",
			Amount:        "1",
			State:         0,
		},
	})
	if err != nil {
		t.Fatalf("save cross task err:%v", err)
	}
}

func TestQueryCrossTask(t *testing.T) {
	tasks, err := dbtest.GetOpenedCrossTasks()
	if err != nil {
		t.Fatalf("get err:%v", err)
	}
	t.Logf("tasks:%v", tasks[0].ID)
}

func TestAddSubCrossTask(t *testing.T) {
	dbtest.SaveCrossSubTask(&types.CrossSubTask{
		TaskNo:       0,
		ParentTaskId: 3,
	})
}
