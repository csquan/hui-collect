package types

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)

	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetAssetTransferTasksWithReBalanceId(reBalanceId uint64) ([]*AssetTransferTask, error)

	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint) ([]*TransactionTask, error)

	GetOpenedCrossTasks() ([]*CrossTask, error)
	GetCrossSubTasks(crossTaskId uint) ([]*CrossSubTask, error)
}

type IWriter interface {
	CreateAssetTransferTask(task *AssetTransferTask) error
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
