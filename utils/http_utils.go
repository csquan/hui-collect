package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
	body := bytes.NewReader(reqData)
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("content-type", "application/json")
	resp, err := httpCli.Do(req)
	if err != nil {
		logrus.Errorf("do http request error:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("http response status not 200,url:%s,method:%s,params:%v", url, method, params)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("read response body error:%v", err)
		return
	}
	return
}
