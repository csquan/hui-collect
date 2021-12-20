package types

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sirupsen/logrus"
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

type FullReBalanceTask struct {
	*Base     `xorm:"extends"`
	*BaseTask `xorm:"extends"`
	Params    string `xorm:"f_params"`
}

func (p *FullReBalanceTask) TableName() string {
	return "t_full_rebalance_task"
}

type PartReBalanceTask struct {
	*Base           `xorm:"extends"`
	*BaseTask       `xorm:"extends"`
	FullRebalanceID uint64 `xorm:"f_full_rebalance_id"`
	Params          string `xorm:"f_params"`
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

func (p *PartReBalanceTask) ReadTransactionParams(txType TransactionType) (result []TransactionParamInterface, err error) {
	params := &Params{}
	if err := json.Unmarshal([]byte(p.Params), params); err != nil {
		logrus.Errorf("Unmarshal PartReBalanceTask params error:%v task:[%v]", err, p)
		return nil, err
	}
	switch txType {
	case Invest:
		for _, v := range params.InvestParams {
			result = append(result, v)
		}
		return
	case ReceiveFromBridge:
		for _, v := range params.ReceiveFromBridgeParams {
			result = append(result, v)
		}
		return
	case SendToBridge:
		for _, v := range params.SendToBridgeParams {
			result = append(result, v)
		}
		return
	default:
		return
	}
	return
}

type TransactionTask struct {
	*Base           `xorm:"extends"`
	*BaseTask       `xorm:"extends"`
	RebalanceId     uint64 `xorm:"f_rebalance_id"`
	FullRebalanceId uint64 `xorm:"f_full_rebalance_id"`
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

type CrossTask struct {
	*Base         `xorm:"extends"`
	RebalanceId   uint64 `xorm:"f_rebalance_id"`
	ChainFrom     string `xorm:"f_chain_from"`
	ChainFromAddr string `xorm:"f_chain_from_addr"`
	ChainTo       string `xorm:"f_chain_to"`
	ChainToAddr   string `xorm:"f_chain_to_addr"`
	CurrencyFrom  string `xorm:"f_currency_from"`
	CurrencyTo    string `xorm:"f_currency_to"`
	Amount        string `xorm:"f_amount"`
	State         int    `xorm:"f_state"`
}

func (t *CrossTask) TableName() string {
	return "t_cross_task"
}

type CrossSubTask struct {
	*Base        `xorm:"extends"`
	TaskNo       uint64 `xorm:"f_task_no"`
	BridgeTaskId uint64 `xorm:"f_bridge_task_id"` //跨链桥task_id
	ParentTaskId uint64 `xorm:"f_parent_id"`      //父任务id
	Amount       string `xorm:"f_amount"`
	State        int    `xorm:"f_state"`
}

type LPInfo struct {
	LPindex          int    `json:"lpIndex"`
	LPAmount         string `json:"lpAmount"`
	BaseTokenAddress string `json:"baseTokenAddress"`
	QoteTokenAddress string `json:"quoteTokenAddress"`
	BaseTokenSymbol  string `json:"baseTokenSymbol"`
	QuoteTokenSymbol string `json:"quoteTokenSymbol"`
	BaseTokenAmount  string `json:"baseTokenAmount"`
	QuoteTokenAmount string `json:"quoteTokenAmount"`
	StrategyAddress  string `json:"strategyAddress"`
}

type LP struct {
	Chain       string    `json:"chain"`
	ChainId     int       `json:"chainId"`
	Symbol      string    `json:"lpSymbol"`
	LPAmount    string    `json:"lpAmount"`
	LPTokenAddr string    `json:"lpTokenAddress"`
	LPPlatform  string    `json:"lpPlatform"`
	Infos       []*LpInfo `json:"lpInfoList"`
}

type Token struct {
	*Base       `xorm:"extends"`
	Currency    string `xorm:"f_currency"`
	Chain       string `xorm:"f_chain"`
	Symbol      string `xorm:"f_symbol"`
	Address     string `xorm:"f_address"`
	Decimal     int    `xorm:"f_decimal"`
	CrossSymbol string `xorm:"f_cross_symbol"`
}

func (t *Token) TableName() string {
	return "t_token"
}

type Currency struct {
	*Base      `xorm:"extends"`
	Name       string          `xorm:"f_name"`
	Min        decimal.Decimal `xorm:"f_cross_min"`
	CrossScale int32           `xorm:"f_cross_scale"`
}

func (t *Currency) TableName() string {
	return "t_currency"
}
