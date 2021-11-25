package types

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)
	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint) ([]*TransactionTask, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint) ([]*CrossSubTask, error)
}

type IWriter interface {
	UpdateTransferTask(task *AssetTransferTask) error
	SaveTxTasks([]*TransactionTask) error

	SaveCrossSubTasks([]*CrossSubTask) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
