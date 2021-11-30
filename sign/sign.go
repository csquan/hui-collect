package sign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const appId = "rebal-si-gateway"
const taskType = "withdraw"
const bustype = "starsHecoBridgeWithdraw"
const platform = "starshecobridge"
const chain = "heco"

// head key, case insensitive
const (
	headKeyData              = "date"
	headKeyXAmzDate          = "X-Amz-Date"
	headKeyAuthorization     = "authorization"
	headKeyHost              = "host"
	iSO8601BasicFormat       = "20060102T150405Z"
	iSO8601BasicFormatShort  = "20060102"
	queryKeySignature        = "X-Amz-Signature"
	queryKeyAlgorithm        = "X-Amz-Algorithm"
	queryKeyCredential       = "X-Amz-Credential"
	queryKeyDate             = "X-Amz-Date"
	queryKeySignatureHeaders = "X-Amz-SignedHeaders"
	aws4HmacSha256Algorithm  = "AWS4-HMAC-SHA256"
	AwsV4SigHeader           = "signer.blockchain.amazonaws.com"
)

var lf = []byte{'\n'}

// Key holds a set of Amazon Security Credentials.
type Key struct {
	AccessKey string
	SecretKey string
}

type Payload struct {
	Addrs         []string `json:"addrs"`
	Data          string   `json:"data"`
	Chain         string   `json:"chain"`
	EncryptParams string   `json:"encrypt_params"`
}

type SigReqData struct {
	//ToTag is the input data for contract revoking params
	ToTag    string `json:"to_tag"`
	Decimal  int    `json:"decimal"`
	Nonce    int    `json:"nonce"`
	From     string `json:"from"`
	To 		 string `json:"to"`        		//to：合约地址
	FeeStep  string `json:"fee_step"`  		//GasLimit
	FeePrice string `json:"fee_price"` 		//GasPrice
	Amount   string `json:"amount"`    		//对于合约交易，可以填写0
	TaskType string     `json:"taskType"` 	//reblance 固定使用withdraw
}

type ReqData struct {
	ToTag    string `json:"to_tag"`   		//去除0x的inputdata
	Asset    string `json:"asset"`
	Decimal  int    `json:"decimal"`
	Platform string `json:"platform"`
	Nonce    int    `json:"nonce"`
	From     string `json:"from"`
	To 		 string `json:"to"`				//to：合约地址
	FeeStep  string `json:"fee_step"`		//GasLimit
	FeePrice string `json:"fee_price"`		//GasPrice
	FeeAsset string `json:"fee_asset"`
	Amount   string `json:"amount"`
	ChainId  string `json:"chain_id"`
}

type EncParams struct {
	From      string     `json:"from"`     //from_address
	To        string     `json:"to"`       //to：合约地址
	Value     string     `json:"value"`    //value -- call contract: 0
	InputData string     `json:"inputData"` //去除0x的inputdata
	Chain     string     `json:"chain"`     //destination chain
	Quantity  string     `json:"quantity"`  //token number
	ToAddress string     `json:"toAddress"` //receipt address
	ToTag     string     `json:"toTag"`     //去除0x的inputdata
	TaskType  string     `json:"taskType"`  //reblance 固定使用withdraw

}

type Task struct {
	TaskId     string 	`json:"task_id"`
	UserId     string 	`json:"user_id"`
	OriginAddr string 	`json:"origin_addr"`
	TaskType   string 	`json:"task_type"`
}

type Response struct {
	Result bool     `json:"result"`
	Data   RespData `json:"data"`
}

type RespData struct {
	EncryptData string `json:"encrypt_data"`
	Extra       RespEx `json:"extra"`
}

type RespEx struct {
	Cipher string `json:"cipher"`
	TxHash string `json:"txhash"`
}


type UnData struct {
	FromAddr string
	Gas      int
	GasPrice decimal.Decimal
	Nonce    int
	//Proof        string
	UnsignedData []byte
	//TaskNonce    string
}


type SigningReq struct {
	AppId 	string          `json:"appId"`
	SReq    SignReq         `json:"sReq"`
}

type SignReq struct {
	SiReq   SigReqData         `json:"siReq"`
	AuReq   BusData         `json:"auReq"`
}


