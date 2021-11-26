package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"time"
)


//used for api
type RiskReq struct {
	AppId string     `json:"appId"`
	RReq RiskControlRequest  `json:"rReq"`
}

type RiskControlRequest struct {
	ReqId        string `json:"reqId"`
	ExchangeId   int `json:"exchangeId"`   //1
	ExchangeCode string `json:"exchangeCode"`//"pro"
	Quantity     float64 `json:"quantity"`   //decimal
	FromAddress  string  `json:"fromAddress"`
	FromCoin     string  `json:"fromCoin"`
	FromChain    string  `json:"fromChain"`

	ToAddress string  `json:"toAddress"`
	ToCoin    string  `json:"toCoin"`
	ToChain   string   `json:"toChain"`

	ToAddrType int  `json:"toAddrType"`      //0-未知类型的上账地址；1-内部上账地址；2-外部上账地址
	OrderTime  int64 `json:"orderTime"`     //订单创建时间。单位【秒】，默认值为0
}

type RuleAction struct {
	Name  string   `json:"name"`
	Value string  `json:"value"`
}

type Action struct {
	Code     string   `json:"code"`
	State    int   `json:"state"`
	OrderNum int   `json:"orderNum"`
	Parars   []RuleAction  `json:"parars"`
}

type RiskData struct {
	TsvToken       string    `json:"tsvToken"`
	HandlerDown    int        `json:"handlerDown"`
	NeedRelegation bool      `json:"needRelegation"`
	HasActions     bool      `json:"hasActions"`
	Actions        []Action   `json:"actions"`
}
// Sign ...
func Sign(request *http.Request, secID string, secrKey string) {
	prepareRequestV2(request, secID)

	stringToSign := stringToSignV2(request)
	//fmt.Println("before signatureV2:\n", stringToSign)
	signature := signatureV2(stringToSign, secrKey)
	//fmt.Println("after signatureV2:\n", signature)

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
