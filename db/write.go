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

func (*Mysql) UpdateTransferTask(task *types.AssetTransferTask) error {
	return nil
}
func (m *Mysql) SaveTxTasks([]*types.TransactionTask) error {
	return nil
}
func (m *Mysql) CreateAssetTransferTask(task *types.AssetTransferTask) (err error) {
	_, err = m.engine.InsertOne(task)
	if err != nil {
		logrus.Errorf("isnert transfer task error:%v", err)
	}

	return
}
