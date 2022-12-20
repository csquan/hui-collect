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
	appID           string = "rebalance"
	apolloSeverDev  string = "http://apollo-config.system-service.huobiapps.com"
	apolloSeverProd string = "http://apollo-config.system-service.apne-1.huobiapps.com:80" //prod
	envProd                = "prod"
	envDev                 = "dev"
)

type Conf struct {
	App string
	Vip *viper.Viper
}

type DataBaseConf struct {
	DB string `mapstructure:"db"` //DB 连接信息
}

type AlertConf struct {
	URL         string   `mapstructure:"url"`
	Mobiles     []string `mapstructure:"mobiles"`
	Secret      string   `mapstructure:"secret"`
	MaxWaitTime int64    `mapstructure:"max_wait_time"`
}

type ServerConf struct {
	Port  int               `mapstructure:"port"`
	Users map[string]string `mapstructure:"users"`
}
type Config struct {
	AppName          string `mapstructure:"app_name"`
	ProfPort         int    `mapstructure:"prof_port"`
	QueryInterval    time.Duration
	QueryIntervalInt uint64                `mapstructure:"query_interval"` //ms
	DataBase         DataBaseConf          `mapstructure:"database"`
	LogConf          Log                   `mapstructure:"log"`
	Alert            AlertConf             `mapstructure:"alert"`
	Chains           map[string]*ChainInfo `mapstructure:"chains"`
	Margin           *Margin               `mapstructure:"margin"`
	Env              string                `mapstructure:"env"`
	ServerConf       ServerConf            `mapstructure:"server_conf"`
	IsCheckParams    bool                  `mapstructure:"is_check_params"`
}
type Margin struct {
	AppID     string `mapstructure:"app_id"`
	SecretKey string `mapstructure:"secret_key"`
}

type ChainInfo struct {
	ID            int    `mapstructure:"id"`
	RpcUrl        string `mapstructure:"rpc_url"`
	Timeout       int    `mapstructure:"timeout"`
	BridgeAddress string `mapstructure:"bridge_address"`
	BlockSafe     uint   `mapstructure:"block_safe"`
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
		ServerConf:       DefaultServerConf,
		Alert:            DefaultAlertConf,
	}

	vip := viper.New()
	vip.SetConfigType("yaml")

	if strings.HasPrefix(strings.ToLower(fpath), "remote") {
		fmt.Println("read configuration from local yaml file :", fpath)
		err := localConfig(fpath, vip)
		if err != nil {
			return nil, err
		}

	}
	err := vip.Unmarshal(conf)
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
