package bridge

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Bridge struct {
	url        string
	apiKey     string
	secretKey  string
	accounts   map[string]uint64 //key:chainid+"/"+addr v:accountId
	chains     map[string]int
	currencies map[string]int
	cli        *http.Client
}

func NewBridge(url, ak, sk string, rpcTimeout time.Duration) (*Bridge, error) {
	if url == "" || ak == "" || sk == "" {
		return nil, fmt.Errorf("param not enough for bridge")
	}
	cli := &http.Client{
		Timeout: rpcTimeout,
	}
	b := &Bridge{
		url:        url,
		apiKey:     ak,
		secretKey:  sk,
		cli:        cli,
		chains:     make(map[string]int),
		currencies: make(map[string]int),
		accounts:   make(map[string]uint64),
	}
	chainIds, err := b.loadChains()
	if err != nil {
		return nil, err
	}
	err = b.loadCurrencies()
	if err != nil {
		return nil, err
	}
	err = b.loadAccounts(chainIds)
	if err != nil {
		return nil, err
	}
	chains, _ := json.Marshal(b.chains)
	currencies, _ := json.Marshal(b.currencies)
	accounts, _ := json.Marshal(b.accounts)
	logrus.Infof("chains:%s", chains)
	logrus.Infof("currencies:%s", currencies)
	logrus.Infof("accounts:%s", accounts)
	return b, nil
}

func md5SignHex(data string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(data))
	cipherStr := md5Ctx.Sum(nil)
	hexStr := hex.EncodeToString(cipherStr)
	return strings.ToLower(hexStr)
}

func (b *Bridge) GetChainList() ([]*Chain, error) {
	form := url.Values{}
	method := "getChainList"
	form.Add("method", method)
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	rawStr := fmt.Sprintf("&timestamp=%d&method=%s&secret_key=%s", now, method, b.secretKey)
	sign := md5SignHex(rawStr)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	res, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s", method, body)
	if err != nil {
		return nil, err
	}
	ret := &ChainListRet{}
	defer res.Body.Close()
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	if ret.Code != 0 {
		return nil, fmt.Errorf("ret code err body:%s", body)
	}
	return ret.Data["chainList"], nil
}

func (b *Bridge) GetCurrencyList() ([]*Currency, error) {
	form := url.Values{}
	method := `getCurrencyList`
	form.Add("method", method)
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	rawStr := fmt.Sprintf("&timestamp=%d&method=%s&secret_key=%s", now, method, b.secretKey)
	sign := md5SignHex(rawStr)

	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s", method, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret := &CurrencyList{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, fmt.Errorf("json decode err:%v", err)
	}
	if ret.Code != 0 {
		return nil, fmt.Errorf("code err body:%s", body)
	}
	return ret.Data["currencyList"], nil
}

func (b *Bridge) AddAccount(a *AccountAdd) (uint64, error) {
	form := url.Values{}
	method := `addAccount`
	form.Add("method", method)
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	form.Add("type", fmt.Sprintf("%d", a.AccounType))
	form.Add("isMaster", fmt.Sprintf("%d", a.IsMaster))
	form.Add("masterAccountId", fmt.Sprintf("%d", a.MasterAccountId))
	form.Add("signerAccountId", fmt.Sprintf("%d", a.SignerAccountId))
	// form.Add("masterAccountId", "")
	// form.Add("signerAccountId", "")
	form.Add("chainId", fmt.Sprintf("%d", a.ChainId))
	form.Add("account", a.Account)
	form.Add("apiKey", a.APIKey)
	params := []string{"method", "timestamp", "type", "isMaster", "masterAccountId", "signerAccountId", "chainId", "account", "apiKey"}
	sort.Slice(params, func(i, j int) bool {
		return params[i] > params[j]
	})
	var rawStr string
	for _, p := range params {
		rawStr += fmt.Sprintf("&%s=%s", p, form.Get(p))
	}
	rawStr += "&secret_key=" + b.secretKey
	fmt.Println(rawStr)
	sign := md5SignHex(rawStr)
	fmt.Println(sign)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s,param:%v", method, body, a)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	ret := &AccountAddRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	if ret.Code != 0 {
		return 0, fmt.Errorf("code err body:%s", body)
	}
	return ret.Data.AccountId, nil
}

func (b *Bridge) GetAccountList(chainId int) ([]*Account, error) {
	form := url.Values{}
	method := "getAccountList"
	form.Add("method", method)
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	form.Add("chainId", fmt.Sprintf("%d", chainId))
	params := []string{"method", "timestamp", "chainId"}
	sort.Slice(params, func(i, j int) bool {
		return params[i] > params[j]
	})
	var rawStr string
	for _, p := range params {
		rawStr += fmt.Sprintf("&%s=%s", p, form.Get(p))
	}
	rawStr += "&secret_key=" + b.secretKey
	sign := md5SignHex(rawStr)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s,param:%d", method, body, chainId)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	ret := &AccountListRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	return ret.Data["accountList"], nil
}

