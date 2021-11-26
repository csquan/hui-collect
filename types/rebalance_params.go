package types

import "math/big"

type CrossBalanceItem struct {
	FromChain    string `json:"from_chain"`
	ToChain      string `json:"to_chain"`
	FromAddr     string `json:"from_addr"`
	ToAddr       string `json:"to_addr"`
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"Amount"`
}

type Params struct {
	CrossBalances   []*CrossBalanceItem     `json:"cross_balances"`
	AssetTransferIn []*AssetTransferInParam `json:"asset_transfer_in"`
	Invest          []*InvestParam          `json:"invest"`
}

type AssetTransferInParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址
	Amount    *big.Int //跨链资金大小
	TaskId    *big.Int //链下跨链任务id
}

type InvestParam struct {
	ChainId   int
	ChainName string
	From      string
	To        string //合约地址
	Data      *InvestData
}

type InvestData struct {
	address    []string
	token1List []*big.Int
	token2List []*big.Int
}
