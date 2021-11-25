package types

import (
	"time"
)

type Base struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseTask struct {
	State   int
	Message string
	Params  []byte // used for create sub tasks
}

type PartReBalanceState = int

const (
	Init PartReBalanceState = iota
	Cross
	TransferIn
	Farm
	Success
	Failed
)

type PartReBalanceTask struct {
	*Base
	*BaseTask
	Data []*RebalanceData
}

type RebalanceData struct {
	vaultAddr string //合约地址
	address   string //跨连桥钱包地址
	amount    uint64 //跨链资金大小
	taskId    uint64 //链下跨链任务id
}

type AssetTransferTask struct {
	*Base
	*BaseTask
	RebalanceId  uint   `xorm:"rebalance_id"`
	TransferType uint8  `xorm:"transfer_type"`
	Progress     string `xorm:"progress"`
}

type TransactionTask struct {
	*Base
	*BaseTask
	RebalanceId     uint   `xorm:"rebalance_id"`
	TransferId      uint   `xorm:"transfer_id"`
	Nonce           int    `xorm:"nonce"`
	ChainId         int    `xorm:"chain_id"`
	From            string `xorm:"from"`
	To              string `xorm:"to"`
	ContractAddress string `xorm:"contract_address"`
	Value           int    `xorm:"value"`
	UnSignData      string `xorm:"unsigned_data"`
	SignData        []byte `xorm:"signed_data"`
	Hash            string `xorm:"hash"`
}

type InvestTask struct {
	*Base
	*BaseTask
	RebalanceId uint64 `xorm:"rebalance_id"`
}
