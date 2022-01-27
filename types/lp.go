package types

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

//LP接口参数
type LPResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
	Data *Data  `json:"data"`
}

type SingleStrategy struct {
	VaultAddress    string `json:"vaultAddress"`
	Amount          string `json:"amount"`
	StrategyAddress string `json:"strategyAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenAddress    string `json:"tokenAddress"`
	Platform        string `json:"platform"`
	Chain           string `json:"chain"`
	ChainId         int    `json:"chainId"`
	Decimal         int    `json:"decimal"`
}

type Data struct {
	Thresholds            []*Threshold         `json:"threshold"`
	VaultInfoList         []*VaultInfo         `json:"vaultInfoList"`
	LiquidityProviderList []*LiquidityProvider `json:"liquidityProviderList"`
	SingleList            []*SingleStrategy    `json:"singleList"`
}

type Pool struct {
	Chain           string
	BaseTokenSymbol string
	VaultAddr       string
	Lps             []*LpMsg
	TxHash          string
}

type LpMsg struct {
	Amount           string
	Platform         string
	BaseTokenSymbol  string
	QuoteTokenSymbol string
	BaseTokenAmount  string
	QuoteTokenAmount string
}

func findPool(pools []*Pool, chain, base string) *Pool {
	for _, pool := range pools {
		if pool.Chain == chain && pool.BaseTokenSymbol == base {
			return pool
		}
	}
	return nil
}

func (d *Data) findVaultInfo(symbol, chain string) *ControllerInfo {
	for _, vault := range d.VaultInfoList {
		if vault.TokenSymbol == symbol {
			return vault.ActiveAmount[chain]
		}
	}
	return nil
}
func (d *Data) GetLpMsgs() ([]*Pool, error) {
	pools := make([]*Pool, 0)
	for _, lp := range d.LiquidityProviderList {
		if lp == nil || len(lp.LpInfoList) == 0 {
			continue
		}
		info := lp.LpInfoList[0]
		pool := &Pool{

			Chain:           lp.Chain,
			BaseTokenSymbol: info.BaseTokenSymbol,
			Lps:             make([]*LpMsg, 0),
		}
		vaultInfo := d.findVaultInfo(info.BaseTokenSymbol, lp.Chain)
		if vaultInfo != nil {
			pool.VaultAddr = vaultInfo.ControllerAddress
		}

		var base, quote, lpAmount = decimal.Zero, decimal.Zero, decimal.Zero

		msg := &LpMsg{}
		for _, lpInfo := range lp.LpInfoList {
			if lpInfo.BaseTokenAmount == "" {
				lpInfo.BaseTokenAmount = "0"
			}
			if lpInfo.QuoteTokenAmount == "" {
				lpInfo.QuoteTokenAmount = "0"
			}
			if lpInfo.LpAmount == "" {
				lpInfo.LpAmount = "0"
			}
			b0, err := decimal.NewFromString(lpInfo.BaseTokenAmount)
			if err != nil {
				return nil, fmt.Errorf("baseTokenAmount err:%v value:%s", err, lpInfo.BaseTokenAmount)
			}
			b1, err := decimal.NewFromString(info.QuoteTokenAmount)
			if err != nil {
				return nil, fmt.Errorf("quoteTokenAmount err:%v value:%s", err, lpInfo.BaseTokenAmount)
			}
			b2, err := decimal.NewFromString(lpInfo.LpAmount)
			if err != nil {
				return nil, fmt.Errorf("lpAmount err:%v value:%s", err, lpInfo.LpAmount)
			}
			base = base.Add(b0)
			quote = quote.Add(b1)
			lpAmount = lpAmount.Add(b2)
		}
		msg.BaseTokenAmount = base.String()
		msg.QuoteTokenAmount = quote.String()
		msg.BaseTokenSymbol = info.BaseTokenSymbol
		msg.QuoteTokenSymbol = info.QuoteTokenSymbol
		msg.Amount = lpAmount.String()
		msg.Platform = lp.Platform
		pool.Lps = append(pool.Lps, msg)

		pools = append(pools, pool)
	}
	for _, solo := range d.SingleList {
		amount, err := decimal.NewFromString(solo.Amount)
		if err != nil {
			logrus.Warnf("solo amount err:%v,v:%s", err, solo.Amount)
			continue
		}
		if amount.Cmp(decimal.Zero) == 0 {
			continue
		}
		pool := findPool(pools, solo.Chain, solo.TokenSymbol)
		if pool == nil {
			pool := &Pool{
				Chain:           solo.Chain,
				BaseTokenSymbol: solo.TokenSymbol,
				Lps: []*LpMsg{
					&LpMsg{
						BaseTokenSymbol: solo.TokenSymbol,
						BaseTokenAmount: solo.Amount,
						Platform:        solo.Platform,
					},
				},
			}
			pools = append(pools, pool)
		} else {
			msg := &LpMsg{
				BaseTokenSymbol: solo.TokenSymbol,
				BaseTokenAmount: solo.Amount,
				Platform:        solo.Platform,
			}
			pool.Lps = append(pool.Lps, msg)
		}
	}
	return pools, nil
}

type Threshold struct {
	TokenAddress    string `json:"tokenAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	Chain           string `json:"chain"`
	ChainId         int    `json:"chainId"`
	ThresholdAmount string `json:"thresholdAmount"`
	Decimal         int    `json:"decimal"`
}
type LiquidityProvider struct {
	Chain    string `json:"chain"`
	ChainId  int    `json:"chainId"`
	LpSymbol string `json:"lpSymbol"`
	//LpAmount       string    `json:"lpAmount"`
	LpTokenAddress string    `json:"lpTokenAddress"`
	Platform       string    `json:"lpPlatform"`
	LpInfoList     []*LpInfo `json:"lpInfoList"`
}
type LpInfo struct {
	LpIndex           int    `json:"lpIndex"`
	LpAmount          string `json:"lpAmount"`
	BaseTokenAddress  string `json:"baseTokenAddress"`
	QuoteTokenAddress string `json:"quoteTokenAddress"`
	BaseTokenSymbol   string `json:"baseTokenSymbol"`
	QuoteTokenSymbol  string `json:"quoteTokenSymbol"`
	BaseTokenAmount   string `json:"baseTokenAmount"`
	QuoteTokenAmount  string `json:"quoteTokenAmount"`
	StrategyAddress   string `json:"strategyAddress"`
}

