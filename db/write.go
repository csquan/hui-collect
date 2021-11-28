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


func (m *Mysql) SaveTxTasks(itf xorm.Interface, tasks []*types.TransactionTask) (err error) {
	_, err = itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert transaction task error:%v, tasks:%v", err, tasks)
	}

	return
}

func (m *Mysql) UpdateTransactionTask(itf xorm.Interface, task *types.TransactionTask) error {
	_, err := m.engine.Table("t_transaction_task").Where("f_id = ?", task.ID).Update(task)
	return err
}

func (m *Mysql) SaveCrossTasks(itf xorm.Interface, tasks []*types.CrossTask) error {
	_, err := itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert cross tasks error:%v", err)
	}

	return err
}

func (m *Mysql) UpdateCrossTaskState(id uint64, state int) error {
	_, err := m.engine.Table("cross_task").Where("id = ?", id).Cols("state").Update(
		&types.CrossTask{
			State: state,
		},
	)
	return err
}

func (m *Mysql) SaveCrossSubTask(subTask *types.CrossSubTask) error {
	_, err := m.engine.Table("cross_sub_task").Insert(subTask)
	return err
}

func (m *Mysql) UpdateCrossSubTaskState(id uint64, state int) error {
	_, err := m.engine.Table("cross_sub_task").Where("id = ?", id).Cols("state").Update(
		&types.CrossSubTask{
			State: state,
		},
	)
	return err
}

func (m *Mysql) UpdateCrossSubTaskBridgeIDAndState(id, bridgeTaskId uint64, state int) error {
	_, err := m.engine.Table("cross_sub_task").Where("id = ?", id).
		Cols("bridge_task_id", "state").Update(
		&types.CrossSubTask{
			BridgeTaskId: bridgeTaskId,
			State:        state,
		})
	return err
}

func (m *Mysql) UpdateOrderID(itf xorm.Interface, id int) error {
	_, err := itf.Update(id)
	return err
}
