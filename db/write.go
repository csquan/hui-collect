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

func (*Mysql) InsertAssetTransfer(itf xorm.Interface, task *types.AssetTransferTask) error {
	return nil
}
func (*Mysql) UpdateAssetTransferTask(task *types.AssetTransferTask) error {
	return nil
}

func (m *Mysql) GetEngine() *xorm.Engine {
	return m.engine
}

func (m *Mysql) GetSession() *xorm.Session {
	return m.engine.NewSession()
}

func (*Mysql) UpdatePartReBalanceTask(itf xorm.Interface, task *types.PartReBalanceTask) error {
	_, err := itf.ID(task.ID).Update(task)
	return err
}

func (*Mysql) UpdateTransferTask(task *types.AssetTransferTask) error {
	return nil
}
func (m *Mysql) SaveTxTasks([]*types.TransactionTask) error {
	return nil
}
func (m *Mysql) CreateAssetTransferTask(itf xorm.Interface, task *types.AssetTransferTask) (err error) {
	_, err = itf.InsertOne(task)
	if err != nil {
		logrus.Errorf("insert transfer task error:%v", err)
	}

	return
}

func (m *Mysql)UpdateTxTask(task *types.SignTask) error {
	return nil
}

func (m *Mysql) SaveCrossTasks(itf xorm.Interface, tasks []*types.CrossTask) error {
	_, err := itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert cross tasks error:%v", err)
	}

	return err
}

func (m *Mysql) SaveCrossSubTasks([]*types.CrossSubTask) error {
	return nil
}

func (*Mysql) UpdateTransactionTask(task *types.TransactionTask) error {
	return nil
}

func (m *Mysql) UpdateCrossTaskState(id uint64, state int) error {
	return nil
}

func (m *Mysql) UpdateCrossTaskNoAndAmount(itf xorm.Interface, id, taskNo, amount uint64) error {
	return nil
}
func (m *Mysql) UpdateCrossSubTaskBridgeID(itf xorm.Interface, id, bridgeTaskId uint64) error {
	return nil
}
func (m *Mysql) SaveCrossSubTask(subTask *types.CrossSubTask) error {
	return nil
}

func (m *Mysql) UpdateCrossSubTaskState(id uint64, state int) error {
	return nil
}
