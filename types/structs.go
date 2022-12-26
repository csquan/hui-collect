package types

import (
	"math/big"
	"time"
)

type TransactionTask struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	UserID    string    `xorm:"f_uid"`
	UUID      int64     `xorm:"f_uuid"`
	Nonce     uint64    `xorm:"f_nonce"`
	GasPrice  string    `xorm:"f_gas_price"`
	GasLimit  string    `xorm:"f_gas_limit"`
	ChainId   int       `xorm:"f_chain_id"`
	From      string    `xorm:"f_from"`
	To        string    `xorm:"f_to"`
	InputData string    `xorm:"f_input_data"`
	SignHash  string    `xorm:"f_sign_hash"`
	TxHash    string    `xorm:"f_tx_hash"`
	State     int       `xorm:"f_state"`
	Receipt   string    `xorm:"f_receipt"`
	Signature []byte    `xorm:"f_signature"`
	Error     string    `xorm:"f_error"`
	Times     int       `xorm:"f_retry_times"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}

type Balance_Erc20 struct {
	Id             string `xorm:"id"`
	Addr           string `xorm:"addr"`
	ContractAddr   string `xorm:"contract_addr"`
	Balance        string `xorm:"balance"`
	Height         string `xorm:"height"`
	Balance_Origin string `xorm:"balance_origin"`
}

type Tx struct {
	TxType               string
	From                 string
	To                   string
	Hash                 string
	Index                string
	Value                string
	Input                string
	Nonce                string
	GasPrice             string
	GasLimit             string
	GasUsed              string
	IsContract           string
	IsContractCreate     string
	BlockTime            string
	BlockNum             string
	BlockHash            string
	ExecStatus           string
	CreateTime           string
	BlockState           string
	MaxFeePerGas         string //交易费上限
	BaseFee              string
	MaxPriorityFeePerGas string //小费上限
	BurntFees            string //baseFee*gasused
}

type Erc20Transfer struct {
	TxHash          string
	Addr            string //合约地址
	Sender          string
	Receiver        string
	Tokens          *big.Int
	LogIndex        int
	SenderBalance   *big.Int
	ReceiverBalance *big.Int
}

type Erc20Info struct {
	Id                   string `xorm:"id"`
	Addr                 string `xorm:"addr"`
	Name                 string `xorm:"name"`
	Symbol               string `xorm:"symbol"`
	Decimals             string `xorm:"decimals"`
	Totoal_Supply        string `xorm:"total_supply"`
	Totoal_Supply_Origin string `xorm:"total_supply_origin"`
	Create_Time          string `xorm:"create_time"`
}

type SignData struct {
	UID     string
	Address string
	Hash    string
}
type HttpRes struct {
	Code    int
	Message string
	Data    string
}
