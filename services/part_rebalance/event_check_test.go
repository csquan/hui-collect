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
		Hash:    "0x4665f55638dfa14b788dead36e9ed7890ef60d448cd904670a76ac5ba5f93036",
		ChainID: 56,
		StrategyAddrs: []string{
			"0xfd8ccf8470da030dabd456de1030698c45352873",
			"0x21977f1ca61a5d761e3f028f7ec16a4d736bfa31",
		},
	})
	t.Logf("check ret:%v,err:%v", ret, err)
}
