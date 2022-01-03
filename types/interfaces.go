package types

import (
	"github.com/go-xorm/xorm"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_db.go -package=mock
type IReader interface {
	GetPartReBalanceTaskByFullRebalanceID(fullRebalanceID uint64) (task *PartReBalanceTask, err error)
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)
	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)
	GetOpenedFullReBalanceTasks() ([]*FullReBalanceTask, error)
	GetFullRelalanceTask(taskId uint64) (*FullReBalanceTask, error)
	GetTransactionTasksWithReBalanceId(reBalanceId uint64, transactionType TransactionType) ([]*TransactionTask, error)

	GetOpenedTransactionTask() ([]*TransactionTask, error)
	//GetApprove(token, spender string) (*ApproveRecord, error)
	//GetOrderID() (int, error)
	GetTransactionTasksWithFullRebalanceId(fullReBalanceId uint64, transactionType TransactionType) ([]*TransactionTask, error)

	GetTransactionTasksWithPartRebalanceId(partRebalanceId uint64, transactionType TransactionType) ([]*TransactionTask, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint64) ([]*CrossSubTask, error)
	GetOpenedCrossSubTasks(parentTaskId uint64) ([]*CrossSubTask, error)

	GetCurrency() ([]*Currency, error)
	GetTokens() ([]*Token, error)
	GetTaskSwitch() (bool, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine
	SaveRebalanceTask(itf xorm.Interface, tasks *PartReBalanceTask) (err error)
	UpdatePartReBalanceTask(itf xorm.Interface, t *PartReBalanceTask) error
	UpdatePartReBalanceTaskMessage(taskID uint64, message string) error
	UpdateFullReBalanceTask(itf xorm.Interface, task *FullReBalanceTask) error
	UpdateFullReBalanceTaskMessage(taskID uint64, message string) error
	SaveTxTasks(xorm.Interface, []*TransactionTask) error
	UpdateTransactionTask(itf xorm.Interface, task *TransactionTask) error
	UpdateTransactionTaskMessage(taskID uint64, message string) error
	SaveFullRebalanceTask(itf xorm.Interface, task *FullReBalanceTask) error

	SaveCrossTasks(itf xorm.Interface, tasks []*CrossTask) error
	//update cross task state
	UpdateCrossTaskState(itf xorm.Interface, id uint64, state int) error
	//add bridge task id to sub task
	UpdateCrossSubTaskBridgeIDAndState(id, bridgeTaskId uint64, state int) error
	//save cross sub task
	SaveCrossSubTask(subTask *CrossSubTask) error
	SaveCrossSubTasks(itf xorm.Interface, subTask []*CrossSubTask) error
	//update cross sub task state
	UpdateCrossSubTaskState(id uint64, state int) error
	UpdateTaskSwitch(isRun bool) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
