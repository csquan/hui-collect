package utils

import (
	"reflect"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func HandleErrorWithRetryMaxTime(handle func() error, retryTimes int, interval time.Duration) error {
	inc := 0
	for {
		err := handle()
		if err == nil {
			return nil
		}
		logrus.Warnf("%v handle error with retry: %v", runtime.FuncForPC(reflect.ValueOf(handle).Pointer()).Name(), err)

		inc++
		if inc > retryTimes {
			return err
		}

		time.Sleep(interval)
	}
}
