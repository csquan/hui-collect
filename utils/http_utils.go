package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func DoRequest(url string, method string, params interface{}) (data []byte, err error){
	reqData, err := json.Marshal(params)
	if err != nil{
		return
	}
	client := &http.Client{Timeout: 123 * time.Second}
	body := bytes.NewReader(reqData)
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("do http request error:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200{
		err = fmt.Errorf("http response status not 200")
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("read response body error:%v", err)
		return
	}
	return
}

