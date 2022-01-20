package db

import (
	"fmt"

	"time"

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
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	engine.SetTZLocation(location)
	engine.SetTZDatabase(location)

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
	_, err := itf.Where("f_id = ?", task.ID).Update(task)
	return err
}

func (m *Mysql) UpdatePartReBalanceTaskMessage(taskID uint64, message string) error {
	_, err := m.engine.Exec("update t_part_rebalance_task set f_message = ? where f_id = ?", message, taskID)
	return err
}

func (*Mysql) UpdateFullReBalanceTask(itf xorm.Interface, task *types.FullReBalanceTask) error {
	_, err := itf.Where("f_id = ?", task.ID).Update(task)
	return err
}
func (m *Mysql) UpdateFullReBalanceTaskMessage(taskID uint64, message string) error {
	_, err := m.engine.Exec("update t_full_rebalance_task set f_message = ? where f_id = ?", message, taskID)
	return err
}

func (m *Mysql) SaveTxTasks(itf xorm.Interface, tasks []*types.TransactionTask) (err error) {
	for _, t := range tasks {
		if t.GasLimit == "" {
			t.GasLimit = "5000000"
		}
		if t.Amount == "" {
			t.Amount = "0"
		}
		if t.Quantity == "" {
			t.Quantity = "0"
		}
	}
	_, err = itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert transaction task error:%v, tasks:%v", err, tasks)
	}

	return
}

func (m *Mysql) SaveRebalanceTask(itf xorm.Interface, tasks *types.PartReBalanceTask) (err error) {
	_, err = itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert transaction task error:%v, tasks:%v", err, tasks)
	}

	return
}

func (m *Mysql) UpdateTransactionTask(itf xorm.Interface, task *types.TransactionTask) error {
	_, err := itf.Table("t_transaction_task").Where("f_id = ?", task.ID).Update(task)
	return err
}
func (m *Mysql) UpdateTransactionTaskMessage(taskID uint64, message string) error {
	_, err := m.engine.Exec("update t_transaction_task set f_message = ? where f_id = ?", message, taskID)
	return err
}

func (m *Mysql) SaveCrossTasks(itf xorm.Interface, tasks []*types.CrossTask) error {
	_, err := itf.Insert(tasks)
	if err != nil {
		logrus.Errorf("insert cross tasks error:%v", err)
	}

	return err
}

func (m *Mysql) UpdateCrossTaskState(itf xorm.Interface, id uint64, state int) error {
	_, err := itf.Table("t_cross_task").Where("f_id = ?", id).Cols("f_state").Update(
		&types.CrossTask{
			State: state,
		},
	)
	return err
}

func (m *Mysql) SaveCrossSubTask(subTask *types.CrossSubTask) error {
	_, err := m.engine.Table("t_cross_sub_task").Insert(subTask)
	return err
}

func (m *Mysql) SaveCrossSubTasks(itf xorm.Interface, subTask []*types.CrossSubTask) error {
	_, err := itf.Table("t_cross_sub_task").Insert(subTask)
	return err
}

func (m *Mysql) UpdateCrossSubTaskState(id uint64, state int) error {
	_, err := m.engine.Table("t_cross_sub_task").Where("f_id = ?", id).Cols("f_state").Update(
		&types.CrossSubTask{
			State: state,
		},
	)
	return err
}

func (m *Mysql) UpdateCrossSubTaskBridgeIDAndState(id, bridgeTaskId uint64, state int) error {
	_, err := m.engine.Table("t_cross_sub_task").Where("f_id = ?", id).
		Cols("f_bridge_task_id", "f_state").Update(
		&types.CrossSubTask{
			BridgeTaskId: bridgeTaskId,
			State:        state,
		})
	return err
}

func (m *Mysql) SaveFullRebalanceTask(itf xorm.Interface, task *types.FullReBalanceTask) error {
	affected, err := itf.Insert(task)
	if err != nil {
		return err
	}
	if affected < 0 {
		return fmt.Errorf("affected 0")
	}
	return nil
}
func (m *Mysql) UpdateTaskSwitch(isRun bool) error {
	_, err := m.engine.Exec(fmt.Sprintf("update t_task_switch set f_is_run = %t", isRun))
	return err
}
