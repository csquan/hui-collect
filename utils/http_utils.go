package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"

	"github.com/sirupsen/logrus"
)

var httpCli *http.Client

func init() {
	httpCli = &http.Client{Timeout: 20 * time.Second}
}

func JoinUrl(urlInput string, pathInput string) (string, error) {
	u, err := url.Parse(urlInput)
	if err != nil {
		logrus.Errorf("parse url error:%v", err)
		return "", err
	}
	u.Path = path.Join(u.Path, pathInput)
	return u.String(), nil
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

func CallTaskManager(conf *config.Config, path string, method string) (resp *types.TaskManagerResponse, err error) {
	var urlStr string
	if urlStr, err = JoinUrl(conf.ApiConf.TaskManager, path); err != nil {
		logrus.Errorf("parse url error:%v", err)
		return
	}
	urlStr, err = url.QueryUnescape(urlStr)
	if err != nil {
		return nil, fmt.Errorf("QueryUnescape err:%v,url:%s", err, urlStr)
	}
	var data []byte
	if data, err = DoRequest(urlStr, method, nil); err != nil {
		return
	}
	resp = &types.TaskManagerResponse{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Errorf("unmashal taskManagerResponse err:%v", err)
		return
	}
	if resp.Code != 200 {
		err = errors.New("task manager response code not 200")
		logrus.Infof("task manager resonse %v", resp)
		return
	}
	return
}