//SignGatewayEvmChain for EVM compatible chain support
func SignGatewayEvmChain(signReq SignReq, appId string) (encResp Response, err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	chain := signReq.AuReq.Chain
	switch chain {
	case "bsc":
		chain = "bnb1"
	case "heco":
		chain = "ht2"
	case "eth":
		chain = "eth"
	case "poly":
		chain = "matic1"
	}
	conf := config.RemoteSignerConfig(appId)
	//reqData := signReq.SiReq
	reqData := ReqData{
		Asset: conf.Vip.GetString("gateway." + chain + ".asset"),
		Platform: conf.Vip.GetString("gateway." + chain + ".platform"),
		FeeAsset: conf.Vip.GetString("gateway." + chain + ".feeAsset"),
		ChainId: conf.Vip.GetString("gateway." + chain + ".chainId"),
		ToTag: signReq.SiReq.ToTag,
		Decimal: signReq.SiReq.Decimal,
		From: strings.ToLower(signReq.SiReq.From),
		To: strings.ToLower(signReq.SiReq.To),
		FeeStep: signReq.SiReq.FeeStep,
		FeePrice: signReq.SiReq.FeePrice,
		Amount: signReq.SiReq.Amount,
		Nonce: signReq.SiReq.Nonce,
	}

	//the gateway url for signing according to different chains
	Url := conf.Vip.GetString("gateway." + chain + ".url")
	//sysAddr := conf.Vip.GetString("gateway." + chain + ".sysAddr")
	sysAddr := reqData.From

	encPara := &EncParams{
		From: sysAddr,
		To: reqData.To,
		Value: reqData.Amount,
		InputData: reqData.ToTag,
		Chain: chain,
		Quantity: signReq.AuReq.Quantity,
		ToAddress: strings.ToLower(signReq.AuReq.ToAddress),
		ToTag: reqData.ToTag,
		TaskType: signReq.SiReq.TaskType,
	}
	encParaByte, err := json.Marshal(encPara)
	if err != nil {
		return
	}

	//marshal the req data into []byte
	reqDataByte, err := json.Marshal(&reqData)
	if err != nil {
		logrus.Errorf("json unmarshal request data error: %v", err)
		return
	}
	logrus.Infof("The request ++++++++ is %s", string(reqDataByte))

	data := &Payload{
		Addrs:         []string{sysAddr},
		Chain:         conf.Vip.GetString("gateway." + chain + ".payloadChain"),
		Data:          string(reqDataByte),
		EncryptParams: string(encParaByte),
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	logrus.Infof("The request body is %s", string(payloadBytes))

	body := bytes.NewReader(payloadBytes)


	req1, err := http.NewRequest("POST", Url, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	key := &Key{
		AccessKey: conf.Vip.GetString("gateway." + chain + ".accessKey"),
		SecretKey: conf.Vip.GetString("gateway." + chain + ".secretKey"),
	}

	req1.Host = AwsV4SigHeader
	_, err = SignRequestWithAwsV4UseQueryString(req1, key, "blockchain", "signer")
	//logrus.Infof("the sp is %v", sp)
	resp, err := myclient.Do(req1)
	if err != nil {
		logrus.Errorf("http do request error")
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("response from gateway error")
		return
	}

	//unmarshal the respBody
	var result Response
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		logrus.Errorf("json unmarshal error")
		return
	}

	//check the signing result is returned with true status
	if !result.Result {
		logrus.Errorf("signing result from gateway is failed")
		return
	}

	encResp = result
	return encResp, nil

}

// Sign ...
func (k *Key) Sign(t time.Time, region, name string) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(region))
	h = ghmac(h, []byte(name))
	h = ghmac(h, []byte("aws4_request"))
	return h
}
func SignRequestWithAwsV4UseQueryString(req *http.Request, key *Key, region, name string) (sp *SignProcess, err error) {
	date := req.Header.Get(headKeyData)
	t := time.Now().UTC()
	if date != "" {
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return
		}
	}
	values := req.URL.Query()
	values.Set(headKeyXAmzDate, t.Format(iSO8601BasicFormat))

	//req.Header.Set(headKeyHost, req.Host)

	sp = new(SignProcess)
	sp.Key = key.Sign(t, region, name)

	values.Set(queryKeyAlgorithm, aws4HmacSha256Algorithm)
	values.Set(queryKeyCredential, key.AccessKey+"/"+creds(t, region, name))
	cc := bytes.NewBufferString("")
	writeHeaderList(req, nil, cc, false)
	values.Set(queryKeySignatureHeaders, cc.String())
	req.URL.RawQuery = values.Encode()

	writeStringToSign(t, req, nil, sp, false, region, name)
	values = req.URL.Query()
	values.Set(queryKeySignature, hex.EncodeToString(sp.AllSHA256))
	req.URL.RawQuery = values.Encode()

	return
}

func creds(t time.Time, region, name string) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + region + "/" + name + "/aws4_request"
}

