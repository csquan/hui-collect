package types

import "time"

type TransactionTask struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	UserID    string    `xorm:"f_uid"`
	State     int       `xorm:"f_state"`
	Nonce     uint64    `xorm:"f_nonce"`
	GasPrice  string    `xorm:"f_gas_price"`
	GasLimit  string    `xorm:"f_gas_limit"`
	ChainId   int       `xorm:"f_chain_id"`
	From      string    `xorm:"f_from"`
	To        string    `xorm:"f_to"`
	InputData string    `xorm:"f_input_data"`
	SignData  string    `xorm:"f_signed_data"`
	Hash      string    `xorm:"f_hash"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}
