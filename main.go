package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/chainmonitor/log"
	"github.com/starslabhq/chainmonitor/tasks"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/starslabhq/chainmonitor/config"
	"net/http"
	_ "net/http/pprof"
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
	err = log.Init(conf.AppName, conf.LogConf)
	if err != nil {
		log.Fatal(err)
	}

	leaseAlive()
	defer removeFile()
	logrus.Info("hermes-rebalance started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	scheduler, err := tasks.NewTaskScheduler(conf, sigCh)
	if err != nil {
		return
	}

	scheduler.Start()
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
