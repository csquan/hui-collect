package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	remote "github.com/shima-park/agollo/viper-remote"
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

type ApiConf struct {
	MarginUrl    string `mapstructure:"margin_url"`
	MarginOutUrl string `mapstructure:"margin_out_url"`
	LpUrl        string `mapstructure:"lp_url"`
	TaskManager  string `mapstructure:"task_manager"`
}

type BridgeConf struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Ak      string        `mapstructure:"ak"`
	Sk      string        `mapstructure:"sk"`
}

//func (o *DataBaseConf) init() {
//	o.RetryInterval = time.Duration(o.RetryIntervalInt) * time.Millisecond
//}

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
	BridgeConf       BridgeConf            `mapstructure:"bridge_conf"`
	Chains           map[string]*ChainInfo `mapstructure:"chains"`
	ApiConf          ApiConf               `mapstructure:"api"`
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

	if !strings.HasPrefix(strings.ToLower(fpath), "remote") {
		fmt.Println("read configuration from local yaml file :", fpath)
		err := localConfig(fpath, vip)
		if err != nil {
			return nil, err
		}

	} else {
		//has prefix of 'remote', get configuration from remote apollo server
		fmt.Println("read configuration from remote apollo server :", fpath)
		err := remoteConfig(fpath, vip)
		if err != nil {
			return nil, err
		}
		//fmt.Println("config from apollo : ", vip)

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

func getApolloServer() string {
	var apolloSever string
	if Env == envDev {
		apolloSever = apolloSeverDev
	} else {
		apolloSever = apolloSeverProd
	}
	return apolloSever
}

func RemoteSignerConfig(appId string) (conf *Conf) {
	remote.SetAppID(appId)
	v := viper.New()
	v.SetConfigType("yaml")

	err := v.AddRemoteProvider("apollo", getApolloServer(), "config.yaml")
	if err != nil {
		logrus.Errorf("signConfErr add remote provider err:%v", err)
	}
	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("signConfErr read remote conf err:%v", err)
	}

	err = v.WatchRemoteConfigOnChannel() // 启动一个goroutine来同步配置更改
	if err != nil {
		logrus.Errorf("signConfErr watch err:%v", err)
	}
	return &Conf{
		App: appId,
		Vip: v,
	}
}

func remoteConfig(namespace string, v *viper.Viper) error {
	remote.SetAppID(appID)

	err := v.AddRemoteProvider("apollo", getApolloServer(), namespace)
	if err != nil {
		return fmt.Errorf("add remote provider error : %v", err)
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		return fmt.Errorf("read remote config error : %v", err)
	}

	if v.Get("app_name") == nil {
		return errors.New("read remote config error : app_name not found! This namespace might not exist!")
	}

	err = v.WatchRemoteConfigOnChannel() // 启动一个goroutine来同步配置更改

	return err
}

func CheckEnv() bool {
	if Env != envDev && Env != envProd {
		return false
	}
	return true
}
