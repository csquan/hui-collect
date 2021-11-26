package types

import "github.com/go-xorm/xorm"

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)

	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetAssetTransferTasksWithReBalanceId(reBalanceId uint64, transferType int) ([]*AssetTransferTask, error)

	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint64) ([]*TransactionTask, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)

	GetCrossTasksByReBalanceId(reBalanceId uint64) ([]*CrossTask, error)

	GetCrossSubTasks(crossTaskId uint64) ([]*CrossSubTask, error)
	GetOpenedCrossSubTasks(uint64) ([]*CrossSubTask, error)

}

type IWriter interface {
	InsertAssetTransfer(itf xorm.Interface, task *AssetTransferTask) error
	UpdateAssetTransferTask(task *AssetTransferTask) error
	UpdateTransactionTask(task *TransactionTask) error

	UpdatePartReBalanceTask(itf xorm.Interface, t *PartReBalanceTask) error

	CreateAssetTransferTask(itf xorm.Interface, task *AssetTransferTask) error
	UpdateTransferTask(task *AssetTransferTask) error

	UpdateTxTask(task *TransactionTask) error
	SaveTxTasks([]*TransactionTask) error

	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	SaveCrossTasks(itf xorm.Interface, tasks []*CrossTask) error
	//update cross task state
	UpdateCrossTaskState(id uint64, state int) error
	//update cross task task_no cur and amount cur
	UpdateCrossTaskNoAndAmount(itf xorm.Interface, id, taskNo, amount uint64) error
	//add bridge task id to sub task
	UpdateCrossSubTaskBridgeID(itf xorm.Interface, id, bridgeTaskId uint64) error

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
