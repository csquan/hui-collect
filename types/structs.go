package types

import (
	"math/big"
	"time"
)

type Base struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type TransactionTask struct {
	*Base        `xorm:"extends"`
	ID           uint64 `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	ParentIDs    string `xorm:"f_parent_ids"`
	UserID       string `xorm:"f_uid"`
	UUID         int64  `xorm:"f_uuid"`
	RequestId    string `xorm:"f_request_id"`
	Nonce        uint64 `xorm:"f_nonce"`
	GasPrice     string `xorm:"f_gas_price"`
	GasLimit     string `xorm:"f_gas_limit"`
	Chain        string `xorm:"f_chain"`
	From         string `xorm:"f_from"`
	To           string `xorm:"f_to"`
	ContractAddr string `xorm:"f_contract_addr"`
	Receiver     string `xorm:"f_receiver"`
	Amount       string `xorm:"f_amount"`
	Value        string `xorm:"f_value"`
	InputData    string `xorm:"f_input_data"`
	SignHash     string `xorm:"f_sign_hash"`
	TxHash       string `xorm:"f_tx_hash"`
	State        int    `xorm:"f_state"`
	Tx_type      int    `xorm:"f_type"`
	Receipt      string `xorm:"f_receipt"`
	Sig          string `xorm:"f_sig"`
	Error        string `xorm:"f_error"`
	Times        int    `xorm:"f_retry_times"`
}

// 资产表,归集源交易表
type CollectTxDB struct {
	Base              `xorm:"extends"`
	Chain             string `xorm:"f_chain"`
	Symbol            string `xorm:"f_symbol"`
	Address           string `xorm:"f_address"`
	Uid               string `xorm:"f_uid"`
	Balance           string `xorm:"f_balance"`
	Status            int    `xorm:"f_status"`
	CollectState      int    `xorm:"f_collect_state"`
	OwnerType         int    `xorm:"f_ownerType"`
	OrderId           string `xorm:"f_order_id"`
	FundFeeOrderId    string `xorm:"f_fundFee_Id"`
	BalanceBeforeFund string `xorm:"f_balance_before_fund"`
	Decimal           uint8  `xorm:"f_decimal"`
	ContractAddress   string `xorm:"f_contract_address"`
}

type Token struct {
	*Base     `xorm:"extends"`
	Threshold string `xorm:"f_threshold"`
	Chain     string `xorm:"f_chain"`
	Symbol    string `xorm:"f_symbol"`
	Address   string `xorm:"f_address"`
	Decimal   int    `xorm:"f_decimal"`
}

type HotWalletArr struct {
	hot string
}

func (t *Token) TableName() string {
	return "t_token"
}

func (t *CollectTxDB) TableName() string {
	return "t_src_tx"
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}

type HttpRes struct {
	RequestId string `json:"requestId"`
	Hash      string `json:"hash"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
}

type Data1 struct {
	UID string `json:"uid" `
}

// HttpData success data
type HttpData struct {
	Code int `json:"code" example:"0"`
	Data Data1
}

type Data struct {
	UID string `json:"uid" `
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
	Id                   string `xorm:"f_id"`
	Addr                 string `xorm:"f_addr"`
	Name                 string `xorm:"f_name"`
	Symbol               string `xorm:"f_symbol"`
	Decimals             string `xorm:"f_decimals"`
	Totoal_Supply        string `xorm:"f_total_supply"`
	Totoal_Supply_Origin string `xorm:"f_total_supply_origin"`
	Create_Time          string `xorm:"f_create_time"`
}

type SignData struct {
	Chain   string
	UID     string
	Address string
	Hash    string
}

type SigData struct {
	Signature string "json:signature"
}

type CallBackData struct {
	RequestID string
	Hash      string
}

type BalanceParam struct {
	Chain    string `json:"chain"`
	Address  string `json:"address"`
	Contract string `json:"contract"`
}

type Fund struct {
	AppId     string `json:"app_id"` // 发起提现请求的appid
	OrderId   string `json:"order_id"`
	AccountId string `json:"account_id"`
	Chain     string `json:"chain"`         // 链, btc, eth
	Symbol    string `json:"mapped_symbol"` // 币种:btc, eth, usdt
	From      string `json:"from"`          // hotwallet 地址
	To        string `json:"to"`            // hotwallet 地址
	Amount    string `json:"amount"`        // 提现金额
	Memo      string `json:"memo"`          //memo
	Extension string `json:"extension"`
}

type TokenParam struct {
	Chain  string `json:"chain"`
	Symbol string `json:"mapped_symbol"`
}

type TokenSymbol struct {
	Symbol       string "json:symbol"
	MappedSymbol string "json:mapped_symbol"
}

type Coin struct {
	MappedSymbol string `json:"mapped_symbol"`
}

type Collect struct {
	AppId     string `json:"app_id"` // 发起提现请求的appid
	OrderId   string `json:"order_id"`
	AccountId string `json:"accoount_id"`
	Chain     string `json:"chain"`  // 链, btc, eth
	Symbol    string `json:"symbol"` // 币种:btc, eth, usdt
	From      string `json:"from"`   // hotwallet 地址
	To        string `json:"to"`     // hotwallet 地址
	Amount    string `json:"amount"` // 提现金额
	Memo      string `json:"memo"`   //memo
	Extension string `json:"extension"`
}

type AssetInParam struct {
	Symbol      string `json:"mapped_symbol"`
	Chain       string `json:"chain"`
	AccountAddr string `json:"address"`
}

type AccountParam struct {
	Verified  string `json:"verified"`
	AccountId string `json:"accountId"`
	ApiKey    string `json:"apiKey"`
}

type AssetInHotwallet struct {
	Addr    string  `json:"addr"`
	Balance float64 `json:"balance"`
}

type AssetInHotwallets []*AssetInHotwallet

func (s AssetInHotwallets) Len() int {
	return len(s)
}
func (s AssetInHotwallets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s AssetInHotwallets) Less(i, j int) bool {
	return s[i].Balance < s[j].Balance
}
