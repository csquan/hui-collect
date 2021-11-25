package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"xorm.io/core"
)

type Mysql struct {
	conf   *config.DataBaseConf
	engine *xorm.Engine
}

func NewMysql(conf *config.DataBaseConf) (m *Mysql, err error) {
	//"test:123@/test?charset=utf8"
	engine, err := xorm.NewEngine("mysql", conf.DB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	engine.ShowSQL(false)
	engine.Logger().SetLevel(core.LOG_DEBUG)

	m = &Mysql{
		conf:   conf,
		engine: engine,
	}

	return
}

func (*Mysql) InsertAssetTransfer(task *types.AssetTransferTask) error {
	return nil
}
func (*Mysql) UpdateAssetTransferTask(task *types.AssetTransferTask) error {
	return nil
}
func (*Mysql) SaveTxTasks([]*types.TransactionTask) error {
	return nil
}

func (*Mysql) UpdateTransactionTask(task *types.TransactionTask) error {
	return nil
}
