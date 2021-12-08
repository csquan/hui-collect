package tokens

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_tokens.go -package=mock
type Tokener interface {
	GetDecimals(symbol string) int
	GetCurrency(symbol, chain string) string
}

var decimals map[string]int = map[string]int{
	"BTCB": 18,
	"HBTC": 18,
	"WBTC": 18,
	"ETH":  18,
}

type Tokens struct {
}

func (t Tokens) GetDecimals(symbol string) int {
	return decimals[symbol]
}
