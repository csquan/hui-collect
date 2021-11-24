package bridge

import (
	"net/http"
	"testing"
)

var b *Bridge

func init() {
	b = &Bridge{
		url:       "http://cex-bridge-api.test-15.huobiapps.com/cex-bridge-api/v1",
		apiKey:    "kB01A8gsMv2CP4Tctd7XnUaG3iDE9rQJ",
		secretKey: "ho8XcsfygSKTLEvPYOZ9W12k0i4IFJbC",
		cli:       &http.Client{},
	}
}
func TestMd5Sign(t *testing.T) {
	ret := md5SignHex("&method=getAccountList&secret_key=zgUkV1wTZJ3lasYup60KHePS5MQxF4dq")
	t.Logf("%s", ret)
}

func TestGetChainList(t *testing.T) {
	chainList, err := b.GetChainList()
	if err != nil {
		t.Fatalf("getChainList err:%v", err)
	}
	t.Logf("chainList :%v", chainList)
}
