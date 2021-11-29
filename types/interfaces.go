package types

import (
	"github.com/go-xorm/xorm"
)

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)

	GetTransactionTasksWithReBalanceId(reBalanceId uint64, transactionType TransactionType) ([]*TransactionTask, error)

	GetOpenedTransactionTask() ([]*TransactionTask, error)
	GetApprove(token, spender string) (*ApproveRecord, error)
	//GetOrderID() (int, error)


	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint64) ([]*CrossSubTask, error)
	GetOpenedCrossSubTasks(parentTaskId uint64) ([]*CrossSubTask, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	UpdatePartReBalanceTask(itf xorm.Interface, t *PartReBalanceTask) error
	SaveTxTasks(xorm.Interface, []*TransactionTask) error
	UpdateTransactionTask(itf xorm.Interface, task *TransactionTask) error
	SaveApprove(approve *ApproveRecord) error

	SaveCrossTasks(itf xorm.Interface, tasks []*CrossTask) error
	//update cross task state
	UpdateCrossTaskState(id uint64, state int) error
	//add bridge task id to sub task
	UpdateCrossSubTaskBridgeIDAndState(id, bridgeTaskId uint64, state int) error
	//save cross sub task
	SaveCrossSubTask(subTask *CrossSubTask) error
	//update cross sub task state
	UpdateCrossSubTaskState(id uint64, state int) error

}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
