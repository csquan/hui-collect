package config

import (
	"encoding/hex"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConf(t *testing.T) {
	msg := "436f6e74726163742076616c6964617465206572726f72203a2056616c6964617465205472616e73666572436f6e7472616374206572726f722c2062616c616e6365206973206e6f742073756666696369656e742e"
	out, _ := hex.DecodeString(msg)
	fmt.Println(string(out))
	conf, err := LoadConf("config.yaml")
	if err != nil {
		t.Fatalf("load conf err:%v", err)
	}
	output, _ := yaml.Marshal(conf)
	t.Logf("conf:%s", output)
	t.Logf("kafka servers:%v", conf.LogConf)
}
