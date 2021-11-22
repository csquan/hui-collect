package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/chainmonitor/log"
	"os"
	"time"

	"github.com/starslabhq/chainmonitor/config"
	"net/http"
	_ "net/http/pprof"
)

var (
	conffile string
)

func init() {
	flag.StringVar(&conffile, "conf", "config.yaml", "conf file")
}

func main() {

	flag.Parse()
	fmt.Println(conffile)
	conf, err := config.LoadConf(conffile)
	if err != nil {
		fmt.Println(err.Error())
	}
	if conf.ProfPort != 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", conf.ProfPort), nil)
			if err != nil {
				panic(fmt.Sprintf("start pprof server err:%v", err))
			}
		}()
	}
	err = log.Init(conf.AppName, conf.LogConf)
	if err != nil {
		log.Fatal(err)
	}

	leaseAlive()
	logrus.Info("hermes-rebalance started")
	removeFile()
}

var fName = `/tmp/huobi.lock`

func removeFile() {
	os.Remove(fName)
}

func leaseAlive() {
	f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("create alive file err:%v", err))
	}
	now := time.Now().Unix()
	fmt.Fprintf(f, "%d", now)
}
