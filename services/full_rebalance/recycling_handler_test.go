package full_rebalance

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/bridge/mock"
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
				VaultAmount:       "3.00000",
				SoloAmount:        "2.000000",
			},
			"Heco": &types.ControllerInfo{
				ControllerAddress: "0x532a24b58067adee3192390a6ff2c7751b5efe4f",
				VaultAmount:       "",
				SoloAmount:        "",
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

func TestAppendForCrossMin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vaultInfo0 := &types.VaultInfo{
		TokenSymbol: "ETH",
		Chain:       "Heco",
		Currency:    "eth",
		ActiveAmount: map[string]*types.ControllerInfo{
			"BSC": &types.ControllerInfo{
				ControllerAddress: "0x8bf20bff6dde40a03a561ec90ae82183ee7fe22f",
				VaultAmount:       "3.00000",
				SoloAmount:        "2.000000",
			},
			"Heco": &types.ControllerInfo{
				ControllerAddress: "0x532a24b58067adee3192390a6ff2c7751b5efe4f",
				VaultAmount:       "",
				SoloAmount:        "",
			},
		},
	}
	vaultInfo1 := &types.VaultInfo{
		TokenSymbol: "ETH",
		Chain:       "Heco",
		Currency:    "usdt",
		ActiveAmount: map[string]*types.ControllerInfo{
			"BSC": &types.ControllerInfo{
				ControllerAddress: "0x8bf20bff6dde40a03a561ec90ae82183ee7fe22a",
				VaultAmount:       "5.00000",
				SoloAmount:        "1.000000",
			},
			"Heco": &types.ControllerInfo{
				ControllerAddress: "0x532a24b58067adee3192390a6ff2c7751b5efe4a",
				VaultAmount:       "",
				SoloAmount:        "",
			},
		},
	}
	valuts := []*types.VaultInfo{
		vaultInfo0,
		vaultInfo1,
	}
	bridgeCli := mock.NewMockIBridge(ctrl)
	bridgeCli.EXPECT().GetCrossMin(gomock.Any(), gomock.Any(), gomock.Any()).Return(decimal.NewFromFloat(5), nil).AnyTimes()
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
		bridge: bridgeCli,
	}

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

	ret := &types.Params{
		SendToBridgeParams:      make([]*types.SendToBridgeParam, 0),
		CrossBalances:           make([]*types.CrossBalanceItem, 0),
		ReceiveFromBridgeParams: make([]*types.ReceiveFromBridgeParam, 0),
		InvestParams:            make([]*types.InvestParam, 0),
	}

	for _, valut := range valuts {
		err := r.appendParam(valut, ret, tokens, currencies)
		if err != nil {
			t.Fatalf("append err:%v", err)
		}
	}
	b, _ := json.Marshal(ret)
	t.Logf("append v2 ret:%s", b)
}

