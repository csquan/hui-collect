package part_rebalance

import (
	"fmt"
	"testing"

	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
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

func NoTestRun(t *testing.T) {
	alert.InitDingding(&config.DefaultAlertConf)
	p, err := NewPartReBalanceService(dbtest, &config.Config{})
	if err != nil {
		t.Fatalf("part_rebalance err:%v", err)
	}
	err = p.Run()
	if err != nil {
		t.Fatalf("part run err:%v", err)
	}
}
