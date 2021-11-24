package types

type IReader interface {
	//GetPartReBalanceTasks(state types.PartReBalanceState) ([]*types.PartReBalanceTask, error)

	GetOpenedPartReBalanceTasks() ([]*PartReBalanceTask, error)
}

type IWriter interface {
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}

