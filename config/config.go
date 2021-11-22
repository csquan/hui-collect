package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/spf13/viper"
)

type DataBaseConf struct {
	DB               string `mapstructure:"db"`             //DB 连接信息
	SqlBatch         int    `mapstructure:"sql_batch"`      //  批量插入sql时，每次批量插入的最大值
	RetryIntervalInt int    `mapstructure:"retry_interval"` // 存储失败时重试的间隔 ms
	RetryTimes       int    `mapstructure:"retry_times"`    // 存储失败，重试次数
	RetryInterval    time.Duration
}

func (o *DataBaseConf) init() {
	o.RetryInterval = time.Duration(o.RetryIntervalInt) * time.Millisecond
}



type Config struct {
	ProfPort int    `mapstructure:"prof_port"`
	AppName  string `mapstructure:"app_name"`
	DataBase DataBaseConf `mapstructure:"output"`
	LogConf  Log          `mapstructure:"log"`
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
		DataBase:       DefaultDataBaseConfig,
		LogConf:      DefaultLogConfig,
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

	conf.DataBase.init()
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
		return fmt.Errorf("read remote config error : app_name not found! This namespace might not exsit!")
	}

	err = v.WatchRemoteConfigOnChannel() // 启动一个goroutine来同步配置更改

	return err
}
