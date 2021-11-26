package types

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type Base struct {
	ID        uint64    `xorm:"f_id" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type BaseTask struct {
	State   int    `xorm:"f_state"`
	Message string `xorm:"f_message"`
}

type PartReBalanceState = int

type CrossState = int
type CrossSubState int

type AssetTransferState = int

const (
	PartReBalanceInit PartReBalanceState = iota
	PartReBalanceCross
	PartReBalanceTransferIn
	PartReBalanceInvest
	PartReBalanceSuccess
	PartReBalanceFailed

	ToCreateSubTask CrossState = iota
	SubTaskCreated
	TaskSuc               //all sub task suc
	ToCross CrossSubState = iota
	Crossing
	Crossed

	AssetTransferIn = iota
	Invest

	AssetTransferInit AssetTransferState = iota
	AssetTransferOngoing
	AssetTransferSuccess
	AssetTransferFailed
)

type PartReBalanceTask struct {
	*Base     `xorm:"extends"`
	*BaseTask `xorm:"extends"`
	Params    string `xorm:"f_params"`
}

func (p *PartReBalanceTask) TableName() string {
	return "t_part_rebalance_task"
}

func (p *PartReBalanceTask) ReadParams() (params *Params, err error) {
	params = &Params{}
	if err = json.Unmarshal([]byte(p.Params), params); err != nil {
		logrus.Errorf("Unmarshal PartReBalanceTask params error:%v task:[%v]", err, p)
		return
	}

	return
}

type AssetTransferTask struct {
	*Base        `xorm:"extends"`
	*BaseTask    `xorm:"extends"`
	RebalanceId  uint64 `xorm:"f_rebalance_id"`
	TransferType uint8  `xorm:"f_transfer_type"`
	Progress     string `xorm:"f_progress"`
	Params       string `xorm:"f_params"`
}

type TransactionTask struct {
	*Base           `xorm:"extends"`
	*BaseTask       `xorm:"extends"`
	RebalanceId     uint64 `xorm:"f_rebalance_id"`
	TransferId      uint   `xorm:"f_transfer_id"`
	TransferType    uint8  `xorm:"f_transfer_type"`
	Nonce           int    `xorm:"f_nonce"`
	ChainId         int    `xorm:"f_chain_id"`
	Params          string `xorm:"f_params"`
	Decimal         int    `xorm:"f_decimal"`
	From            string `xorm:"f_from"`
	To              string `xorm:"f_to"`
	ContractAddress string `xorm:"f_contract_address"`
	Value           string `xorm:"f_value"`
	Input_data      string `xorm:"f_input_data"`
	Cipher          string `xorm:"f_cipher"`
	EncryptData     string `xorm:"f_encryptData"`
	SignData        []byte `xorm:"f_signed_data"`
	OrderId         int    `xorm:"f_order_id"`
	Hash            string `xorm:"f_hash"`
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}

type InvestTask struct {
	*Base
	*BaseTask
	RebalanceId uint64 `xorm:"rebalance_id"`
}

type CrossTask struct {
	*Base         `xorm:"extends"`
	RebalanceId   uint64 `xorm:"rebalance_id"`
	ChainFrom     string `xorm:"chain_from"`
	ChainFromAddr string `xorm:"chain_from_addr"`
	ChainTo       string `xorm:"chain_to"`
	ChainToAddr   string `xorm:"chain_to_addr"`
	CurrencyFrom  string `xorm:"currency_from"`
	CurrencyTo    string `xorm:"currency_to"`
	Amount        string `xorm:"amount"`
	State         int    `xorm:"state"`
}

type CrossSubTask struct {
	*Base        `xorm:"extends"`
	TaskNo       uint64 `xorm:"task_no"`
	BridgeTaskId uint64 `xorm:"bridge_task_id"` //跨链桥task_id
	ParentTaskId uint64 `xorm:"parent_id"`      //父任务id
	// ChainFrom    string
	// ChainTo      string
	// CurrencyFrom string
	// CurrencyTo   string
	Amount string `xorm:"amount"`
	State  int    `xorm:"state"`
}
