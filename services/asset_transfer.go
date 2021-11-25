package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

const (
	AssetTransferOut = iota
	AssetTransferIn
)

type AssetTransferState = int
const (
	AssetTransferInit AssetTransferState = iota
	AssetTransferOngoing
	AssetTransferSuccess
	AssetTransferFailed
)

type AssetTransfer struct {
	db     types.IDB
	config *config.Config
}

func NewAssetTransferService(db types.IDB, conf *config.Config) (p *AssetTransfer, err error) {
	p = &AssetTransfer{
		db:     db,
		config: conf,
	}
	return
}

func (t *AssetTransfer) Name() string {
	return "asset_transfer"
}

func (t *AssetTransfer) Run() (err error) {
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
func getNonce() int {
	return 0
}
func (t *AssetTransfer) handleAssetTransferInit(task *types.AssetTransferTask) (err error) {
	err = utils.CommitWithSession(t.db, func(session *xorm.Session) error {
		params := make([]*types.AssetTransferInParam, 0)
		if err := json.Unmarshal([]byte(task.Params), params); err != nil {
			var txTasks []*types.TransactionTask
			nonce := getNonce()
			for _, param := range params {
				if b, err := json.Marshal(param); err != nil {
					return err
				} else {
					baseTask := &types.BaseTask{State: int(SignState)}
					task := &types.TransactionTask{BaseTask: baseTask, Nonce: nonce, Params: string(b)}
					nonce++
					txTasks = append(txTasks, task)
				}
			}
			if err := t.db.SaveTxTasks(txTasks); err != nil {
				return err
			}
			task.State = int(AssetTransferOngoing)
			return t.db.UpdateAssetTransferTask(task)
		} else {
			return err
		}
	})
	return err
}

type Progress struct {
	AllCount     int
	SuccessCount int
	FailedCount  int
}

func (p *Progress) toString() string {
	return fmt.Sprintf("%d/%d failed:%d", p.SuccessCount, p.AllCount, p.FailedCount)
}

func (t *AssetTransfer) handleAssetTransferOngoing(task *types.AssetTransferTask) (err error) {
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
	return t.db.UpdateAssetTransferTask(task)
}
