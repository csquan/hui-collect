package part_rebalance

import (
	"encoding/json"
	"testing"

	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

func TestLpAlertMsg(t *testing.T) {
	c, err := formatLpAlertmsg("lps", 1, false, []*types.Pool{
		&types.Pool{
			Chain:           "bsc",
			BaseTokenSymbol: "btc",
			Lps: []*types.LpMsg{
				&types.LpMsg{
					Amount:           "100",
					Platform:         "biswap",
					BaseTokenSymbol:  "btc",
					BaseTokenAmount:  "1",
					QuoteTokenSymbol: "usdt",
					QuoteTokenAmount: "10",
				},
			},
		},
		&types.Pool{
			Chain:           "bsc",
			BaseTokenSymbol: "eth",
			Lps: []*types.LpMsg{
				&types.LpMsg{
					Amount:           "100",
					Platform:         "pancakewap",
					BaseTokenSymbol:  "eth",
					BaseTokenAmount:  "1",
					QuoteTokenSymbol: "usdt",
					QuoteTokenAmount: "10",
				},
			},
		},
	}, nil)
	t.Logf("alert c:%s,err:%v", c, err)
}

func TestEmptyLpAlertMsg(t *testing.T) {
	c, err := formatLpAlertmsg("lps_empty", 1, true, nil, nil)
	t.Logf("empty msg c:%s,err:%v", c, err)
}

func TestGetLpMsg(t *testing.T) {
	c := `{"threshold":[{"tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","tokenSymbol":"Cake","chain":"BSC","chainId":56,"thresholdAmount":"10.000000000000000000","decimal":18},{"tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","tokenSymbol":"WBNB","chain":"BSC","chainId":56,"thresholdAmount":"0.500000000000000000","decimal":18},{"tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","tokenSymbol":"USDT","chain":"Heco","chainId":128,"thresholdAmount":"400.000000000000000000","decimal":18},{"tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","tokenSymbol":"ETH","chain":"Heco","chainId":128,"thresholdAmount":"0.040000000000000000","decimal":18},{"tokenAddress":"0x66a79d23e58475d2738179ca52cd0b41d73f0bea","tokenSymbol":"HBTC","chain":"Heco","chainId":128,"thresholdAmount":"0.003000000000000000","decimal":18}],"vaultInfoList":[{"tokenSymbol":"Cake","chain":"BSC","currency":"cake","activeAmount":{"BSC":{"vaultAddress":"0xcb1f103da812a6b35ae9c087e4f556ad805b374c","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x59cee3d34657476b05c949181a6d53fa8048f45e","tokenSymbol":"Cake"}],"Biswap":[],"PancakeSwap":[{"strategyAddress":"0x5e3d015810348d53c94ee0bfd77bcb9493831db3","tokenSymbol":"Cake-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"WBNB","chain":"BSC","currency":"bnb","activeAmount":{"BSC":{"vaultAddress":"0x53af839424072ae72bd68bf96a8c6b9ca659503a","activeAmount":"0.000000000000000000","claimedReward":"0.016847009569118144","soloAmount":"9.557882116358817547","vaultAmount":"0.016847009569118144","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xbd5dda36762d36e8f254360db912f09b69a88add","tokenSymbol":"WBNB"}],"Biswap":[{"strategyAddress":"0x5f99f1534ef1c62215a3f4bf6027255233867258","tokenSymbol":"WBNB-USDT"}],"PancakeSwap":[{"strategyAddress":"0xa204e7ef452e9506bc38ae5e0fa6c845716623e8","tokenSymbol":"WBNB-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"USDT","chain":"Heco","currency":"usdt","activeAmount":{"BSC":{"vaultAddress":"0x08d77ba01919e25f5c886add6cca9e21de0631e1","activeAmount":"0.000000000000000000","claimedReward":"0.000374747940272439","soloAmount":"0.000000000000000000","vaultAmount":"0.000374747940272439","decimal":"18"},"Heco":{"vaultAddress":"0x11d50af7fcbecd6c83a44ecc1ec15896709b0b0c","activeAmount":"0.000000000000000000","claimedReward":"11.072972626819448422","soloAmount":"11.072972626819448422","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x82f3fe9592808afff04088649bf925927cd57f1d","tokenSymbol":"USDT"}],"Biswap":[],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x406156ae13536c02c778ad6bcc770a62ce6fdfc5","tokenSymbol":"USDT"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"ETH","chain":"Heco","currency":"eth","activeAmount":{"BSC":{"vaultAddress":"0x2b662ab01704bd3f8ad070a74bb32dfd73447d93","activeAmount":"0.000000000000000000","claimedReward":"0.000008645807892915","soloAmount":"0.000000000000000000","vaultAmount":"0.000008645807892915","decimal":"18"},"Heco":{"vaultAddress":"0xd2b9cdded71a5c0a688b36d405f45fbef5fce4eb","activeAmount":"0.000000000000000000","claimedReward":"0.005155372700089676","soloAmount":"0.005155372700089676","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x69ce0b8c319151dcca7ccce2eda3013a5569b881","tokenSymbol":"ETH"}],"Biswap":[{"strategyAddress":"0xbd83a56dd0fc6538cdecbdeaa6c001a8e990e7ba","tokenSymbol":"ETH-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x77729465cbce49721eb0e3c09b6df0cfcf93a0b6","tokenSymbol":"ETH"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"BTC","chain":"Heco","currency":"btc","activeAmount":{"BSC":{"vaultAddress":"0xcad762811fb0eb1bbca99f6ca5824631c8ebe3c2","activeAmount":"0.000000000000000000","claimedReward":"0.000057251682552266","soloAmount":"0.215000000000000000","vaultAmount":"0.000057251682552266","decimal":"18"},"Heco":{"vaultAddress":"0x1ea0937f6afef3df2c7cd4b79d74e911dfbe81da","activeAmount":"0.000000000000000000","claimedReward":"0.000248840973377603","soloAmount":"0.000248840973377603","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x80e46cc1ac6c625d66afca61e509bead8764da58","tokenSymbol":"BTCB"}],"Biswap":[{"strategyAddress":"0xf4e17045a5cc3cc59b2d1cbfd2831dd82be2aad2","tokenSymbol":"BTCB-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x31fad95d0646cbe96066d8dd8e7dc2b799b8cbed","tokenSymbol":"HBTC"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}}],"liquidityProviderList":[{"chain":"BSC","chainId":56,"lpSymbol":"ETH-USDT","lpTokenAddress":"0x63b30de1a998e9e64fd58a21f68d323b9bcd8f85","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":17,"lpAmount":"141.764989186238564048","strategyAddress":"0xbd83a56dd0fc6538cdecbdeaa6c001a8e990e7ba","baseTokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"ETH","quoteTokenSymbol":"USDT","baseTokenAmount":"3.139937520000000000","quoteTokenAmount":"6883.660437806193462701"}]},{"chain":"BSC","chainId":56,"lpSymbol":"WBNB-USDT","lpTokenAddress":"0x8840c6252e2e86e545defb6da98b2a0e26d8c1ba","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":16,"lpAmount":"5.852224406440872793","strategyAddress":"0x5f99f1534ef1c62215a3f4bf6027255233867258","baseTokenAddress":"0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"WBNB","quoteTokenSymbol":"USDT","baseTokenAmount":"0.341117883641182453","quoteTokenAmount":"116.339562193806537299"}]}],"singleList":[{"amount":"0.000000000000000000","strategyAddress":"0x59cee3d34657476b05c949181a6d53fa8048f45e","vaultAddress":"0xcb1f103da812a6b35ae9c087e4f556ad805b374c","tokenSymbol":"Cake","currency":"cake","tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"9.557882116358817547","strategyAddress":"0xbd5dda36762d36e8f254360db912f09b69a88add","vaultAddress":"0x53af839424072ae72bd68bf96a8c6b9ca659503a","tokenSymbol":"WBNB","currency":"bnb","tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"11.072972626819448422","strategyAddress":"0x406156ae13536c02c778ad6bcc770a62ce6fdfc5","vaultAddress":"0x11d50af7fcbecd6c83a44ecc1ec15896709b0b0c","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x82f3fe9592808afff04088649bf925927cd57f1d","vaultAddress":"0x08d77ba01919e25f5c886add6cca9e21de0631e1","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0x55d398326f99059ff775485246999027b3197955","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.005155372700089676","strategyAddress":"0x77729465cbce49721eb0e3c09b6df0cfcf93a0b6","vaultAddress":"0xd2b9cdded71a5c0a688b36d405f45fbef5fce4eb","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x69ce0b8c319151dcca7ccce2eda3013a5569b881","vaultAddress":"0x2b662ab01704bd3f8ad070a74bb32dfd73447d93","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.000248840973377603","strategyAddress":"0x31fad95d0646cbe96066d8dd8e7dc2b799b8cbed","vaultAddress":"0x1ea0937f6afef3df2c7cd4b79d74e911dfbe81da","tokenSymbol":"BTC","currency":"btc","tokenAddress":"0x66a79d23e58475d2738179ca52cd0b41d73f0bea","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.215000000000000000","strategyAddress":"0x80e46cc1ac6c625d66afca61e509bead8764da58","vaultAddress":"0xcad762811fb0eb1bbca99f6ca5824631c8ebe3c2","tokenSymbol":"BTCB","currency":"btc","tokenAddress":"0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18}]}`
	data := &types.Data{}
	err := json.Unmarshal([]byte(c), data)
	if err != nil {
		t.Fatalf("decode re data err:%v", err)
	}
	t.Logf("%d", len(data.LiquidityProviderList))
	msgs, err := data.GetLpMsgs()
	if len(msgs) == 0 {
		t.Fatalf("get lp msgs empty")
	}
	if err != nil {
		t.Fatalf("get lp msgs err:%v", err)
	}
	alertC, err := formatLpAlertmsg("lp_from_data", 1, true, msgs, []*types.TransactionTask{
		&types.TransactionTask{
			ChainName: "bsc",
			Hash:      "bsc_hash0",
		},
		&types.TransactionTask{
			ChainName: "bsc",
			Hash:      "bsc_hash1",
		},
	})
	t.Logf("msg:%s,err:%v", alertC, err)
}

func TestSendLpInfoMsg(t *testing.T) {
	c := `{"threshold":[{"tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","tokenSymbol":"Cake","chain":"BSC","chainId":56,"thresholdAmount":"10.000000000000000000","decimal":18},{"tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","tokenSymbol":"WBNB","chain":"BSC","chainId":56,"thresholdAmount":"0.500000000000000000","decimal":18},{"tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","tokenSymbol":"USDT","chain":"Heco","chainId":128,"thresholdAmount":"400.000000000000000000","decimal":18},{"tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","tokenSymbol":"ETH","chain":"Heco","chainId":128,"thresholdAmount":"0.040000000000000000","decimal":18},{"tokenAddress":"0x66a79d23e58475d2738179ca52cd0b41d73f0bea","tokenSymbol":"HBTC","chain":"Heco","chainId":128,"thresholdAmount":"0.003000000000000000","decimal":18}],"vaultInfoList":[{"tokenSymbol":"Cake","chain":"BSC","currency":"cake","activeAmount":{"BSC":{"vaultAddress":"0xcb1f103da812a6b35ae9c087e4f556ad805b374c","activeAmount":"0.000000000000000000","claimedReward":"0.000000000000000000","soloAmount":"0.000000000000000000","vaultAmount":"0.000000000000000000","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x59cee3d34657476b05c949181a6d53fa8048f45e","tokenSymbol":"Cake"}],"Biswap":[],"PancakeSwap":[{"strategyAddress":"0x5e3d015810348d53c94ee0bfd77bcb9493831db3","tokenSymbol":"Cake-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"WBNB","chain":"BSC","currency":"bnb","activeAmount":{"BSC":{"vaultAddress":"0x53af839424072ae72bd68bf96a8c6b9ca659503a","activeAmount":"0.000000000000000000","claimedReward":"0.016847009569118144","soloAmount":"9.557882116358817547","vaultAmount":"0.016847009569118144","decimal":"18"},"Heco":{},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0xbd5dda36762d36e8f254360db912f09b69a88add","tokenSymbol":"WBNB"}],"Biswap":[{"strategyAddress":"0x5f99f1534ef1c62215a3f4bf6027255233867258","tokenSymbol":"WBNB-USDT"}],"PancakeSwap":[{"strategyAddress":"0xa204e7ef452e9506bc38ae5e0fa6c845716623e8","tokenSymbol":"WBNB-USDT"}]},"Heco":{"Solo.top":[]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"USDT","chain":"Heco","currency":"usdt","activeAmount":{"BSC":{"vaultAddress":"0x08d77ba01919e25f5c886add6cca9e21de0631e1","activeAmount":"0.000000000000000000","claimedReward":"0.000374747940272439","soloAmount":"0.000000000000000000","vaultAmount":"0.000374747940272439","decimal":"18"},"Heco":{"vaultAddress":"0x11d50af7fcbecd6c83a44ecc1ec15896709b0b0c","activeAmount":"0.000000000000000000","claimedReward":"11.072972626819448422","soloAmount":"11.072972626819448422","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x82f3fe9592808afff04088649bf925927cd57f1d","tokenSymbol":"USDT"}],"Biswap":[],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x406156ae13536c02c778ad6bcc770a62ce6fdfc5","tokenSymbol":"USDT"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"ETH","chain":"Heco","currency":"eth","activeAmount":{"BSC":{"vaultAddress":"0x2b662ab01704bd3f8ad070a74bb32dfd73447d93","activeAmount":"0.000000000000000000","claimedReward":"0.000008645807892915","soloAmount":"0.000000000000000000","vaultAmount":"0.000008645807892915","decimal":"18"},"Heco":{"vaultAddress":"0xd2b9cdded71a5c0a688b36d405f45fbef5fce4eb","activeAmount":"0.000000000000000000","claimedReward":"0.005155372700089676","soloAmount":"0.005155372700089676","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x69ce0b8c319151dcca7ccce2eda3013a5569b881","tokenSymbol":"ETH"}],"Biswap":[{"strategyAddress":"0xbd83a56dd0fc6538cdecbdeaa6c001a8e990e7ba","tokenSymbol":"ETH-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x77729465cbce49721eb0e3c09b6df0cfcf93a0b6","tokenSymbol":"ETH"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}},{"tokenSymbol":"BTC","chain":"Heco","currency":"btc","activeAmount":{"BSC":{"vaultAddress":"0xcad762811fb0eb1bbca99f6ca5824631c8ebe3c2","activeAmount":"0.000000000000000000","claimedReward":"0.000057251682552266","soloAmount":"0.215000000000000000","vaultAmount":"0.000057251682552266","decimal":"18"},"Heco":{"vaultAddress":"0x1ea0937f6afef3df2c7cd4b79d74e911dfbe81da","activeAmount":"0.000000000000000000","claimedReward":"0.000248840973377603","soloAmount":"0.000248840973377603","vaultAmount":"0.000000000000000000","decimal":"18"},"Polygon":{}},"strategies":{"BSC":{"Solo.top":[{"strategyAddress":"0x80e46cc1ac6c625d66afca61e509bead8764da58","tokenSymbol":"BTCB"}],"Biswap":[{"strategyAddress":"0xf4e17045a5cc3cc59b2d1cbfd2831dd82be2aad2","tokenSymbol":"BTCB-USDT"}],"PancakeSwap":[]},"Heco":{"Solo.top":[{"strategyAddress":"0x31fad95d0646cbe96066d8dd8e7dc2b799b8cbed","tokenSymbol":"HBTC"}]},"Polygon":{"Solo.top":[],"QuickSwap":[]}}}],"liquidityProviderList":[{"chain":"BSC","chainId":56,"lpSymbol":"ETH-USDT","lpTokenAddress":"0x63b30de1a998e9e64fd58a21f68d323b9bcd8f85","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":17,"lpAmount":"141.764989186238564048","strategyAddress":"0xbd83a56dd0fc6538cdecbdeaa6c001a8e990e7ba","baseTokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"ETH","quoteTokenSymbol":"USDT","baseTokenAmount":"3.139937520000000000","quoteTokenAmount":"6883.660437806193462701"}]},{"chain":"BSC","chainId":56,"lpSymbol":"WBNB-USDT","lpTokenAddress":"0x8840c6252e2e86e545defb6da98b2a0e26d8c1ba","lpPlatform":"Biswap","lpInfoList":[{"lpIndex":16,"lpAmount":"5.852224406440872793","strategyAddress":"0x5f99f1534ef1c62215a3f4bf6027255233867258","baseTokenAddress":"0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c","quoteTokenAddress":"0x55d398326f99059ff775485246999027b3197955","baseTokenSymbol":"WBNB","quoteTokenSymbol":"USDT","baseTokenAmount":"0.341117883641182453","quoteTokenAmount":"116.339562193806537299"}]}],"singleList":[{"amount":"0.000000000000000000","strategyAddress":"0x59cee3d34657476b05c949181a6d53fa8048f45e","vaultAddress":"0xcb1f103da812a6b35ae9c087e4f556ad805b374c","tokenSymbol":"Cake","currency":"cake","tokenAddress":"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"9.557882116358817547","strategyAddress":"0xbd5dda36762d36e8f254360db912f09b69a88add","vaultAddress":"0x53af839424072ae72bd68bf96a8c6b9ca659503a","tokenSymbol":"WBNB","currency":"bnb","tokenAddress":"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"11.072972626819448422","strategyAddress":"0x406156ae13536c02c778ad6bcc770a62ce6fdfc5","vaultAddress":"0x11d50af7fcbecd6c83a44ecc1ec15896709b0b0c","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0xa71edc38d189767582c38a3145b5873052c3e47a","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x82f3fe9592808afff04088649bf925927cd57f1d","vaultAddress":"0x08d77ba01919e25f5c886add6cca9e21de0631e1","tokenSymbol":"USDT","currency":"usdt","tokenAddress":"0x55d398326f99059ff775485246999027b3197955","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.005155372700089676","strategyAddress":"0x77729465cbce49721eb0e3c09b6df0cfcf93a0b6","vaultAddress":"0xd2b9cdded71a5c0a688b36d405f45fbef5fce4eb","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x64ff637fb478863b7468bc97d30a5bf3a428a1fd","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.000000000000000000","strategyAddress":"0x69ce0b8c319151dcca7ccce2eda3013a5569b881","vaultAddress":"0x2b662ab01704bd3f8ad070a74bb32dfd73447d93","tokenSymbol":"ETH","currency":"eth","tokenAddress":"0x2170ed0880ac9a755fd29b2688956bd959f933f8","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18},{"amount":"0.000248840973377603","strategyAddress":"0x31fad95d0646cbe96066d8dd8e7dc2b799b8cbed","vaultAddress":"0x1ea0937f6afef3df2c7cd4b79d74e911dfbe81da","tokenSymbol":"BTC","currency":"btc","tokenAddress":"0x66a79d23e58475d2738179ca52cd0b41d73f0bea","platform":"Solo.top","chain":"Heco","chainId":128,"decimal":18},{"amount":"0.215000000000000000","strategyAddress":"0x80e46cc1ac6c625d66afca61e509bead8764da58","vaultAddress":"0xcad762811fb0eb1bbca99f6ca5824631c8ebe3c2","tokenSymbol":"BTCB","currency":"btc","tokenAddress":"0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c","platform":"Solo.top","chain":"BSC","chainId":56,"decimal":18}]}`
	data := &types.Data{}
	err := json.Unmarshal([]byte(c), data)
	if err != nil {
		t.Fatalf("decode re data err:%v", err)
	}
	// test-7
	alert.InitDingding(&config.AlertConf{
		URL:    "https://oapi.dingtalk.com/robot/send?access_token=ce39072051c4b7e067a7482ffdaa850161b058a0840db3e3d75ec5de012e0ca7",
		Secret: "SECd8a0e38eedb20f72bf79a221470e6efbd225f3a415ce2b583a7859ad397c2a99",
	})
	err = SendLpInfoWithData(data, 1, "lp_claim_before", true, []*types.TransactionTask{
		&types.TransactionTask{
			ChainName: "bsc",
			Hash:      "hash0",
		},
		&types.TransactionTask{
			ChainName: "bsc",
			Hash:      "hash1",
		},
	})
	if err != nil {
		t.Fatalf("send alert err:%v", err)
	}
}
