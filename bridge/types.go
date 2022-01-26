package bridge

import "github.com/shopspring/decimal"

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_bridge.go -package=mock
type IBridge interface {
	GetCurrencyID(currency string) (int, bool)
	GetAccountId(addr string, chainId int) (uint64, bool)
	GetChainId(chain string) (int, bool)
	GetTaskDetail(taskID uint64) (*TaskDetailResult, error)
	EstimateTask(t *Task) (*EstimateTaskResult, error)
	AddTask(t *Task) (uint64, error)
	GetCrossMin(currency, fromChain, toChain string) (decimal.Decimal, error)
}
type Task struct {
	TaskNo         uint64
	FromAccountId  uint64
	ToAccountId    uint64
	FromCurrencyId int
	ToCurrencyId   int
	Amount         string
}

type AccountAdd struct {
	IsMaster        int    `json:"isMaster"` //当前场景传1
	MasterAccountId int    `json:"masterAccountId"`
	SignerAccountId int    `json:"signerAccountId"`
	AccounType      uint8  `json:"type"` //(1中转账户,2业务账户,3合约账户,4出入口钱包)
	ChainId         uint8  `json:"chainId"`
	Account         string `json:"account"` // 钱包地址/CEX UID
	APIKey          string `json:"apiKey"`  // CEX api key
}

type AccountRet struct {
	Account     string `json:"account"`
	Address     string `json:"address"`
	ChainId     uint8  `json:"chainId"`
	ChainName   string `json:"chainName"`
	ChainType   uint8  `json:"chainType"`
	AccountId   uint64 `json:"accountId"`
	AccountType uint8  `json:"type"`
}

type Chain struct {
	ChainId   int    `json:"chainId"`
	IsNode    int    `json:"isNode"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	ChainType int    `json:"type"`
}

type ChainListRet struct {
	Code int                 `json:"code"`
	Data map[string][]*Chain `json:"data"`
}

type Currency struct {
	CurrencyId int              `json:"currencyId"`
	Currency   string           `json:"currency"`
	Tokens     []*CurrencyToken `json:"chainList"`
}

type CurrencyToken struct {
	ChainId         uint64 `json:"chainId"`
	ContractAddress string `json:"contractAddress"`
	Decimals        uint64 `json:"decimals"`
	Symbol          string `json:"symbol"`
}

type CurrencyList struct {
	Code int                    `json:"code"`
	Data map[string][]*Currency `json:"data"`
}

type AccountAddResult struct {
	AccountId uint64 `json:"accountId"`
}

type AccountAddRet struct {
	Code int               `json:"code"`
	Data *AccountAddResult `json:"data"`
}

type TaskAddResult struct {
	TaskId uint64 `json:"taskId"`
}

type TaskAddRet struct {
	Code int            `json:"code"`
	Data *TaskAddResult `json:"data"`
}

type EstimateTaskResult struct {
	MinAmount    string `json:"minAmount"`
	MaxAmount    string `json:"maxAmount"`
	RemainAmount string `json:"remainAmount"`
	TotalQuota   string `json:"totalQuota"`
	SingleQuota  string `json:"singleQuota"`
	FeeEnough    int    `json:"feeEnough"`
}

type EstimateTaskRet struct {
	Code int                 `json:"code"`
	Data *EstimateTaskResult `json:"data"`
}

type TaskDetailResult struct {
	Amount        string `json:"amount"`
	CurrencyId    uint64 `json:"currencyId"`
	DstAmount     string `json:"dstAmount"`
	FromAccountId uint64 `json:"fromAccountId"`
	Status        int    `json:"status"`
	TaskId        uint64 `json:"taskId"`
	TaskNo        string `json:"taskNo"`
	ToAccountId   uint64 `json:"toAccountId"`
}

type TaskDetailRet struct {
	Code int               `json:"code"`
	Data *TaskDetailResult `json:"data"`
}

type Account struct {
	Account     string `json:"account"`
	ChainId     int    `json:"chainId"`
	ChainName   string `json:"chainName"`
	ChainType   int    `json:"chainType"`
	AccountId   uint64 `json:"accountId"`
	AccountType int    `json:"type"`
}

type AccountListRet struct {
	Code int                   `json:"code"`
	Data map[string][]*Account `json:"data"`
}

type CrossMin struct {
	MinAmount string `json:"minAmount"`
	Status    int    `json:"status"`
}
type CrossMinRet struct {
	Code int       `json:"code"`
	Data *CrossMin `json:"data"`
}
