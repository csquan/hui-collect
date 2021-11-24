package services

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type AssetTransferState int

const (
	AssetTransferInit AssetTransferState = iota
	AssetTransferOngoing
	AssetTransferSuccess
	AssetTransferFailed
)

type Transfer struct {
	db     types.IDB
	config *config.Config
}

func NewAssetTransferService(db types.IDB, conf *config.Config) (p *Transfer, err error) {
	p = &Transfer{
		db:     db,
		config: conf,
	}
	return
}

func (t *Transfer) Name() string {
	return "transfer"
}

func (t *Transfer) Run() (err error) {
	tasks, err := t.db.GetOpenedAssetTransferTasks()
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		logrus.Infof("no available transfer task.")
		return
	}

	if len(tasks) > 1 {
		logrus.Errorf("more than one transfer services are being processed. tasks:%v", tasks)
	}

	switch AssetTransferState(tasks[0].State) {
	case AssetTransferInit:
		return t.handleAssetTransferInit(tasks[0])
	case AssetTransferOngoing:
		return t.handleAssetTransferOngoing(tasks[0])
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
	}

	return
}

func (t *Transfer) handleAssetTransferInit(task *types.AssetTransferTask) (err error) {
	//解析params 创建txTasks自任务，切换到TransferOngoing
	//TODO 放在事物中
	var txTasks []*types.TransactionTask
	if err := t.db.SaveTxTasks(txTasks); err != nil {
		return err
	}
	task.State = int(AssetTransferOngoing)
	return t.db.UpdateTransferTask(task)
}

type Progress struct {
	AllCount     int
	SuccessCount int
	FailedCount  int
}

func (p *Progress) toString() string {
	return fmt.Sprintf("%d/%d failed:%d", p.SuccessCount, p.AllCount, p.FailedCount)
}

func (t *Transfer) handleAssetTransferOngoing(task *types.AssetTransferTask) (err error) {
	//扫描子txTasks，更新状态
	txTasks, err := t.db.GetTxTasks(task.ID)
	if err != nil {
		return
	}

	progress := &Progress{
		AllCount: len(txTasks),
	}
	for _, tx := range txTasks {
		if tx.State == int(TxSuccess) {
			progress.SuccessCount++
		}
		if tx.State == int(TxFailed) {
			progress.FailedCount++
			task.State = int(AssetTransferFailed)
		}
	}
	if progress.SuccessCount == progress.AllCount {
		task.State = int(AssetTransferSuccess)
	}
	task.Progress = progress.toString()
	return t.db.UpdateTransferTask(task)
}
