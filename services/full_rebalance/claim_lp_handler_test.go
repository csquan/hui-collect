package full_rebalance

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/tokens/mock_tokens"
	"github.com/starslabhq/hermes-rebalance/types"
)

var h *claimLPHandler

func init() {
	r := strings.NewReader(vaultClaimAbi)
	abi, err := abi.JSON(r)
	if err != nil {
		logrus.Fatalf("claim abi err:%v", err)
	}
	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		logrus.Fatalf("create mysql cli err:%v", err)
	}
	h = &claimLPHandler{
		abi: abi,
		db:  dbtest,
		conf: &config.Config{
			Chains: map[string]*config.ChainInfo{
				"eth": &config.ChainInfo{
					BridgeAddress: "bridge0",
				},
				"btc": &config.ChainInfo{
					BridgeAddress: "bridge1",
				},
				"bsc": &config.ChainInfo{
					BridgeAddress: "bridge2",
				},
			},
		},
	}
}
func TestNumber(t *testing.T) {
	num, _ := decimal.NewFromString("1.2")
	ten, _ := decimal.NewFromString("10")
	ten = ten.Pow(decimal.NewFromFloat(18))
	num = num.Mul(ten)

	t.Logf("num:%s", num.String())
}

func TestGetClaimParamsAndCreateTask(t *testing.T) {
	var (
		strategyAddr0 string = "addr0"
		strategyAddr1 string = "addr1"
		strategyAddr2 string = "addr2"
	)
	lps := []*types.LiquidityProvider{
		&types.LiquidityProvider{
			Chain: "BSC",
			LpInfoList: []*types.LpInfo{
				&types.LpInfo{
					StrategyAddress:  strategyAddr0,
					BaseTokenSymbol:  "BTCB",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1.5",
					QuoteTokenAmount: "2",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr0,
					BaseTokenSymbol:  "BTCB",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1",
					QuoteTokenAmount: "3",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr1,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1.2",
					QuoteTokenAmount: "1",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr1,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "2",
					QuoteTokenAmount: "2",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr2,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDC",
					BaseTokenAmount:  "2",
					QuoteTokenAmount: "2",
				},
			},
		},
	}
	valuts := []*types.VaultInfo{
		&types.VaultInfo{
			Currency: "btc",
			ActiveAmount: map[string]*types.ControllerInfo{
				"BSC": &types.ControllerInfo{
					ControllerAddress: "vault0",
				},
			},
		},
		&types.VaultInfo{
			Currency: "eth",
			ActiveAmount: map[string]*types.ControllerInfo{
				"BSC": &types.ControllerInfo{
					ControllerAddress: "vault1",
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := mock_tokens.NewMockTokener(ctrl)
	token.EXPECT().GetCurrency(gomock.Any(), "ETH").Return("eth").AnyTimes()
	token.EXPECT().GetCurrency(gomock.Any(), "BTCB").Return("btc")
	h.token = token
	params, err := h.getClaimParams(lps, valuts)
	if err != nil {
		t.Fatalf("get claim params err:%v", err)
	}
	b, _ := json.Marshal(params)
	t.Logf("claim params:%s", b)

	token.EXPECT().GetDecimals(gomock.Any(), gomock.Any()).Return(18, true).AnyTimes()
	tasks, err := h.createTxTask(1, params)
	if err != nil {
		t.Fatalf("create tx task err:%v", err)
	}
	b, _ = json.Marshal(tasks)
	t.Logf("tasks:%s", b)

	err = h.insertTxTasksAndUpdateState(tasks, &types.FullReBalanceTask{
		Base: &types.Base{
			ID: 1,
		},
		BaseTask: &types.BaseTask{
			State: 0,
		},
	}, types.FullReBalanceClaimLP)
	if err != nil {
		t.Fatalf("insert tx tasks and update state err:%v", err)
	}
}

func TestGetTasks(t *testing.T) {
	tasks, err := h.getTxTasks(1)
	if err != nil {
		t.Fatalf("get tasks err:%v", err)
	}
	b, _ := json.Marshal(tasks)
	t.Logf("tasks:%s", b)
}

func TestCreateClaimMsg(t *testing.T) {
	msg, err := createClaimMsg("calim_ok", []*types.TransactionTask{
		&types.TransactionTask{
			FullRebalanceId: 1,
			TransactionType: 4,
			Nonce:           2,
			GasPrice:        "10",
			GasLimit:        "20",
			Amount:          "amount",
			ChainId:         128,
			From:            "addr_from",
			To:              "addr_to",
			ChainName:       "heco",
		},
	}, &types.FullReBalanceTask{
		Base: &types.Base{
			ID: 1,
		},
		Params: "{full_params}",
	})
	if err != nil {
		t.Fatalf("claim err:%v", err)
	}
	t.Logf("claim msg:%s", msg)
}

func TestGetSolo(t *testing.T) {
	// ctrl := gomock.NewController(t)
	// defer ctrl.Finish()
	// token := mock_tokens.NewMockTokener(ctrl)
	// token.EXPECT().GetCurrency("BSC", "ETH").Return("eth").AnyTimes()
	// token.EXPECT().GetCurrency("BSC", "WBNB").Return("bnb")
	// h.token = token
	data := &types.Data{}
	content := `{"threshold":[{"tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","tokenSymbol":"ETH","chain":"Heco","chainId":128,"thresholdAmount":"0.020000000000000000","decimal":18},{"tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","tokenSymbol":"WBNB","chain":"BSC","chainId":56,"thresholdAmount":"50.000000000000000000","decimal":18},{"tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","tokenSymbol":"USDT","chain":"Heco","chainId":128,"thresholdAmount":"10.000000000000000000","decimal":18}],"vaultInfoList":[{"tokenSymbol":"ETH","chain":"Heco","currency":"eth","activeAmount":{"BSC":{"vaultAddress":"0xc9845f796280769e1fa58d5bed73037e3563b68a","activeAmount":"0.000000000000000000","claimedReward":"0.000004275748998473","soloAmount":"0.110000000000000000","vaultAmount":"0.000004275748998473","decimal":"18"},"Heco":{"vaultAddress":"0xf66532ad882dfa1a7fce8b91d19c8d953ecd771e","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x757f8311015144db1ca00c99c39219165140d86c","tokenSymbol":"ETH"}],"Biswap":[{"strategyAddress":"0x663c0a7740e0b6a7e0ec206599e0ce4fb5decc60","tokenSymbol":"ETH-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x61bf700b6b4dcdea4646fdfe93b75b0f252a557b","tokenSymbol":"ETH"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"WBNB","chain":"BSC","currency":"bnb","activeAmount":{"BSC":{"vaultAddress":"0xf2dfe126dc9a82f1fc90a9344b4ea1f01ce87193","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x35f2a79696c78a73f3c98b6ef3e327c4fad767a3","tokenSymbol":"WBNB"}],"Biswap":[{"strategyAddress":"0xa79c3946b8df2ad5053f798a077e3d56fc5d6c10","tokenSymbol":"WBNB-USDT"}],"PancakeSwap":[{"strategyAddress":"0x69e48adcab73cac8996bb35f1f3bd626a2383f8b","tokenSymbol":"WBNB-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"USDT","chain":"Heco","currency":"usdt","activeAmount":{"BSC":{"vaultAddress":"0x992524692ab2537be5e47b8216a139904466f3c8","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"30.000000000000000000","vaultAmount":"0.000000000000000007","decimal":"18"},"Heco":{"vaultAddress":"0x0b9eb942cc0988422c7f8ef907bc3aa01e8a164e","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000003","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000003","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xb6a3be91225f1de691bc2ba65112f03a5b919b97","tokenSymbol":"USDT"}],"Biswap":[],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x21a33837d0d25b75330527e2510ebb7fa2a40c65","tokenSymbol":"USDT"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}}],"liquidityProviderList":[]}`
	err := json.Unmarshal([]byte(content), data)
	if err != nil {
		t.Fatalf("json decode err:%v", err)
	}
	t.Logf("----:%d", len(data.VaultInfoList))
	params, err := h.getSoloClaimParam(data.VaultInfoList)
	if err != nil {
		t.Fatalf("get solo param err:%v", err)
	}
	b, _ := json.Marshal(params)
	t.Logf("solo params:%s", b)
}

func TestGetAll(t *testing.T) {
	c := `{"threshold":[{"tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","tokenSymbol":"ETH","chain":"Heco","chainId":128,"thresholdAmount":"0.020000000000000000","decimal":18},{"tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","tokenSymbol":"WBNB","chain":"BSC","chainId":56,"thresholdAmount":"50.000000000000000000","decimal":18},{"tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","tokenSymbol":"USDT","chain":"Heco","chainId":128,"thresholdAmount":"10.000000000000000000","decimal":18}],"vaultInfoList":[{"tokenSymbol":"ETH","chain":"Heco","currency":"eth","activeAmount":{"BSC":{"vaultAddress":"0xc9845f796280769e1fa58d5bed73037e3563b68a","activeAmount":"0.000000000000000001","claimedReward":"0.000003306383541574","soloAmount":"0.027391515952287324","vaultAmount":"0.000003306383541575","decimal":"18"},"Heco":{"vaultAddress":"0xf66532ad882dfa1a7fce8b91d19c8d953ecd771e","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x757f8311015144db1ca00c99c39219165140d86c","tokenSymbol":"ETH"}],"Biswap":[{"strategyAddress":"0x663c0a7740e0b6a7e0ec206599e0ce4fb5decc60","tokenSymbol":"ETH-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x61bf700b6b4dcdea4646fdfe93b75b0f252a557b","tokenSymbol":"ETH"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"WBNB","chain":"BSC","currency":"bnb","activeAmount":{"BSC":{"vaultAddress":"0xf2dfe126dc9a82f1fc90a9344b4ea1f01ce87193","activeAmount":"0.005955595191195971","claimedReward":"0.000000037595785768","soloAmount":"0.000000000000000000","vaultAmount":"0.005955632786981739","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x35f2a79696c78a73f3c98b6ef3e327c4fad767a3","tokenSymbol":"WBNB"}],"Biswap":[{"strategyAddress":"0xa79c3946b8df2ad5053f798a077e3d56fc5d6c10","tokenSymbol":"WBNB-USDT"}],"PancakeSwap":[{"strategyAddress":"0x69e48adcab73cac8996bb35f1f3bd626a2383f8b","tokenSymbol":"WBNB-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"USDT","chain":"Heco","currency":"usdt","activeAmount":{"BSC":{"vaultAddress":"0x992524692ab2537be5e47b8216a139904466f3c8","activeAmount":"-0.000000000000000007","claimedReward":"0.000316953207034132","soloAmount":"0.000000000000000000","vaultAmount":"0.000316953207034132","decimal":"18"},"Heco":{"vaultAddress":"0x0b9eb942cc0988422c7f8ef907bc3aa01e8a164e","activeAmount":"0.000000000000000000","claimedReward":"0.000002918660040499","soloAmount":"0.000002918660040499","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xb6a3be91225f1de691bc2ba65112f03a5b919b97","tokenSymbol":"USDT"}],"Biswap":[],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x21a33837d0d25b75330527e2510ebb7fa2a40c65","tokenSymbol":"USDT"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}}],"liquidityProviderList":[{"chain":"BSC","chainId":56,"lpSymbol":"ETH-USDT","lpTokenAddress":"0x63b30de1a998e9e64fd58a21f68d323b9bcd8f85","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":4,"lpAmount":"0.155774281534439010","strategyAddress":"0x663c0a7740e0b6a7e0ec206599e0ce4fb5decc60","baseTokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"ETH","quoteTokenSymbol":"USDT","baseTokenAmount":"0.002608484047712675","quoteTokenAmount":"9.927810501930963346"}]},{"chain":"BSC","chainId":56,"lpSymbol":"WBNB-USDT","lpTokenAddress":"0x8840c6252e2e86e545defb6da98b2a0e26d8c1ba","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":5,"lpAmount":"0.085564863013879368","strategyAddress":"0xa79c3946b8df2ad5053f798a077e3d56fc5d6c10","baseTokenAddress":"0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"WBNB","quoteTokenSymbol":"USDT","baseTokenAmount":"0.004044404808804029","quoteTokenAmount":"2.072189498069036661"}]}]}`
	data := &types.Data{}
	err := json.Unmarshal([]byte(c), data)
	if err != nil {
		t.Fatalf("json decode err:%v", err)
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := mock_tokens.NewMockTokener(ctrl)
	token.EXPECT().GetCurrency(gomock.Any(), "ETH").Return("eth").AnyTimes()
	// token.EXPECT().GetCurrency(gomock.Any(), "USDT").Return("usdt")
	token.EXPECT().GetCurrency(gomock.Any(), "WBNB").Return("bnb")
	h.token = token
	params, err := h.getClaimParams(data.LiquidityProviderList, data.VaultInfoList)
	if err != nil {
		t.Fatalf("get claim err:%v", err)
	}
	b, _ := json.Marshal(params)
	t.Logf("claim paramsall:%s", b)
}