func gsha256(data []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

type SignProcess struct {
	Key           []byte
	Body          []byte
	BodySHA256    []byte
	Request       []byte
	RequestSHA256 []byte
	All           []byte
	AllSHA256     []byte
}

func writeHeaderList(r *http.Request, signedHeadersMap map[string]bool, requestData io.Writer, isServer bool) {
	a := make([]string, 0)
	for k := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		a = append(a, strings.ToLower(k))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{';'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeStringToSign(
	t time.Time,
	r *http.Request,
	signedHeadersMap map[string]bool,
	sp *SignProcess,
	isServer bool,
	region, name string) {
	lastData := bytes.NewBufferString(aws4HmacSha256Algorithm)
	lastData.Write(lf)

	lastData.Write([]byte(t.Format(iSO8601BasicFormat)))
	lastData.Write(lf)

	lastData.Write([]byte(creds(t, region, name)))
	lastData.Write(lf)

	writeRequest(r, signedHeadersMap, sp, isServer)
	lastData.WriteString(hex.EncodeToString(sp.RequestSHA256))

	sp.All = lastData.Bytes()
	sp.AllSHA256 = ghmac(sp.Key, sp.All)
}

func writeRequest(r *http.Request, signedHeadersMap map[string]bool, sp *SignProcess, isServer bool) {
	requestData := bytes.NewBufferString("")
	//content := strings.Split(r.Host, ":")
	r.Header.Set(headKeyHost, "signer.blockchain.amazonaws.com")

	requestData.Write([]byte(r.Method))
	requestData.Write(lf)

	writeURI(r, requestData)
	requestData.Write(lf)

	writeQuery(r, requestData)
	requestData.Write(lf)

	writeHeader(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)
	requestData.Write(lf)

	writeHeaderList(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)

	writeBody(r, requestData, sp)

	sp.Request = requestData.Bytes()
	sp.RequestSHA256 = gsha256(sp.Request)
}

func writeURI(r *http.Request, requestData io.Writer) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	_, _ = requestData.Write([]byte(path))
}

func writeQuery(r *http.Request, requestData io.Writer) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		if strings.ToLower(k) == queryKeySignature {
			continue
		}
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{'&'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeHeader(r *http.Request, signedHeadersMap map[string]bool, requestData *bytes.Buffer, isServer bool) {
	a := make([]string, 0)
	for k, v := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		sort.Strings(v)
		a = append(a, strings.ToLower(k)+":"+strings.Join(v, ","))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write(lf)
		}
		_, _ = requestData.WriteString(s)
	}
}

func writeBody(r *http.Request, requestData io.StringWriter, sp *SignProcess) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	sp.Body = b

	sp.BodySHA256 = gsha256(b)
	_, _ = requestData.WriteString(hex.EncodeToString(sp.BodySHA256))
}

func (p *SignProcess) String() string {
	result := new(strings.Builder)
	fmt.Fprintf(result, "key(hex): %s\n\n", hex.EncodeToString(p.Key))
	fmt.Fprintf(result, "body:\n%s\n", string(p.Body))
	fmt.Fprintf(result, "body sha256: %s\n\n", hex.EncodeToString(p.BodySHA256))
	fmt.Fprintf(result, "request:\n%s\n", string(p.Request))
	fmt.Fprintf(result, "request sha256: %s\n\n", hex.EncodeToString(p.RequestSHA256))
	fmt.Fprintf(result, "all:\n%s\n", string(p.All))
	fmt.Fprintf(result, "all sha256: %s\n", hex.EncodeToString(p.AllSHA256))
	return result.String()
}

func  SignTx(input string,decimal int,nonce int,from string,to string,GasLimit string,GasPrice string,Amount string,quantity string,receiver string )(signResp Response, err error) {
	//delete "0x" if have
	if strings.Contains(input, "0x") {
		input = input[2:]
	}
	if strings.Contains(from, "0x") {
		from = from[2:]
	}
	if strings.Contains(to, "0x") {
		to = to[2:]
	}
	if strings.Contains(receiver, "0x") {
		receiver = receiver[2:]
	}

	var si SigReqData
	si.ToTag = input
	si.Decimal = decimal
	si.Nonce = nonce
	si.From = from
	si.To = to
	si.FeeStep = GasLimit
	si.FeePrice = GasPrice
	si.Amount = Amount
	si.TaskType = taskType

	var au BusData
	au.Chain = "heco"
	au.Quantity = quantity
	au.ToAddress = receiver
	au.ToTag = input

	var req SignReq
	req.SiReq = si
	req.AuReq = au

	resp, err := SignGatewayEvmChain(req, appId)
	if err != nil{
		signResp.Result = false
		return signResp,err
	}

	logrus.Info(resp)

	logrus.Info("EncryptData")
	logrus.Info(resp.Data.EncryptData)

	logrus.Info("CipherKey")
	logrus.Info(resp.Data.Extra.Cipher)

	return resp,nil
}





