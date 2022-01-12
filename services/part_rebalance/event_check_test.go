package part_rebalance

import (
	"net/http"
	"testing"
)

func TestInvestEventCheck(t *testing.T) {
	ie := &investEventCheckHandler{
		url: "http://neptune-hermes-mgt-v2.test-1.huobiapps.com/v1/open/hash",
		c:   http.DefaultClient,
	}
	ret, err := ie.checkEventHandled(&checkEventParam{
		Hash:    "0xb3a669d9de51a2a9cd761e52b91cdacd8e2eeef252154ff6ecb9098677dce0ac",
		ChainID: 56,
		StrategyAddrs: []string{
			"0x7cda523ace086ba0d64eb4a64b4725253627b988",
		},
	})
	t.Logf("check ret:%v,err:%v", ret, err)
}
