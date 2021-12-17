package alert

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/utils"
	"time"

	"github.com/starslabhq/hermes-rebalance/config"
)

type Ding struct {
	config *config.AlertConf
}

var Dingding *Ding

func InitDingding(conf *config.AlertConf) (err error) {
	Dingding, err = newDingding(conf)
	return
}

func newDingding(conf *config.AlertConf) (d *Ding, err error) {
	d = &Ding{
		config: conf,
	}

	return
}

func (d *Ding) SendMessage(title string, content string) error {
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  content,
		},
	}

	url := d.config.URL

	if d.config.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.calcSignature(timestamp, d.config.Secret)
		url += fmt.Sprintf("&timestamp=%v&sign=%v", timestamp, sign)
	}

	_, err := utils.DoRequest(url, "POST", body)
	if err != nil {
		return err
	}

	return nil
}

func (d *Ding) SendAlert(title string, content string, atMobiles []string) error {
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  content,
		},
	}

	if atMobiles != nil && len(atMobiles) > 0 {
		body["at"] = map[string]interface{}{
			"isAtAll":   false,
			"atMobiles": atMobiles,
		}
	} else if len(d.config.Mobiles) > 0 {
		body["at"] = map[string]interface{}{
			"isAtAll":   false,
			"atMobiles": d.config.Mobiles,
		}
	} else {
		body["at"] = map[string]interface{}{
			"isAtAll": true,
		}
	}

	url := d.config.URL

	if d.config.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.calcSignature(timestamp, d.config.Secret)
		url += fmt.Sprintf("&timestamp=%v&sign=%v", timestamp, sign)
	}

	_, err := utils.DoRequest(url, "POST", body)
	if err != nil {
		return err
	}

	return nil
}

func (d *Ding) calcSignature(timestamp int64, secret string) string {
	input := fmt.Sprintf("%v\n%v", timestamp, secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(input))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
