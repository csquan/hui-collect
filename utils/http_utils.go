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

func DoPost(url string, params interface{}) (data []byte, err error){
	reqData, err := json.Marshal(params)
	if err != nil{
		return
	}
	client := &http.Client{Timeout: 123 * time.Second}
	body := bytes.NewReader(reqData)
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("do http request error:%v", err)
		return
	}
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

func DoGet(url string, params interface{}) (data []byte, err error){
	reqData, err := json.Marshal(params)
	if err != nil{
		return
	}
	body := bytes.NewReader(reqData)
	client := &http.Client{Timeout: 123 * time.Second}
	req, err := http.NewRequest("GET", url, body)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("do http request error:%v", err)
		return
	}
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