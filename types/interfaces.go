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
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine
	UpdateTransactionTask(itf xorm.Interface, task *TransactionTask) error
	UpdateTransactionTaskMessage(taskID uint64, message string) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
