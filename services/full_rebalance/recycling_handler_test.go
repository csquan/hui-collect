package full_rebalance

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/types"
)

func TestAppendParams(t *testing.T) {
	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		logrus.Fatalf("create mysql cli err:%v", err)
	}
	tokens, err := dbtest.GetTokens()
	if err != nil {
		t.Fatalf("get tokens err:%v", err)

	}
	currencies, err := dbtest.GetCurrency()
	if err != nil {
		t.Fatalf("get currency err:%v", err)
	}
	r := &recyclingHandler{
		conf: &config.Config{
			Chains: map[string]*config.ChainInfo{
				"heco": &config.ChainInfo{
					BridgeAddress: "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
				},
				"bsc": &config.ChainInfo{
					RpcUrl:        "rpc_url",
					BridgeAddress: "0x74938228ae77e5fcc3504ad46fac4a965d210761",
				},
			},
		},
	}
	partRebalanceParam := &types.Params{
		SendToBridgeParams:      make([]*types.SendToBridgeParam, 0),
		CrossBalances:           make([]*types.CrossBalanceItem, 0),
		ReceiveFromBridgeParams: make([]*types.ReceiveFromBridgeParam, 0),
		InvestParams:            make([]*types.InvestParam, 0),
	}
	vaultInfo := &types.VaultInfo{
		TokenSymbol: "ETH",
		Chain:       "Heco",
		Currency:    "eth",
		ActiveAmount: map[string]*types.ControllerInfo{
			"BSC": &types.ControllerInfo{
				ControllerAddress: "0x8bf20bff6dde40a03a561ec90ae82183ee7fe22f",
				ActiveAmount:      "1.00000",
				ClaimedReward:     "2.000000",
			},
			"Heco": &types.ControllerInfo{
				ControllerAddress: "0x532a24b58067adee3192390a6ff2c7751b5efe4f",
				ActiveAmount:      "0.030200000000000000",
				ClaimedReward:     "0.000000000000000000",
			},
		},
	}
	err = r.appendParam(vaultInfo, partRebalanceParam, tokens, currencies)
	if err != nil {
		t.Fatalf("append param err:%v", err)
	}
	params, err := json.Marshal(partRebalanceParam)
	t.Logf("part rebalance:%s,err:%v", params, err)
}
