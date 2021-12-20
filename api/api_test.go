package api

import (
	"fmt"
	"testing"
	"time"

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
func TestAPIRun(t *testing.T) {
	go Run(config.APIConf{
		Port: 8080,
	}, dbtest)
	t.Logf("http server start")
	time.Sleep(600 * time.Second)
}