func (b *Bridge) AddTask(t *Task) (uint64, error) {
	form := url.Values{}
	now := time.Now().Unix()

	form.Add("timestamp", fmt.Sprintf("%d", now))
	form.Add("method", "addTask")
	form.Add("taskNo", fmt.Sprintf("%d", t.TaskNo))
	form.Add("fromAccountId", fmt.Sprintf("%d", t.FromAccountId))
	form.Add("toAccountId", fmt.Sprintf("%d", t.ToAccountId))
	form.Add("fromCurrencyId", fmt.Sprintf("%d", t.FromCurrencyId))
	form.Add("toCurrencyId", fmt.Sprintf("%d", t.ToCurrencyId))
	form.Add("amount", t.Amount)
	params := []string{"timestamp", "method", "taskNo", "fromAccountId", "toAccountId", "fromCurrencyId", "toCurrencyId", "amount"}
	sort.Slice(params, func(i, j int) bool {
		return params[i] > params[j]
	})
	var rawStr string
	for _, p := range params {
		rawStr += fmt.Sprintf("&%s=%s", p, form.Get(p))
	}
	rawStr += "&secret_key=" + b.secretKey
	sign := md5SignHex(rawStr)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s,param:%v", "addTask", body, t)
	if err != nil {
		return 0, err
	}
	log.Printf("add task ret:%s", body)
	ret := &TaskAddRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	if ret.Code != 0 {
		return 0, fmt.Errorf("code err body:%s", body)
	}
	if ret.Data != nil {
		return ret.Data.TaskId, nil
	} else {
		return 0, fmt.Errorf("data empty")
	}
}

func (b *Bridge) EstimateTask(t *Task) (*EstimateTaskResult, error) {
	form := url.Values{}
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	form.Add("method", "estimateTask")
	form.Add("fromAccountId", fmt.Sprintf("%d", t.FromAccountId))
	form.Add("toAccountId", fmt.Sprintf("%d", t.ToAccountId))
	form.Add("fromCurrencyId", fmt.Sprintf("%d", t.FromCurrencyId))
	form.Add("toCurrencyId", fmt.Sprintf("%d", t.ToCurrencyId))
	params := []string{"method", "timestamp", "fromAccountId", "toAccountId", "fromCurrencyId", "toCurrencyId"}
	sort.Slice(params, func(i, j int) bool {
		return params[i] > params[j]
	})
	var rawStr string
	for _, p := range params {
		rawStr += fmt.Sprintf("&%s=%s", p, form.Get(p))
	}
	rawStr += "&secret_key=" + b.secretKey
	sign := md5SignHex(rawStr)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s,param:%v", "estimateTask", body, t)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	ret := &EstimateTaskRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	if ret.Code != 0 {
		return nil, fmt.Errorf("code err body:%s", body)
	}
	return ret.Data, nil
}

func (b *Bridge) GetTaskDetail(taskID uint64) (*TaskDetailResult, error) {
	form := url.Values{}
	now := time.Now().Unix()
	form.Add("timestamp", fmt.Sprintf("%d", now))
	form.Add("method", "getTaskDetail")
	form.Add("taskId", fmt.Sprintf("%d", taskID))
	var rawStr = fmt.Sprintf("&method=%s&taskId=%d&secret_key=%s", "getTaskDetail", taskID, b.secretKey)
	sign := md5SignHex(rawStr)
	req, err := http.NewRequest("POST", b.url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", b.apiKey)
	req.Header.Add("sign", sign)
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	res, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	logrus.Infof("method:%s,ret:%s,param:%v", "getTaskDetail", body, taskID)
	if err != nil {
		return nil, err
	}
	ret := &TaskDetailRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, fmt.Errorf("json decode err:%v,body:%s", err, body)
	}
	if ret.Code != 0 {
		return nil, fmt.Errorf("code err body:%s", body)
	}
	return ret.Data, nil
}

func (b *Bridge) GetChainId(chain string) (int, bool) {
	id, ok := b.chains[chain]
	return id, ok
}

func (b *Bridge) GetAccountId(addr string, chainId int) (uint64, bool) {
	k := fmt.Sprintf("%d/%s", chainId, addr)
	acoountId, ok := b.accounts[k]
	if !ok {
		logrus.Warnf("chainId not exist chainId:%d,accounts:%v", chainId, b.accounts)
		return 0, false
	}
	return acoountId, true
}

func (b *Bridge) GetCurrencyID(currency string) (int, bool) {
	ok, id := b.currencies[currency]
	return ok, id
}

func (b *Bridge) loadChains() ([]int, error) {
	chains, err := b.GetChainList()
	if err != nil {
		return nil, err
	}
	var ids []int
	for _, chain := range chains {
		logrus.Infof("chains name:%s,id:%d", chain.Name, chain.ChainId)
		b.chains[chain.Name] = chain.ChainId
		ids = append(ids, chain.ChainId)
	}
	return ids, nil
}

func (b *Bridge) loadCurrencies() error {
	cs, err := b.GetCurrencyList()
	if err != nil {
		return err
	}
	for _, c := range cs {
		logrus.Infof("currency name:%s,id:%d", c.Currency, c.CurrencyId)
		b.currencies[c.Currency] = int(c.CurrencyId)
	}
	return nil
}

func (b *Bridge) loadAccounts(chainIds []int) error {
	for _, chainId := range chainIds {
		accounts, err := b.GetAccountList(chainId)
		if err != nil {
			return err
		}
		for _, account := range accounts {
			key := fmt.Sprintf("%d/%s", account.ChainId, account.Account)
			b.accounts[key] = account.AccountId
		}
	}
	return nil
}
