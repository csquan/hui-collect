package bridge

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

var (
	b  *Bridge
	a1 string = "0x9a0d58f60dba36d495c0ea8d1b3f8c82b03c3f8d"
	a2 string = "0x6610328142ac5ee3b554d42ea1ee14219970ac6a"
)

func init() {
	// b = &Bridge{
	// 	url:       "http://cex-bridge-api.test-15.huobiapps.com/cex-bridge-api/v1",
	// 	apiKey:    "kB01A8gsMv2CP4Tctd7XnUaG3iDE9rQJ",
	// 	secretKey: "ho8XcsfygSKTLEvPYOZ9W12k0i4IFJbC",
	// 	cli:       &http.Client{},
	// }
	var err error
	url := `http://cex-bridge-api2.tctest-1.tc-jp1.huobiapps.com/cex-bridge-api/v1`
	// url = `http://cex-bridge-api.test-15.huobiapps.com/cex-bridge-api/v1`
	b, err = NewBridge(url,
		"kB01A8gsMv2CP4Tctd7XnUaG3iDE9rQJ", "ho8XcsfygSKTLEvPYOZ9W12k0i4IFJbC", 5*time.Second)
	if err != nil {
		panic(fmt.Sprintf("new bridge err:%v", err))
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

func TestGetCurrencyList(t *testing.T) {
	clist, err := b.GetCurrencyList()
	if err != nil {
		t.Fatalf("get c list err:%v", err)
	}
	b, _ := json.Marshal(clist)
	t.Logf("clist:%s", b)
}

func TestAddAccount(t *testing.T) {
	a := &AccountAdd{
		AccounType: 2,
		ChainId:    128,
		IsMaster:   1,
		// Account:    strings.ToLower(a1),//first account
		Account: strings.ToLower(a2), //second account
	}

	id, err := b.AddAccount(a)
	if err != nil {
		t.Fatalf("add account err:%v", err)
	}
	t.Logf("accountid:%d", id)
}

func TestGetAccountList(t *testing.T) {
	accounts, err := b.GetAccountList(128)
	if err != nil {
		t.Fatalf("get accounts err:%v", err)
	}
	b, _ := json.Marshal(accounts)
	t.Logf("accounts:%s", b)
}

func TestBridgeIDs(t *testing.T) {
	hecoID, ok := b.GetChainId("HECO")
	t.Logf("heco id:%d,ok:%v", hecoID, ok)
	currencyId, ok := b.GetCurrencyID("BTC")
	t.Logf("currencyId:%d,ok:%v", currencyId, ok)
	accountId, ok := b.GetAccountId("0x9a0d58f60dba36d495c0ea8d1b3f8c82b03c3f8d", 128)
	t.Logf("accountId:%d,ok:%v", accountId, ok)
}

func TestEstimateTask(t *testing.T) {
	ret, err := b.EstimateTask(&Task{
		TaskNo:         0,
		FromAccountId:  80,
		ToAccountId:    76,
		FromCurrencyId: 1,
		ToCurrencyId:   1,
		Amount:         "1",
	})
	t.Logf("estimate ret:%v,err:%v", ret, err)
}

func TestAddTask(t *testing.T) {
	ret, err := b.AddTask(&Task{
		TaskNo:         0,
		FromAccountId:  80,
		ToAccountId:    76,
		FromCurrencyId: 1,
		ToCurrencyId:   1,
		Amount:         "0.1",
	})
	t.Logf("add task ret:%v,err:%v", ret, err)
}

func TestGetTaskDetail(t *testing.T) {
	ret, err := b.GetTaskDetail(1345)
	b, _ := json.Marshal(ret)
	t.Logf("ret:%s,err:%v", b, err)
}
