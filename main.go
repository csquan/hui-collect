package main

import (
	"flag"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/api"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/starslabhq/hermes-rebalance/clients"

	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/services"
)

var (
	confFile string
)

func init() {
	flag.StringVar(&confFile, "conf", "config.yaml", "conf file")
}

func main() {
	flag.Parse()
	logrus.Info(confFile)
	conf, err := config.LoadConf(confFile)
	if err != nil {
		logrus.Errorf("load config error:%v", err)
		return
	}

	if conf.ProfPort != 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", conf.ProfPort), nil)
			if err != nil {
				panic(fmt.Sprintf("start pprof server err:%v", err))
			}
		}()
	}

	//setup log print
	err = log.Init(conf.AppName, conf.LogConf)
	if err != nil {
		log.Fatal(err)
	}

	leaseAlive()
	defer removeFile()
	logrus.Info("hermes-rebalance started")

	//setup alert
	err = alert.InitDingding(&conf.Alert)
	if err != nil {
		logrus.Fatalf("set up alert error:%v", err)
	}

	//listen kill signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	//setup db connection
	dbConnection, err := db.NewMysql(&conf.DataBase)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	go api.Run(conf.Port, dbConnection)

	//setup rpc clients
	clients.Init(conf)

	//setup scheduler
	scheduler, err := services.NewServiceScheduler(conf, dbConnection, sigCh)
	if err != nil {
		return
	}
	scheduler.Start()
}

var fName = `/tmp/huobi.lock`

func removeFile() {
	_ = os.Remove(fName)
}

func leaseAlive() {
	f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("create alive file err:%v", err))
	}
	now := time.Now().Unix()
	_, _ = fmt.Fprintf(f, "%d", now)
}