type Strategy struct {
	Addr        string `json:"strategyAddress"`
	TokenSymbol string `json:"tokenSymbol"`
}
type VaultInfo struct {
	TokenSymbol  string                            `json:"tokenSymbol"`
	Chain        string                            `json:"chain"`
	Currency     string                            `json:"currency"`
	ActiveAmount map[string]*ControllerInfo        `json:"activeAmount"`
	Strategies   map[string]map[string][]*Strategy `json:"strategies"`
}

type ControllerInfo struct {
	ActiveAmount      string `json:"activeAmount"`
	ControllerAddress string `json:"vaultAddress"`
	ClaimedReward     string `json:"claimedReward"`
	VaultAmount       string `json:"vaultAmount"`
	SoloAmount        string `json:"soloAmount"`
	//下面几个字段不是从json解析出来的
	//Chain string
	//ChainID int
	//BridgeAddress string
}

//平无常接口参数

type ImpermanectLostReq struct {
	BizNo  string   `json:"bizNo"`
	LpList []*LpReq `json:"lpList"`
}

type LpReq struct {
	Chain          string       `json:"chain"`
	LpTokenAddress string       `json:"lpTokenAddress"`
	LpAmount       string       `json:"lpAmount"`
	TokenList      []*TokenInfo `json:"tokenList"`
	//Token0OriginAmount string `json:"token0OriginAmount"`
	//Token1OriginAmount string `json:"token1OriginAmount"`
}
type TokenInfo struct {
	TokenAddress      string `json:"tokenAddress"`
	TokenOriginAmount string `json:"tokenOriginAmount"`
}

type NormalResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	//Data map[string]interface{} `json:"data"`
}

type MarginStatusResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type TaskManagerResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}
