package services

import (
	"os"
	"sync"
	"time"

	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
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
	partReBalance, err := part_rebalance.NewPartReBalanceService(t.db, t.conf)
	if err != nil {
		logrus.Fatalf("new rebalance service error: %v", err)
	}
	t.services = append(t.services, partReBalance)

	transaction, err := NewTransactionService(t.db, t.conf)
	if err != nil {
		logrus.Fatalf("new transfer service error: %v", err)
	}

	//create cross service
	bridgeConf := t.conf.BridgeConf
	bridgeCli, err := bridge.NewBridge(bridgeConf.URL, bridgeConf.Ak, bridgeConf.Sk, bridgeConf.Timeout)
	if err != nil {
		logrus.Fatalf("new bridge cli err:%v", err)
	}
	crossService := NewCrossService(t.db, bridgeCli, t.conf)
	crossSubService := NewCrossSubTaskService(t.db, bridgeCli, t.conf)
	//create cross service

	t.services = []types.IAsyncService{
		partReBalance,
		transaction,
		crossService,
		crossSubService,
	}

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
