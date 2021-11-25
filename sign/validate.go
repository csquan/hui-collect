package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
	"fmt"
)

type DecParams struct {
	Tasks  []Task `json:"tasks"`
	TxType string `json:"tx_type"`
	RawTx  string `json:"raw_tx"`
}

type ValidatorResp struct {
	Data DecParams 	`json:"data"`
	OK 	 bool 		`json:"ok"`
}

type ValidatorReq struct {
	EncryptData string `json:"encrypt_data"`
	Cipher      string `json:"cipher"`
}

type VaResp struct {
	RawTx    string `json:"rawTx"`
	OK       bool   `json:"ok"`
}

type VaReq struct {
	VReq	ValidatorReq	`json:"vReq"`
	AppId   string          `json:"appId"`
}

func ValidateEnc(vaReq ValidatorReq, appId string) (vaResp *VaResp, err error) {
	encData := vaReq
	conf := RemoteConfig(appId)
	targetUrl := conf.Vip.GetString("validator.v1url")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	payloadBytes, err := json.Marshal(&encData)
	if err != nil {
		logrus.Errorf("validator decrytion error %v", err)
		return
	}
	body := bytes.NewReader(payloadBytes)
	//set the request header according to aws v4 signature
	req1, err := http.NewRequest("POST", targetUrl, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	req1.Host = AwsV4SigHeader

	awsKey := Key{
		AccessKey: conf.Vip.GetString("validator.v1accessKey"),
		SecretKey: conf.Vip.GetString("validator.v1secretKey"),
	}

	_, err = SignRequestWithAwsV4UseQueryString(req1, &awsKey, "blockchain", "signer")

	//Post the response
	resp, err := myclient.Do(req1)
	if err != nil {
		logrus.Errorf("Validator service check failed with error %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//unmarshall the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Read the response error %v", err)
		return nil, err
	}

	var DecData ValidatorResp
	err = json.Unmarshal(respBody, &DecData)
	if err != nil {
		logrus.Errorf("json unmarshal the dec data error %v", err)
		return nil, err
	}

	return &VaResp{
		RawTx: DecData.Data.RawTx,
		OK: DecData.OK,
	}, nil

}

type ValReq struct {
	VReq	ValidReq	`json:"vReq"`
	AppId   string          `json:"appId"`
}

type ValidReq struct {
	Id          int    `json:"id"`   //0 default
	Platform    string  `json:"platform"`
	Chain       string   `json:"chain"`
	EncryptData string `json:"encrypt_data"`
	CipherKey      string `json:"cipher_key"`
}

type ValidResp struct {
	Id       int  `json:"id"` //audit request id
	Success  bool    `json:"success"`
	Error    ValidErr   `json:"error"`
	RawTx    string  `json:"raw_tx"`
}

type ValidErr struct {
	Code     int      `json:"code"`
	Message  string   `json:"message"`
}

func Validator(vaReq ValidReq, appId string) (vaResp *VaResp, err error) {
	//encData := vaReq
	conf := RemoteConfig(appId)
	targetUrl := conf.Vip.GetString("validator.url")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	//map offset with appId
	offset := conf.Vip.GetInt("offset")
	vaReq.Id = vaReq.Id + offset
	payloadBytes, err := json.Marshal(&vaReq)
	if err != nil {
		logrus.Errorf("validator decrytion error %v", err)
		return
	}

	jsonPayload := string(payloadBytes)
	logrus.Infof("the payload json is %s", jsonPayload)

	body := bytes.NewReader(payloadBytes)
	//set the request header according to aws v4 signature

	//assemble the url for api:
	Url := targetUrl + "/" + vaReq.Platform + "/" + vaReq.Chain + "/" + "validate"
	req1, err := http.NewRequest("POST", Url, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	req1.Host = AwsV4SigHeader
	awsKey := Key{
		AccessKey: conf.Vip.GetString("validator." + vaReq.Chain + ".accessKey"),
		SecretKey: conf.Vip.GetString("validator." + vaReq.Chain + ".secretKey"),
	}
	_, err = SignRequestWithAwsV4UseQueryString(req1, &awsKey, "blockchain", "signer")
	logrus.Infof("after SignRequestWithAwsV4UseQueryString")
	//Post the response
	resp, err := myclient.Do(req1)
	if err != nil {
		logrus.Errorf("Validator service check failed with error %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	logrus.Infof("ReadAll the response body")
	//unmarshall the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Read the response error %v", err)
		return nil, err
	}

	fmt.Println(respBody)
	fmt.Println(string(respBody))
	logrus.Infof(" response body :%s",string(respBody))
	logrus.Infof("unmarshall the response body")
	var DecData ValidResp
	err = json.Unmarshal(respBody, &DecData)
	if err != nil {
		logrus.Errorf("json unmarshal the dec data error %v", err)
		return nil, err
	}

	return &VaResp{
		RawTx: DecData.RawTx,
		OK: DecData.Success,
	}, nil

}


type VaRespInfo struct {
	Version    string      `json:"version"`
	Devlang    string      `json:"devlang"`
	Success    bool        `json:"success"`
}


func ValidatorInfo() (*VaResp, error) {
	targetUrl := "https://wallet-test-4.sinnet.huobiidc.com:9528/info"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	req1, err := http.NewRequest("GET", targetUrl, nil)
	resp, err := myclient.Do(req1)
	if err != nil {
		logrus.Errorf("Validator service check failed with error %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//unmarshall the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Read the response error %v", err)
		return nil, err
	}

	var DecData VaRespInfo
	err = json.Unmarshal(respBody, &DecData)
	if err != nil {
		logrus.Errorf("json unmarshal the dec data error %v", err)
		return nil, err
	}

	return &VaResp{
		RawTx: DecData.Version,
		OK: DecData.Success,
	}, nil

}
