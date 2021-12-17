package alert

import (
	"github.com/starslabhq/hermes-rebalance/config"
	"testing"
)

func TestDing_SendMessage(t *testing.T) {

	InitDingding(&config.AlertConf{
		URL:     "https://oapi.dingtalk.com/robot/send?access_token=8d1c311ffd2d7a3c84091073e99b4d8e6e3ea888ddede0b55e504823b7434c5e",
		Mobiles: []string{"13120343530"},
		Secret:  "SECb3a502a0839ce6eae60ce7a4c7c25dbe907c0afdb5926492fa5dbc35c9fee1c8",
	})

	Dingding.SendMessage("This is Title", `
	#### Hello Everyone
	#### This is a test message from wang

`)
}

func TestDing_SendAlert(t *testing.T) {
	InitDingding(&config.AlertConf{
		URL:     "https://oapi.dingtalk.com/robot/send?access_token=8d1c311ffd2d7a3c84091073e99b4d8e6e3ea888ddede0b55e504823b7434c5e",
		Mobiles: []string{"13120343530"},
		Secret:  "SECb3a502a0839ce6eae60ce7a4c7c25dbe907c0afdb5926492fa5dbc35c9fee1c8",
	})

	Dingding.SendAlert("This is Title", `
	#### Hello Everyone
	#### This is a test message from wang

`, nil)
}
