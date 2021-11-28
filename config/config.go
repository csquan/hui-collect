package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"

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
	url string `mapstructure:"url"`
}

//func (o *DataBaseConf) init() {
//	o.RetryInterval = time.Duration(o.RetryIntervalInt) * time.Millisecond
//}

type Config struct {
	AppName          string `mapstructure:"app_name"`
	ProfPort         int    `mapstructure:"prof_port"`
	QueryInterval    time.Duration
	QueryIntervalInt uint64                `mapstructure:"query_interval"` //ms
	DataBase         DataBaseConf          `mapstructure:"database"`
	LogConf          Log                   `mapstructure:"log"`
	Alert            AlertConf             `mapstructure:"alert"`
	Chains           map[string]*ChainInfo `mapstructure:"chains"`
	ClientMap        map[string]*ethclient.Client
}

type ChainInfo struct {
	RpcUrl  string `mapstructure:"rpc_url"`
	Timeout int    `mapstructure:"timeout"`
}

func (c *Config) init() {
	c.QueryInterval = time.Duration(c.QueryIntervalInt) * time.Millisecond
	c.ClientMap = make(map[string]*ethclient.Client)
	for k, chain := range c.Chains {
		client, err := rpc.DialHTTPWithClient(chain.RpcUrl, &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
			Timeout: time.Duration(chain.Timeout) * time.Millisecond,
		})
		if err != nil {
			logrus.Fatalf("config init err:%v", err)
		}
		c.ClientMap[k] = ethclient.NewClient(client)
	}
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
	appID       string = "hermes-rebalance"
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
