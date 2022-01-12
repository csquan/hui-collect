package part_rebalance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/utils"
)

func checkEventsHandled(checker EventChecker, params []*checkEventParam) (bool, error) {
	if len(params) == 0 {
		return true, nil
	}
	for _, p := range params {
		ok, err := checker.checkEventHandled(p)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

type checkEventParam struct {
	Hash          string
	ChainID       int
	StrategyAddrs []string
}

//go:generate mockgen -source=$GOFILE -destination=./mock_invest_handler.go -package=part_rebalance
type EventChecker interface {
	checkEventHandled(*checkEventParam) (bool, error)
}

type eventCheckHandler struct {
	url string
	c   *http.Client
}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
	Data bool   `json:"data"`
}

func (e *eventCheckHandler) checkEventHandled(p *checkEventParam) (result bool, err error) {
	path := fmt.Sprintf("/v1/open/hash?hash=%s&chainId=%d", p.Hash, p.ChainID)
	urlStr, err := utils.JoinUrl(e.url, path)
	if err != nil {
		logrus.Warnf("parse url error:%v", err)
		return
	}
	urlStr, err = url.QueryUnescape(urlStr)
	if err != nil {
		logrus.Warnf("checkEventHandled QueryUnescape error:%v", err)
		return
	}
	data, err := utils.DoRequest(urlStr, "GET", nil)
	if err != nil {
		return
	}
	resp := &response{}
	if err = json.Unmarshal(data, resp); err != nil {
		logrus.Warnf("unmarshar resp err:%v,body:%s", err, data)
		return
	}
	if resp.Code != 200 {
		logrus.Infof("checkEvent response %v", resp)
		return false, errors.New("response code not 200")
	}
	return resp.Data, nil
}

type investEventCheckHandler struct {
	url string
	c   *http.Client
}

func newInvestEventCheckHandler(url string, c *http.Client) *investEventCheckHandler {
	return &investEventCheckHandler{
		url: url,
		c:   c,
	}
}

func (ie *investEventCheckHandler) checkEventHandled(p *checkEventParam) (result bool, err error) {
	hash := p.Hash
	for _, sAddr := range p.StrategyAddrs {
		req, err := http.NewRequest(http.MethodGet, ie.url, nil)
		if err != nil {
			return false, fmt.Errorf("investEventChecker new req err:%v", err)
		}
		params := req.URL.Query()
		params.Add("hash", hash)
		params.Add("chainId", fmt.Sprintf("%d", p.ChainID))
		params.Add("strategyAddress", sAddr)
		req.URL.RawQuery = params.Encode()
		ret, err := ie.c.Do(req)
		if err != nil {
			return false, fmt.Errorf("investEventChecker do req err:%v,url:%s", err, req.URL.String())
		}
		defer ret.Body.Close()
		b, _ := ioutil.ReadAll(ret.Body)
		res := &response{}
		if err := json.Unmarshal(b, res); err != nil {
			return false, fmt.Errorf("investEventChecker json decode ret err:%v,ret:%s", err, b)
		}
		logrus.Infof("investEventChecker req:%s,ret:%s", req.URL.String(), b)
		if !res.Data {
			return false, err
		}
	}
	return true, nil
}
