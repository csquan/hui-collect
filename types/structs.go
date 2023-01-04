package types

import (
	"math/big"
	"time"
)

type TransactionTask struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	UserID    string    `xorm:"f_uid"`
	UUID      int64     `xorm:"f_uuid"`
	RequestId string    `xorm:"f_request_id"`
	Nonce     uint64    `xorm:"f_nonce"`
	GasPrice  string    `xorm:"f_gas_price"`
	GasLimit  string    `xorm:"f_gas_limit"`
	ChainId   int       `xorm:"f_chain_id"`
	From      string    `xorm:"f_from"`
	To        string    `xorm:"f_to"`
	Value     string    `xorm:"f_value"`
	InputData string    `xorm:"f_input_data"`
	SignHash  string    `xorm:"f_sign_hash"`
	TxHash    string    `xorm:"f_tx_hash"`
	State     int       `xorm:"f_state"`
	Receipt   string    `xorm:"f_receipt"`
	Sig       string    `xorm:"f_sig"`
	Error     string    `xorm:"f_error"`
	Times     int       `xorm:"f_retry_times"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type CollectTxDB struct {
	Id               uint64 `xorm:"id"`
	TxType           uint8  `xorm:"tx_type"`
	AddrTo           string `xorm:"addr_to"`
	AddrFrom         string `xorm:"addr_from"`
	Hash             string `xorm:"tx_hash"`
	Index            int    `xorm:"tx_index"`
	Value            string `xorm:"tx_value"`
	Input            string `xorm:"input"`
	Nonce            uint64 `xorm:"nonce"`
	GasPrice         string `xorm:"gas_price"`
	GasLimit         uint64 `xorm:"gas_limit"`
	GasUsed          uint64 `xorm:"gas_used"`
	IsContract       bool   `xorm:"is_contract"`
	IsContractCreate bool   `xorm:"is_contract_create"`
	BlockNum         uint64 `xorm:"block_num"`
	BlockHash        string `xorm:"block_hash"`
	BlockTime        int64  `xorm:"block_time"`
	ExecStatus       uint64 `xorm:"exec_status"`
	BlockState       uint8  `xorm:"block_state"`
	//eip1559
	BaseFee              string `xorm:"base_fee"`
	MaxFeePerGas         string `xorm:"max_fee_per_gas"`
	MaxPriorityFeePerGas string `xorm:"max_priority_fee_per_gas"`
	BurntFees            string `xorm:"burnt_fees"`
	State                int    `xorm:"state"`
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

type SigData struct {
	Signature string "json:signature"
}

type HttpRes struct {
	RequestId string `json:"requestId"`
	Hash      string `json:"hash"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
}

type CallBackData struct {
	RequestID string
	Hash      string
}
