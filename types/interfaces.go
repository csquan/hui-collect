package types

import (
	"github.com/go-xorm/xorm"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_db.go -package=mock
type IReader interface {
	//查询可以进行交易组装的任务--状态为Init
	GetOpenedAssemblyTasks() ([]*TransactionTask, error)
	//查询可以进行签名的任务--状态为Assembly
	GetOpenedSignTasks() ([]*TransactionTask, error)
	//查询可以进行广播的任务--状态为sign
	GetOpenedBroadcastTasks() ([]*TransactionTask, error)
	//查询可以进行广播的任务--状态为boradcast
	GetOpenedCheckReceiptTasks() ([]*TransactionTask, error)
	//查询可以进行成功回调的任务--状态为checkreceipt
	GetOpenedOkCallBackTasks() ([]*TransactionTask, error)
	//查询可以进行失败回调的任务--状态为checkreceipt
	GetOpenedFailCallBackTasks() ([]*TransactionTask, error)

	GetOpenedCollectTask() ([]*CollectTxDB, error)

	UpdateTransactionTaskState(taskID uint64, state int) error

	//查询指定的task
	GetSpecifyTasks(task *TransactionTask) (*TransactionTask, error)

	//查询非完成状态的task
	GetTaskNonce(from string) (*TransactionTask, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine
	SaveTxTask(itf xorm.Interface, task *TransactionTask) (err error)
	UpdateTransactionTask(itf xorm.Interface, task *TransactionTask) error
	UpdateTransactionTaskMessage(taskID uint64, message string) error
	InsertCollectSubTx(itf xorm.Interface, tasks *TransactionTask) (err error)
	UpdateCollectTx(itf xorm.Interface, task *CollectTxDB) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
