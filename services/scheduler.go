package services

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/chainmonitor/config"
	"github.com/starslabhq/chainmonitor/types"
	"os"
	"sync"
	"time"
)

type ServiceScheduler struct {
	conf *config.Config

	services []types.IAsyncService

	closeCh <-chan os.Signal
}

func NewServiceScheduler(conf *config.Config, closeCh <-chan os.Signal) (t *ServiceScheduler, err error) {
	t = &ServiceScheduler{
		conf:     conf,
		closeCh:  closeCh,
		services: make([]types.IAsyncService, 0),
	}

	return
}

func (t *ServiceScheduler) Start() {
	timer := time.NewTimer(t.conf.QueryInterval)
	for {
		select {
		case <-t.closeCh:
			return
		case <-timer.C:

			wg := sync.WaitGroup{}

			for _, service := range t.services {
				wg.Add(1)
				go func(asyncService types.IAsyncService) {
					defer wg.Done()

					err := asyncService.Run()
					if err != nil {
						logrus.Errorf("run service [%v] failed. err:%v", asyncService.Name(), err)
					}
				}(service)
			}

			wg.Wait()

			if !timer.Stop() && len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(t.conf.QueryInterval)
		}
	}
}
