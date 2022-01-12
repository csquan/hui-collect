package utils

import (
	"github.com/sirupsen/logrus"
	"time"
)

func CostLog(start int64, taskID uint64, step string) {
	cost := time.Now().Unix() - start
	logrus.Infof("task cost:%d step:%s taskID:%d", cost, step, taskID)
}