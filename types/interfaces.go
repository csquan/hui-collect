package types

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)
	GetOpenedAssetTransferTasks() ([]*AssetTransferTask, error)
	GetOpenedTransactionTask() (*TransactionTask, error)
	GetTxTasks(uint) ([]*TransactionTask, error)
}

type IWriter interface {
	InsertAssetTransfer(task *AssetTransferTask) error
	UpdateAssetTransferTask(task *AssetTransferTask) error
	UpdateTransactionTask(task *TransactionTask) error
	SaveTxTasks([]*TransactionTask) error
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}

