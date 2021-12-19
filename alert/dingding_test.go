package alert

import (
	"github.com/starslabhq/hermes-rebalance/config"
	"testing"
)

func TestDing_SendMessage(t *testing.T) {

	InitDingding(&config.AlertConf{
		URL:     "https://oapi.dingtalk.com/robot/send?access_token=d2bcd355f3d1d54038e147a3c5fd31858b4492dafcf002cb5f175439df13231b",
		Mobiles: []string{"13120343530"},
		Secret:  "SECa3d121d162bd2801a627fc02a9dbfb2bee9040b3bb0fb49a0d8f15e9f80e6fb0",
	})

	Dingding.SendMessage("This is Title", `
	#### Hello Everyone
	#### This is a test message from wang

`)
}

func TestDing_SendAlert(t *testing.T) {
	InitDingding(&config.AlertConf{
		URL:     "https://oapi.dingtalk.com/robot/send?access_token=d2bcd355f3d1d54038e147a3c5fd31858b4492dafcf002cb5f175439df13231b",
		Mobiles: []string{"13120343530"},
		Secret:  "SECa3d121d162bd2801a627fc02a9dbfb2bee9040b3bb0fb49a0d8f15e9f80e6fb0",
	})

	Dingding.SendAlert("This is Title", `
	#### Hello Everyone
	#### This is a test message from wang

`, nil)
}
