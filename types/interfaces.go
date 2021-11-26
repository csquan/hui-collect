package types

import "github.com/go-xorm/xorm"

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)

	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetAssetTransferTasksWithReBalanceId(reBalanceId uint64, transferType int) ([]*AssetTransferTask, error)

	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint64) ([]*TransactionTask, error)

	GetOrderID() (int, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint64) ([]*CrossSubTask, error)
	GetOpenedCrossSubTasks(parentTaskId uint64) ([]*CrossSubTask, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	UpdatePartReBalanceTask(itf xorm.Interface, t *PartReBalanceTask) error

	SaveAssetTransferTask(itf xorm.Interface, task *AssetTransferTask) error

	InsertAssetTransfer(itf xorm.Interface, task *AssetTransferTask) error
	UpdateAssetTransferTask(task *AssetTransferTask) error

	UpdateTransactionTask(task *TransactionTask) error
	UpdateTxTask(itf xorm.Interface, task *TransactionTask) error
	SaveTxTasks([]*TransactionTask) error

	SaveCrossTasks(itf xorm.Interface, tasks []*CrossTask) error
	//update cross task state
	UpdateCrossTaskState(id uint64, state int) error

	//add bridge task id to sub task
	UpdateCrossSubTaskBridgeIDAndState(id, bridgeTaskId uint64, state int) error

	//save cross sub task
	SaveCrossSubTask(subTask *CrossSubTask) error

	//update cross sub task state
	UpdateCrossSubTaskState(id uint64, state int) error

	UpdateOrderID(itf xorm.Interface, id int) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
