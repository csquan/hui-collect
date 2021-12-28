package part_rebalance

import (
	"testing"

	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/clients"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

func TestCheckAmount(t *testing.T) {
	conf := &config.Config{
		Chains: map[string]*config.ChainInfo{
			"bsc": &config.ChainInfo{
				RpcUrl:  "https://bsc-dataseed.binance.org",
				Timeout: 5000,
			},
			"heco": &config.ChainInfo{
				RpcUrl:  "https://http-mainnet.hecochain.com",
				Timeout: 5000,
			},
		},
	}
	clients.Init(conf)
	alert.InitDingding(&config.DefaultAlertConf)
	i := newInitHandler(dbtest, conf)
	ok, err := i.checkSendToBridgeParam(&types.SendToBridgeParam{
		ChainName: "heco",
		From:      "",
		To:        "0xdc9ad5b483d14fd2d2cc8005b87fd0d0ea097ef8",
		Amount:    "0",
	}, &types.PartReBalanceTask{
		Base: &types.Base{
			ID: 1,
		},
		Params: `{"send_to_bridge_params":[{"chain_name":"heco","chain_id":128,"from":"0x9f0583a209fedbc404c4968e2157c2e7d4359803","to":"0xdc9ad5b483d14fd2d2cc8005b87fd0d0ea097ef8","bridge_address":"0x9f0583a209fedbc404c4968e2157c2e7d4359803","amount":"20000000000000000000","task_id":"164059423697210475800"}],"receive_from_bridge_params":[{"chain_name":"bsc","chain_id":56,"from":"0x74938228ae77e5fcc3504ad46fac4a965d210761","to":"0x401d6ab0658e52c37b4cbfa0189e81a2412252d2","erc20_contract_addr":"0x55d398326f99059ff775485246999027b3197955","amount":"20000000000000000000","task_id":"164059423697210475800"}],"cross_balances":[{"from_chain":"heco","to_chain":"bsc","from_addr":"0x9f0583a209fedbc404c4968e2157c2e7d4359803","to_addr":"0x74938228ae77e5fcc3504ad46fac4a965d210761","from_currency":"usdt","to_currency":"usdt","amount":"20"}],"invest_params":[{"chain_name":"bsc","chain_id":56,"from":"0x74938228ae77e5fcc3504ad46fac4a965d210761","to":"0x2bdadc84df36f0b5b9a2d50c39b68b87474d8aa6","strategy_addresses":["0x5dd4d1e3c93dcc4d33551ce3af4895fd78682e21"],"base_token_amount":["400000000000000000"],"counter_token_amount":["0"]},{"chain_name":"bsc","chain_id":56,"from":"0x74938228ae77e5fcc3504ad46fac4a965d210761","to":"0x1188684106af6117cf6e908fe03c4da7587c5586","strategy_addresses":["0xcf807998b179d11565806121ffd05c9c31fbd72d","0x48c34e0931ff0c3a1115e78e9d3a28e20d5c561c"],"base_token_amount":["19112023364739141","60887976635260858"],"counter_token_amount":["78000000000000000000","0"]}]}`,
	})
	if err != nil {
		t.Fatalf("check param err:%v", err)
	}
	t.Logf("check ret:%v", ok)
}
