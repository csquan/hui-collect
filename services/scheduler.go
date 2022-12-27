package services

import (
	"os"
	"sync"
	"time"

	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
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
	//create assembly service
	assemblyService := NewAssemblyService(t.db, t.conf)
	//create sign service
	signService := NewSignService(t.db, t.conf)
	//create boradcast service
	boradcastService := NewBoradcastService(t.db, t.conf)
	//create boradcast service
	checkreceiptService := NewCheckReceiptService(t.db, t.conf)
	//create callback service
	okcallbackService := NewOkCallBackService(t.db, t.conf)

	t.services = []types.IAsyncService{
		assemblyService,
		signService,
		boradcastService,
		checkreceiptService,
		okcallbackService,
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
