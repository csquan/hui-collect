package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var Env string

const (
	appID string = "HuiCollect"
)

type Conf struct {
	App string
	Vip *viper.Viper
}

type DataBaseConf struct {
	DB string `mapstructure:"db"` //DB 连接信息
}

type AccountConf struct {
	EndPoint string `mapstructure:"endpoint"`
}

type WalletConf struct {
	Url string `mapstructure:"url"`
}

type ChainNodeConf struct {
	Url string `mapstructure:"url"`
}

type SingleFeeConf struct {
	Fee string `mapstructure:"fee"`
}

type TrxSingleFeeConf struct {
	Fee string `mapstructure:"fee"`
}

type Trx20SingleFeeConf struct {
	Fee string `mapstructure:"fee"`
}

type TokenConf struct {
	Url string `mapstructure:"url"`
}

type GasConf struct {
	Addr string `mapstructure:"addr"`
}

type ServerConf struct {
	Port string `yaml:"port"`
}

type Config struct {
	AppName          string `mapstructure:"app_name"`
	ProfPort         int    `mapstructure:"prof_port"`
	QueryInterval    time.Duration
	QueryIntervalInt uint64                `mapstructure:"query_interval"`
	DataBase         DataBaseConf          `mapstructure:"database"`
	Account          AccountConf           `mapstructure:"account"`
	Wallet           WalletConf            `mapstructure:"wallet"`
	ChainNode        ChainNodeConf         `mapstructure:"chainnode"`
	SingleFee        SingleFeeConf         `mapstructure:"single_fee"`
	Trx20SingleFee   Trx20SingleFeeConf    `mapstructure:"trx20_single_fee"`
	TrxSingleFee     TrxSingleFeeConf      `mapstructure:"trx_single_fee"`
	Token            TokenConf             `mapstructure:"token"`
	Gas              GasConf               `mapstructure:"gas"`
	LogConf          Log                   `mapstructure:"log"`
	Chains           map[string]*ChainInfo `mapstructure:"chains"`
	Env              string                `mapstructure:"env"`
	ServerConf       ServerConf            `mapstructure:"server"`
}

type ChainInfo struct {
	ID      int    `mapstructure:"id"`
	RpcUrl  string `mapstructure:"rpc_url"`
	Timeout int    `mapstructure:"timeout"`
}

func (c *Config) MustGetChainInfo(chain string) *ChainInfo {
	chain = strings.ToLower(chain)
	if v, ok := c.Chains[chain]; ok {
		return v
	}
	logrus.Fatalf("chain info not found chain:%s", chain)
	return nil
}

func (c *Config) init() {
	c.QueryInterval = time.Duration(c.QueryIntervalInt) * time.Millisecond
}

type Log struct {
	Stdout stdout `mapstructure:"stdout"`
	File   file   `mapstructure:"file"`
	Kafka  kafka  `mapstructure:"kafka"`
}

type stdout struct {
	Enable bool `mapstructure:"enable"`
	Level  int  `mapstructure:"level"`
}

type file struct {
	Enable bool   `mapstructure:"enable"`
	Path   string `mapstructure:"path"`
	Level  int    `mapstructure:"level"`
}

type kafka struct {
	Enable  bool     `mapstructure:"enable"`
	Level   int      `mapstructure:"level"`
	Brokers []string `mapstructure:"kafka_servers"`
	Topic   string   `mapstructure:"topic"`
}

func LoadConf(fpath string) (*Config, error) {
	if fpath == "" {
		return nil, fmt.Errorf("fpath empty")
	}

	if !strings.HasSuffix(strings.ToLower(fpath), ".yaml") {
		return nil, fmt.Errorf("fpath must has suffix of .yaml")
	}

	//load default config first
	conf := &Config{
		LogConf:          DefaultLogConfig,
		QueryIntervalInt: 3000,
	}

	vip := viper.New()
	vip.SetConfigType("yaml")

	err := localConfig(fpath, vip)
	if err != nil {
		return nil, err
	}

	err = vip.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	conf.init()

	return conf, nil
}

func localConfig(filename string, v *viper.Viper) error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("err : %v", err)
	}

	v.AddConfigPath(path) //设置读取的文件路径

	v.SetConfigName(filename) //设置读取的文件名

	err = v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read conf file err : %v", err)
	}

	return err
}