func TestJsonDataAppendFroCrossMin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	data := `{"code":200,"msg":"OK","ts":1642316536549,"data":{"threshold":[{"tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","tokenSymbol":"Cake","chain":"BSC","chainId":56,"thresholdAmount":"0.100000000000000000","decimal":18},{"tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","tokenSymbol":"ETH","chain":"Heco","chainId":128,"thresholdAmount":"0.020000000000000000","decimal":18},{"tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","tokenSymbol":"WBNB","chain":"BSC","chainId":56,"thresholdAmount":"0.010000000000000000","decimal":18},{"tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","tokenSymbol":"USDT","chain":"Heco","chainId":128,"thresholdAmount":"20.000000000000000000","decimal":18}],"vaultInfoList":[{"tokenSymbol":"Cake","chain":"BSC","currency":"cake","activeAmount":{"BSC":{"vaultAddress":"0xaab9a58d23e0e68b6e4d8c10789ad0ca4f7b8328","activeAmount":"0.002019955407748414","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.002019955407748414","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xb0045fa06741a7a04e263841b5acd0ee0f72560f","tokenSymbol":"Cake"}],"Biswap":[],"PancakeSwap":[{"strategyAddress":"0xae01575b02cf8ea16123545caa59338169cc7928","tokenSymbol":"Cake-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"ETH","chain":"Heco","currency":"eth","activeAmount":{"BSC":{"vaultAddress":"0xc9845f796280769e1fa58d5bed73037e3563b68a","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{"vaultAddress":"0xf66532ad882dfa1a7fce8b91d19c8d953ecd771e","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x757f8311015144db1ca00c99c39219165140d86c","tokenSymbol":"ETH"}],"Biswap":[{"strategyAddress":"0x663c0a7740e0b6a7e0ec206599e0ce4fb5decc60","tokenSymbol":"ETH-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x61bf700b6b4dcdea4646fdfe93b75b0f252a557b","tokenSymbol":"ETH"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"WBNB","chain":"BSC","currency":"bnb","activeAmount":{"BSC":{"vaultAddress":"0xf2dfe126dc9a82f1fc90a9344b4ea1f01ce87193","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x35f2a79696c78a73f3c98b6ef3e327c4fad767a3","tokenSymbol":"WBNB"}],"Biswap":[{"strategyAddress":"0xa79c3946b8df2ad5053f798a077e3d56fc5d6c10","tokenSymbol":"WBNB-USDT"}],"PancakeSwap":[{"strategyAddress":"0x69e48adcab73cac8996bb35f1f3bd626a2383f8b","tokenSymbol":"WBNB-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"USDT","chain":"Heco","currency":"usdt","activeAmount":{"BSC":{"vaultAddress":"0x992524692ab2537be5e47b8216a139904466f3c8","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"8.383000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{"vaultAddress":"0x0b9eb942cc0988422c7f8ef907bc3aa01e8a164e","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xb6a3be91225f1de691bc2ba65112f03a5b919b97","tokenSymbol":"USDT"}],"Biswap":[],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x21a33837d0d25b75330527e2510ebb7fa2a40c65","tokenSymbol":"USDT"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}}],"liquidityProviderList":[{"chain":"BSC","chainId":56,"lpSymbol":"Cake-USDT","lpTokenAddress":"0xa39af17ce4a8eb807e076805da1e2b8ea7d0755b","lpPlatform":"PancakeSwap","lpInfoList":[{"lpIndex":36,"lpAmount":"0.404412913559706579","strategyAddress":"0xae01575b02cf8ea16123545caa59338169cc7928","baseTokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"Cake","quoteTokenSymbol":"USDT","baseTokenAmount":"0.147980044592251586","quoteTokenAmount":"1.617000000000000000"}]}],"singleList":[{"amount":"0.000000000000000000","strategyAddress":"0xb0045fa06741a7a04e263841b5acd0ee0f72560f","vaultAddress":"0xaab9a58d23e0e68b6e4d8c10789ad0ca4f7b8328","tokenSymbol":"Cake","currency":"cake","tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x61bf700b6b4dcdea4646fdfe93b75b0f252a557b","vaultAddress":"0xf66532ad882dfa1a7fce8b91d19c8d953ecd771e","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x757f8311015144db1ca00c99c39219165140d86c","vaultAddress":"0xc9845f796280769e1fa58d5bed73037e3563b68a","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x35f2a79696c78a73f3c98b6ef3e327c4fad767a3","vaultAddress":"0xf2dfe126dc9a82f1fc90a9344b4ea1f01ce87193","tokenSymbol":"WBNB","currency":"bnb","tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x21a33837d0d25b75330527e2510ebb7fa2a40c65","vaultAddress":"0x0b9eb942cc0988422c7f8ef907bc3aa01e8a164e","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"8.383000000000000000","strategyAddress":"0xb6a3be91225f1de691bc2ba65112f03a5b919b97","vaultAddress":"0x992524692ab2537be5e47b8216a139904466f3c8","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0x55d398326f99059ff775485246999027b3197955","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18}]}}`
	lpData := &types.LPResponse{}
	err := json.Unmarshal([]byte(data), lpData)
	if err != nil {
		t.Fatalf("data err:%v", err)
	}
	bridgeCli := mock.NewMockIBridge(ctrl)
	bridgeCli.EXPECT().GetCrossMin(gomock.Any(), gomock.Any(), gomock.Any()).Return(decimal.NewFromFloat(5.5), nil).AnyTimes()
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
		bridge: bridgeCli,
	}

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
	ret := &types.Params{
		SendToBridgeParams:      make([]*types.SendToBridgeParam, 0),
		CrossBalances:           make([]*types.CrossBalanceItem, 0),
		ReceiveFromBridgeParams: make([]*types.ReceiveFromBridgeParam, 0),
		InvestParams:            make([]*types.InvestParam, 0),
	}
	for _, valut := range lpData.Data.VaultInfoList {

		err := r.appendParam(valut, ret, tokens, currencies)
		if err != nil {
			t.Fatalf("append err:%v", err)
		}
	}
	b, _ := json.Marshal(ret)
	t.Logf("append v2 ret:%s", b)
}
