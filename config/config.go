package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/spf13/viper"
)

type Conf struct {
	App string
	Vip *viper.Viper
}

type DataBaseConf struct {
	DB string `mapstructure:"db"` //DB 连接信息
}

type AlertConf struct {
	URL     string   `mapstructure:"url"`
	Mobiles []string `mapstructure:"mobiles"`
	Secret  string   `mapstructure:"secret"`
}

type ApiConf struct {
	MarginUrl    string `mapstructure:"margin_url"`
	MarginOutUrl string `mapstructure:"margin_out_url"`
	LpUrl        string `mapstructure:"lp_url"`
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

type APIConf struct {
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
	APIConf          APIConf               `mapstructure:"api_conf"`
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

func RemoteSignerConfig(appId string) (conf *Conf) {
	remote.SetAppID(appId)
	v := viper.New()
	v.SetConfigType("yaml")

	v.AddRemoteProvider("apollo", "http://apollo-config.system-service.huobiapps.com", "config.yaml")

	v.ReadRemoteConfig()

	v.WatchRemoteConfigOnChannel() // 启动一个goroutine来同步配置更改

	return &Conf{
		App: appId,
		Vip: v,
	}
}

const (
	appID       string = "rebalance"
	apolloSever string = "http://apollo-config.system-service.huobiapps.com"
)

func remoteConfig(namespace string, v *viper.Viper) error {
	remote.SetAppID(appID)

	err := v.AddRemoteProvider("apollo", apolloSever, namespace)
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
