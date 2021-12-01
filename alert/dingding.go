package alert

import (
	"net/http"
	"time"

	"github.com/starslabhq/hermes-rebalance/config"
)

type Ding struct {
	config *config.AlertConf
	client *http.Client
}

var Dingding *Ding

func InitDingding(conf *config.AlertConf) (err error) {
	Dingding, err = newDingding(conf)
	return
}

func newDingding(conf *config.AlertConf) (d *Ding, err error) {
	d = &Ding{
		config: conf,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	return
}

func (d *Ding) SendMessage(title string, content string) error {
	return nil
}

func (d *Ding) SendAlert(title string, content string) error {
	return nil
}
