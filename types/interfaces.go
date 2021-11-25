package types

import "github.com/go-xorm/xorm"

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)
	UpdatePartReBalanceTask(t *PartReBalanceTask) error

	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetAssetTransferTasksWithReBalanceId(reBalanceId uint64) ([]*AssetTransferTask, error)

	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint64) ([]*TransactionTask, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossTasksByReBalanceId(reBalanceId uint64)([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint64) ([]*CrossSubTask, error)
}

type IWriter interface {

	CreateAssetTransferTask(task *AssetTransferTask) error
	UpdateTransferTask(task *AssetTransferTask) error
	SaveTxTasks([]*TransactionTask) error

	SaveCrossTasks([]*CrossTask) error
	SaveCrossSubTasks([]*CrossSubTask) error

	GetSession() *xorm.Session
	GetEngine() *xorm.Engine
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
