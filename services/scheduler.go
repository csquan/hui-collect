package services

import (
	"os"
	"sync"
	"time"

	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/sirupsen/logrus"
)

type ServiceScheduler struct {
	conf *config.Config

	db types.IDB

	services []types.IAsyncService

	closeCh <-chan os.Signal
}

func NewServiceScheduler(conf *config.Config, db types.IDB, closeCh <-chan os.Signal) (t *ServiceScheduler, err error) {
	t = &ServiceScheduler{
		conf:     conf,
		closeCh:  closeCh,
		db:       db,
		services: make([]types.IAsyncService, 0),
	}

	return
}

func (t *ServiceScheduler) Start() {
	//create collect service
	collectService := NewCollectService(t.db, t.conf)

	//create update service
	CheckService := NewCheckService(t.db, t.conf)

	t.services = []types.IAsyncService{
		collectService,
		CheckService,
	}

	timer := time.NewTimer(t.conf.QueryInterval)
	for {
		select {
		case <-t.closeCh:
			return
		case <-timer.C:

			wg := sync.WaitGroup{}

			for _, s := range t.services {
				wg.Add(1)
				go func(asyncService types.IAsyncService) {
					defer wg.Done()
					defer func(start time.Time) {
						//logrus.Infof("%v task process cost %v", asyncService.Name(), time.Since(start))
					}(time.Now())

					err := asyncService.Run()
					if err != nil {
						logrus.Errorf("run s [%v] failed. err:%v", asyncService.Name(), err)
					}
				}(s)
			}

			wg.Wait()

			if !timer.Stop() && len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(t.conf.QueryInterval)
		}
	}
}
