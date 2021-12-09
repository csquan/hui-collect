package tokens

import (
	"strings"

	"github.com/starslabhq/hermes-rebalance/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock_tokens/mock_tokens.go -package=mock_tokens
type Tokener interface {
	GetDecimals(chain, symbol string) (int, bool)
	GetCurrency(chain, symbol string) string
}

type Tokens struct {
	tokens []*types.Token
	db     types.IDB
}

func NewTokens(db types.IDB) (*Tokens, error) {
	t := &Tokens{
		db: db,
	}
	err := t.load()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tokens) load() error {
	tokens, err := t.db.GetTokens()
	if err != nil {
		return err
	}
	t.tokens = tokens
	return nil
}

func (t *Tokens) GetDecimals(chain, symbol string) (int, bool) {
	for _, token := range t.tokens {
		if token.Chain == strings.ToLower(chain) && token.Symbol == strings.ToUpper(symbol) {
			return token.Decimal, true
		}
	}
	return 0, false
}

func (t *Tokens) GetCurrency(chain, symbol string) string {
	for _, token := range t.tokens {
		if token.Chain == strings.ToLower(chain) && token.Symbol == strings.ToUpper(symbol) {
			return token.Currency
		}
	}
	return ""
}
