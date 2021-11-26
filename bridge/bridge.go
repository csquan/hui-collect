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
)

type Bridge struct {
	url       string
	apiKey    string
	secretKey string
	cli       *http.Client
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
	if err != nil {
		return nil, err
	}
	log.Printf("ret:%s", body)
	ret := &ChainListRet{}
	defer res.Body.Close()
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret := &CurrencyList{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, err
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
	log.Printf("account ret:%s", body)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	ret := &AccountAddRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, err
	}
	if ret.Code != 0 {
		return 0, fmt.Errorf("code err ret:%s", body)
	}
	return ret.Data.AccountId, nil
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
	form.Add("amount", fmt.Sprintf("%d", t.Amount))
	params := []string{"timestamp", "method", "taskNo", "fromAccountId", "toAccountId", "fromCurrencyId", "toCurrencyId", "amount"}
	sort.Slice(params, func(i, j int) bool {
		return params[i] > params[j]
	})
	var rawStr string
	for _, p := range params {
		rawStr += fmt.Sprintf("&%s=%s", p, form.Get(p))
	}
	rawStr += "secret_key=%s" + b.secretKey
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
	if err != nil {
		return 0, err
	}
	ret := &TaskAddRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, err
	}
	return ret.Data.TaskId, nil
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
	rawStr += "secret_key=%s" + b.secretKey
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
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret := &EstimateTaskRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}
	ret := &TaskDetailRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, err
	}
	return ret.Data, nil
}

func (b *Bridge) GetAccountId(toAddr string) uint64 {
	return 0
}

func (b *Bridge) GetCurrencyID(currency string) uint64 {
	return 0
}
