package types

type CrossBalanceItem struct {
	FromChain    string `json:"from_chain"`
	ToChain      string `json:"to_chain"`
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"amount"`
}

type Params struct {
	CrossBalances []*CrossBalanceItem `json:"cross_balances"`
}
