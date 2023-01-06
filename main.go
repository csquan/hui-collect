package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/db"
	"github.com/ethereum/Hui-TxState/log"
	"github.com/ethereum/Hui-TxState/services"
	"github.com/sirupsen/logrus"
)

var (
	confFile string
)

func init() {
	flag.StringVar(&confFile, "conf", "config.yaml", "conf file")
	flag.StringVar(&config.Env, "env", "dev", "env")
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
	err = log.Init(conf.AppName, conf.LogConf, conf.Env)
	if err != nil {
		log.Fatal(err)
	}

	leaseAlive()
	defer removeFile()
	logrus.Info("Hui-Collect started")

	//listen kill signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	//setup db connection
	dbConnection, err := db.NewMysql(&conf.DataBase)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	block_dbConnection, err := db.NewBlockMysql(&conf.Monitor)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	//setup scheduler
	scheduler, err := services.NewServiceScheduler(conf, dbConnection, block_dbConnection, sigCh)
	if err != nil {
		return
	}
	scheduler.Start()
}

var fName = `/tmp/hui.lock`

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
