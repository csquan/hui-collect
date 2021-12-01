package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type CrossBalanceItem struct {
	FromChain    string `json:"from_chain"`
	ToChain      string `json:"to_chain"`
	FromAddr     string `json:"from_addr"`
	ToAddr       string `json:"to_addr"`
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"Amount"`
}

type ReceiveFromBridgeParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址

	Erc20ContractAddr common.Address //erc20 token地址，用于授权

	Amount *big.Int //链上精度值的amount，需要提前转换
	TaskID *big.Int
}
type InvestParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址

	StrategyAddresses  []*common.Address
	BaseTokenAmount    []*big.Int
	CounterTokenAmount []*big.Int
}

type Params struct {
	CrossBalances           []*CrossBalanceItem       `json:"cross_balances"`
	ReceiveFromBridgeParams []*ReceiveFromBridgeParam `json:"receive_from_bridge_params"`
	InvestParams            []*InvestParam            `json:"invest_params"`
}
