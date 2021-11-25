package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)


func RemoteConfig(appId string) (conf *Conf) {
	remote.SetAppID(appId)
	v := viper.New()
	v.SetConfigType("yaml")

	v.AddRemoteProvider("apollo", "http://apollo-config.system-service.huobiapps.com", "config.yaml")

	v.ReadRemoteConfig()

	v.WatchRemoteConfigOnChannel() // 启动一个goroutine来同步配置更改

	return &Conf{
		App: appId,
		Vip: v,
	}
}

type BusData struct {
	Chain    string `json:"chain"`
	Quantity    string `json:"quantity"`
	ToAddress string `json:"toAddress"`
	ToTag     string `json:"toTag"`
}

type AuditReq struct {
	AppId 	string          `json:"appId"`
	AuReq   AuditRequest    `json:"auReq"`
}

type AuditRequest struct {
	BusType        string `json:"busType"`
	BusStep   int   `json:"busStep"` //1
	BusId			string  `json:"busId"`
	BusData 	BusData    `json:"busData"`
	Result      int   `json:"result"`
}

type AuditResData struct {
	CheckResult int `json:"checkResult"`
}

type AuditResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Success bool `json:"success"`
}

func PostAuditInfo(request AuditRequest, appId string) (AuditResponse, error) {
	//init the transport client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	chain := request.BusData.Chain
	switch chain {
	case "bsc":
		request.BusData.Chain = "bnb1"
	case "heco":
		request.BusData.Chain = "ht2"
	case "eth":
		request.BusData.Chain = "eth"
	}

	//convert to lower case
	request.BusData.ToAddress = strings.ToLower(request.BusData.ToAddress)

	//fetch the audit url
	conf := RemoteConfig(appId)
	Url := conf.Vip.GetString("audit.url")

	//map the offset Id with appId here
	offset := conf.Vip.GetInt("offset")
	busId, _ :=  strconv.Atoi(request.BusId)
	busId = busId + offset

	request.BusId = strconv.FormatInt(int64(busId), 10)

	//begin the assemble the request body
	reqDataByte, err := json.Marshal(request)
	if err != nil {
		//.Error("Error when marshal the audit request", err)
		return AuditResponse{}, err
	}

	logrus.Infof("audit request body json: %s",string(reqDataByte))

	body := bytes.NewReader(reqDataByte)

	req1, err := http.NewRequest("POST", Url, body)
	req1.Header.Set("content-type", "application/json")

	//aign with aws v2
	Sign(req1, conf.Vip.GetString("audit.appId"), conf.Vip.GetString("audit.appKey"))

	resp, err := myclient.Do(req1)
	if err != nil {
		//.Error("Error when posting the request", err)
		return AuditResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//.Error("Error when reading the response", err)
		return AuditResponse{}, err
	}

	var result AuditResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		//.Error("Error when reading the response", err)
		return AuditResponse{}, err
	}

	if !result.Success {
		//.Error("The response status is failed", err)
		return AuditResponse{}, err
	}
	return result, nil
}

func (signer *Signer) audit(input string,to string,quantity string,orderID int) (AuditResponse, error)  {
	var bus BusData
	bus.Chain = chain
	bus.Quantity = quantity //保持和签名请求中的一致
	bus.ToAddress = to
	bus.ToTag = input

	var AuditInput AuditReq
	AuditInput.AppId = appId
	AuditInput.AuReq.BusType = bustype
	AuditInput.AuReq.BusStep = 1 //推荐值，不修改
	AuditInput.AuReq.BusId = fmt.Sprintf("%d", orderID)  //ID保持和validator中的id一样,确保每次调用增1
	AuditInput.AuReq.BusData = bus
	AuditInput.AuReq.Result = 1 //推荐值，不修改

	resp, err := PostAuditInfo(AuditInput.AuReq, appId)
	if err != nil {
		return resp,err
	}
	fmt.Println(resp)

	return resp,nil
}



