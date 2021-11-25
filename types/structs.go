package types

import "time"

type Base struct {
	ID        uint64    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseTask struct {
	State   int
	Message string
	params  string // used for create sub tasks
}

type PartReBalanceState = int

const (
	PartReBalanceInit PartReBalanceState = iota
	PartReBalanceCross
	PartReBalanceTransferIn
	PartReBalanceInvest
	PartReBalanceSuccess
	PartReBalanceFailed
)

type PartReBalanceTask struct {
	*Base
	*BaseTask
	Params string `xorm:"params"`
}

type AssetTransferTask struct {
	*Base
	*BaseTask
	RebalanceId  uint64 `xorm:"rebalance_id"`
	TransferType uint8  `xorm:"transfer_type"`
	Progress     string `xorm:"progress"`
}

type TransactionTask struct {
	*Base
	*BaseTask
	RebalanceId     uint64 `xorm:"rebalance_id"`
	TransferId      uint64 `xorm:"transfer_id"`
	Nonce           int    `xorm:"nonce"`
	ChainId         int    `xorm:"chain_id"`
	From            string `xorm:"from"`
	To              string `xorm:"to"`
	ContractAddress string `xorm:"contract_address"`
	Value           int    `xorm:"value"`
	UnSignData      string `xorm:"unsigned_data"`
	SignData        string `xorm:"signed_data"`
}

type InvestTask struct {
	*Base
	*BaseTask
	RebalanceId uint64 `xorm:"rebalance_id"`
}

type CrossTask struct {
	*Base
	*BaseTask
	ReBalanceId  uint64 `xorm:"rebalance_id"`
	ChainFrom    string `xorm:"chain_from"`
	ChainTo      string `xorm:"chain_to"`
	CurrencyFrom string `xorm:"currency_from"`
	CurrencyTo   string `xorm:"currency_to"`
	Amount       string `xorm:"amount"`
}

type CrossSubTask struct {
	*Base
	No           uint   `xorm:"no"` //taskNo
	ParentId     uint64 `xorm:"parent_id"`
	ChainFrom    string `xorm:"chain_from"`
	ChainTo      string `xorm:"chain_to"`
	CurrencyFrom string `xorm:"currency_from"`
	CurrencyTo   string `xorm:"currency_to"`
	Amount       string `xorm:"amount"`
	State        int    `xorm:"state"`
}
