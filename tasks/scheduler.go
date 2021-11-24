package tasks

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/chainmonitor/config"
	"github.com/starslabhq/chainmonitor/types"
	"os"
	"sync"
	"time"
)

type TaskScheduler struct {
	conf *config.Config

	tasks []types.IAsyncTask

	closeCh <-chan os.Signal
}

func NewTaskScheduler(conf *config.Config, closeCh <-chan os.Signal) (t *TaskScheduler, err error) {
	t = &TaskScheduler{
		conf:    conf,
		closeCh: closeCh,
		tasks:   make([]types.IAsyncTask, 0),
	}

	return
}

func (t *TaskScheduler) Start() {
	timer := time.NewTimer(t.conf.QueryInterval)
	for {
		select {
		case <-t.closeCh:
			return
		case <-timer.C:

			wg := sync.WaitGroup{}

			for _, task := range t.tasks {
				wg.Add(1)
				go func(asyncTask types.IAsyncTask) {
					defer wg.Done()

					err := asyncTask.Run()
					if err != nil {
						logrus.Errorf("run task [%v] failed. err:%v", asyncTask.Name(), err)
					}
				}(task)
			}

			wg.Wait()

			if !timer.Stop() && len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(t.conf.QueryInterval)
		}
	}
}
