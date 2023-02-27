package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/types"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

var httpCli *http.Client

func init() {
	httpCli = &http.Client{Timeout: 20 * time.Second}
}

func DoRequest(url string, method string, params interface{}) (data []byte, err error) {
	reqData, err := json.Marshal(params)
	if err != nil {
		return
	}
	return DoRequestWithHeaders(url, method, reqData, nil)
}

func DoRequestWithHeaders(url string, method string, reqData []byte, headers map[string]string) (data []byte, err error) {
	body := bytes.NewReader(reqData)
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("content-type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpCli.Do(req)
	if err != nil {
		logrus.Errorf("do http request error:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("StatusCode:%d, url:%s,method:%s", resp.StatusCode, url, method)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("read response body error:%v", err)
		return
	}
	logrus.Infof("DoRequestWithHeaders host:%s path:%s, input:%s, response:%v", req.Host, req.URL.Path, string(reqData), string(data))
	return
}

func Post(requestUrl string, bytesData []byte) (ret string, err error) {
	res, err := http.Post(requestUrl,
		"application/json;charset=utf-8", bytes.NewBuffer([]byte(bytesData)))
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	str := (*string)(unsafe.Pointer(&content)) //转化为string,优化内存
	return *str, nil
}

func GetAsset(symbol string, chain string, addr string, url string) (string, error) {
	param := types.AssetInParam{
		Symbol:      symbol,
		Chain:       chain,
		AccountAddr: addr,
	}
	msg, err := json.Marshal(param)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	url = url + "/" + "getAsset"
	str, err := Post(url, msg)
	return str, nil
}

// 目前策略：取出每个热钱包的地址--对于同一个钱包地址，需要调用两次post，第一次获取代币，第二次获取本币？
func GetHotAddress(collectTx *types.CollectTxDB, addrs []string, url string) (addr string, err error) {
	sortMap := make(map[string]float64)

	localPoint := make(map[string]int)
	var localCurrency []*types.AssetInHotwallet

	var foreignToken []*types.AssetInHotwallet
	foreignPoint := make(map[string]int)

	for _, addr := range addrs {
		logrus.Info("In GetHotAddress symbol:" + collectTx.Symbol + " collectTx.Chain:" + collectTx.Chain + " hotwallet.Addr:" + addr)
		//代币
		str, err := GetAsset(collectTx.Symbol, collectTx.Chain, addr, url)
		if err != nil {
			return "", err
		}
		assetStatus := gjson.Get(str, "status")
		pendingBalance := gjson.Get(str, "pending_withdrawal_balance")
		balance := gjson.Get(str, "balance")

		if assetStatus.Int() == 0 && pendingBalance.Int() == 0 {
			assetInHotwallet := types.AssetInHotwallet{}
			assetInHotwallet.Addr = addr
			assetInHotwallet.Balance = balance.Float()

			if collectTx.Symbol == "hui" {
				localCurrency = append(localCurrency, &assetInHotwallet)
			} else {
				foreignToken = append(foreignToken, &assetInHotwallet)
			}
		}

		//todo：这个链对应的本币应该从db查询，目前就一条链,所以先写死
		if collectTx.Symbol != "hui" { //本币
			str1, err := GetAsset("hui", collectTx.Chain, addr, url)
			if err != nil {
				return "", err
			}
			assetStatus = gjson.Get(str1, "statjius")
			pendingBalance = gjson.Get(str1, "pending_withdrawal_balance")
			balance = gjson.Get(str1, "balance")

			if assetStatus.Int() == 0 && pendingBalance.Int() == 0 {
				assetInHotwallet := types.AssetInHotwallet{}
				assetInHotwallet.Addr = addr
				assetInHotwallet.Balance = balance.Float()
				localCurrency = append(localCurrency, &assetInHotwallet)
			}
		}
	}
	//先对余额数组排序
	sort.Sort(types.AssetInHotwallets(localCurrency))
	sort.Sort(types.AssetInHotwallets(foreignToken))

	logrus.Info("本币的余额数组:")
	for index, value := range localCurrency {
		str1 := fmt.Sprintf("%d", index)
		str2 := fmt.Sprintf("%f", value.Balance)
		logrus.Info("index:" + str1 + " addr:" + value.Addr + " balance:" + str2)
		localPoint[value.Addr] = index
		sortMap[value.Addr] = float64((index + 1) * 100) //本币的权重为100
	}
	logrus.Info("代币的余额数组:")
	for index, value := range foreignToken {
		str1 := fmt.Sprintf("%d", index)
		str2 := fmt.Sprintf("%f", value.Balance)
		logrus.Info("index:" + str1 + " addr:" + value.Addr + " balance:" + str2)
		foreignPoint[value.Addr] = index
		sortMap[value.Addr] = sortMap[value.Addr] + float64(100/(index+1)) //代币越小，则这里的结果越大
	}

	logrus.Info("排序后的权重map:")
	logrus.Info(sortMap)

	//下面从大到小排序这个map，按照这个map中的point
	var listAsset []types.AssetInHotwallet
	for k, v := range sortMap {
		listAsset = append(listAsset, types.AssetInHotwallet{k, v})
	}
	sort.Slice(listAsset, func(i, j int) bool {
		return listAsset[i].Balance > listAsset[j].Balance // 降序
	})

	logrus.Info("降序后的数组:")
	logrus.Info(listAsset)

	logrus.Info("选择的地址为:")
	logrus.Info(listAsset[0].Addr)
	return listAsset[0].Addr, nil
}
