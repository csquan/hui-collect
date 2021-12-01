package sign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
)

type BusData struct {
	Chain     string `json:"chain"`
	Quantity  string `json:"quantity"`
	ToAddress string `json:"toAddress"`
	ToTag     string `json:"toTag"`
}

type AuditReq struct {
	AppId string       `json:"appId"`
	AuReq AuditRequest `json:"auReq"`
}

type AuditRequest struct {
	BusType string  `json:"busType"`
	BusStep int     `json:"busStep"` //1
	BusId   string  `json:"busId"`
	BusData BusData `json:"busData"`
	Result  int     `json:"result"`
}

type AuditResData struct {
	CheckResult int `json:"checkResult"`
}

type AuditResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// Sign ...
func Sign(request *http.Request, secID string, secrKey string) {
	prepareRequestV2(request, secID)

	stringToSign := stringToSignV2(request)
	signature := signatureV2(stringToSign, secrKey)

	values := url.Values{}
	values.Set("Signature", signature)

	augmentRequestQuery(request, values)
}

func signatureV2(strToSign string, keys string) string {
	secKey := []byte(keys)
	h := hmac.New(sha256.New, secKey)
	_, _ = h.Write([]byte(strToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func stringToSignV2(request *http.Request) string {
	var host = strings.ToLower(request.URL.Host)
	hosts := strings.Split(host, ":") // 去掉端口号
	str := request.Method + "\n"
	str += hosts[0] + "\n"
	str += request.URL.Path + "\n"
	str += canonicalQueryStringV2(request)
	return str
}

func canonicalQueryStringV2(request *http.Request) string {
	return request.URL.RawQuery
}

func prepareRequestV2(request *http.Request, secID string) *http.Request {
	values := url.Values{}
	values.Set("AWSAccessKeyId", secID)
	values.Set("SignatureVersion", "2")
	values.Set("SignatureMethod", "HmacSHA256")
	values.Set("Timestamp", timestampV2())

	augmentRequestQuery(request, values)

	if request.URL.Path == "" {
		request.URL.Path += "/"
	}

	return request
}

func timestampV2() string {
	return time.Now().UTC().Format(timeFormatV2)
}

func augmentRequestQuery(request *http.Request, values url.Values) {
	for key, array := range request.URL.Query() {
		for _, value := range array {
			values.Set(key, value)
		}
	}

	request.URL.RawQuery = values.Encode()
}

const timeFormatV2 = "2006-01-02T15:04:05"

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
	case "poly":
		request.BusData.Chain = "matic1"
	}

	//convert to lower case
	request.BusData.ToAddress = strings.ToLower(request.BusData.ToAddress)

	//fetch the audit url
	conf := config.RemoteSignerConfig(appId)
	Url := conf.Vip.GetString("audit.url")

	//map the offset Id with appId here
	offset := conf.Vip.GetInt("offset")
	busId, _ := strconv.Atoi(request.BusId)
	busId = busId + offset

	request.BusId = strconv.FormatInt(int64(busId), 10)

	//begin the assemble the request body
	reqDataByte, err := json.Marshal(request)
	if err != nil {
		//.Error("Error when marshal the audit request", err)
		return AuditResponse{}, err
	}

	logrus.Infof("audit request body json: %s", string(reqDataByte))

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

func AuditTx(input string, to string, quantity string, orderID int64) (AuditResponse, error) {

	if strings.Contains(input, "0x") {
		input = input[2:]
	}
	if strings.Contains(to, "0x") {
		to = to[2:]
	}

	var bus BusData
	bus.Chain = chain
	bus.Quantity = quantity //保持和签名请求中的一致
	bus.ToAddress = to
	bus.ToTag = input

	var AuditInput AuditReq
	AuditInput.AppId = appId
	AuditInput.AuReq.BusType = bustype
	AuditInput.AuReq.BusStep = 1                        //推荐值，不修改
	AuditInput.AuReq.BusId = fmt.Sprintf("%d", orderID) //ID保持和validator中的id一样,确保每次调用增1
	AuditInput.AuReq.BusData = bus
	AuditInput.AuReq.Result = 1 //推荐值，不修改
	resp, err := PostAuditInfo(AuditInput.AuReq, appId)
	if err != nil {
		return resp, err
	}
	logrus.Info(resp)
	return resp, nil
}
