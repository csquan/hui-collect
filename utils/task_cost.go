package utils

import (
	"fmt"
	"time"
)

var fullReCost *Cost
var partReCost *Cost

type Cost struct {
	Start    int64
	Report   string
	lastStep string
}

func GetFullReCost(taskID uint64) *Cost {
	if fullReCost == nil {
		InitFullReCost(taskID)
	}
	return fullReCost
}
func GetPartReCost(taskID uint64) *Cost {
	if partReCost == nil {
		InitPartReCost(taskID)
	}
	return partReCost
}

func InitFullReCost(taskID uint64) {
	fullReCost = &Cost{Start: time.Now().Unix(), Report: fmt.Sprintf("#### 大Re耗时 taskID:%d\n", taskID)}
}

func InitPartReCost(taskID uint64) {
	partReCost = &Cost{Start: time.Now().Unix(), Report: fmt.Sprintf("#### 小Re耗时 taskID:%d\n", taskID)}
}

func (f *Cost) AppendReport(step string) {
	if step == f.lastStep { //当checkFinish成功之后，moveToNextState不一定成功，有可能多次执行checkFinish，这种情况不做处理，只记录第一次的时间
		return
	}
	now := time.Now().Unix()
	cost := now - f.Start
	f.Start = now
	f.Report = fmt.Sprintf(`%s
- %s:%ds`, f.Report, step, cost)
	f.lastStep = step
}
