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
	rawStr := fmt.Sprintf("&method=%s&secret_key=%s", method, b.secretKey)
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
	rawStr := fmt.Sprintf("&method=%s&secret_key=%s", method, b.secretKey)
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
	form.Add("chainId", fmt.Sprintf("%d", a.ChainId))
	form.Add("address", a.Address)
	form.Add("account", fmt.Sprintf("%d", a.Account))
	form.Add("apiKey", a.APIKey)
	params := []string{"method", "chainId", "address", "account", "apiKey"}
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
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	ret := &AccountAddRet{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, err
	}
	return ret.Data.AccountId, nil
}

func (b *Bridge) addTask(t *Task) (uint64, error) {
	form := url.Values{}
	form.Add("method", "addTask")
	form.Add("taskNo", fmt.Sprintf("%d", t.TaskNo))
	form.Add("fromAccountId", fmt.Sprintf("%d", t.FromAccountId))
	form.Add("toAccountId", fmt.Sprintf("%d", t.ToAccountId))
	form.Add("fromCurrencyId", fmt.Sprintf("%d", t.FromCurrencyId))
	form.Add("toCurrencyId", fmt.Sprintf("%d", t.ToCurrencyId))
	form.Add("amount", fmt.Sprintf("%d", t.Amount))
	params := []string{"method", "taskNo", "fromAccountId", "toAccountId", "fromCurrencyId", "toCurrencyId", "amount"}
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

func (b *Bridge) estimateTask(t *Task) (*EstimateTaskResult, error) {
	form := url.Values{}
	form.Add("method", "estimateTask")
	form.Add("fromAccountId", fmt.Sprintf("%d", t.FromAccountId))
	form.Add("toAccountId", fmt.Sprintf("%d", t.ToAccountId))
	form.Add("fromCurrencyId", fmt.Sprintf("%d", t.FromCurrencyId))
	form.Add("toCurrencyId", fmt.Sprintf("%d", t.ToCurrencyId))
	params := []string{"method", "fromAccountId", "toAccountId", "fromCurrencyId", "toCurrencyId"}

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

func (b *Bridge) getTaskDetail(taskID uint64) (*TaskDetailResult, error) {
	form := url.Values{}
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
