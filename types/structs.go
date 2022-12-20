package types

import (
	"time"
)

type Base struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type BaseTask struct {
	State   int    `xorm:"f_state"`
	Message string `xorm:"f_message"`
}

type TransactionTask struct {
	*Base           `xorm:"extends"`
	*BaseTask       `xorm:"extends"`
	TransactionType int    `xorm:"f_type"`
	Nonce           uint64 `xorm:"f_nonce"`
	GasPrice        string `xorm:"f_gas_price"`
	GasLimit        string `xorm:"f_gas_limit"`
	Amount          string `xorm:"f_amount"`
	Quantity        string `xorm:"f_quantity"`
	ChainId         int    `xorm:"f_chain_id"`
	ChainName       string `xorm:"f_chain_name"`
	Params          string `xorm:"f_params"`
	From            string `xorm:"f_from"`
	To              string `xorm:"f_to"`
	ContractAddress string `xorm:"f_contract_address"` //当交易类型为授权时，此字段保存spender
	InputData       string `xorm:"f_input_data"`
	Cipher          string `xorm:"f_cipher"`
	EncryptData     string `xorm:"f_encrypt_data"`
	SignData        string `xorm:"f_signed_data"`
	OrderId         int64  `xorm:"f_order_id"`
	Hash            string `xorm:"f_hash"`
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}
