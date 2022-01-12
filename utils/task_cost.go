package utils

import (
	"fmt"
	"time"
)

var fullReCost *Cost
var partReCost *Cost

type Cost struct {
	Start  int64
	Report string
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
	now := time.Now().Unix()
	cost := now - f.Start
	f.Start = now
	f.Report = fmt.Sprintf(`%s
- %s:%ds`, f.Report, step, cost)
}
