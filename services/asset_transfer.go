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
	AssetTransferIn = iota
	Invest
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
	//TODO
	return 0
}
func (t *AssetTransfer) handleAssetTransferInit(task *types.AssetTransferTask) (err error) {
	err = utils.CommitWithSession(t.db, func(session *xorm.Session) error {
		if task.TransferType == AssetTransferIn {
			err = t.createTransferInTx(session, task)
		} else if task.TransferType == Invest {
			err = t.createInvestTx(session, task)
		}
		if err != nil {
			return err
		}
		task.State = int(AssetTransferOngoing)
		return t.db.UpdateAssetTransferTask(session, task)
	})
	if err != nil{
		logrus.Errorf("handleAssetTransferInit err:%v task:%v", err, task)
	}
	return err
}

func (t *AssetTransfer) createTransferInTx(session *xorm.Session, task *types.AssetTransferTask) (err error) {
	params := make([]*types.AssetTransferInParam, 0)
	err = json.Unmarshal([]byte(task.Params), params)
	if err != nil {
		return err
	}
	var txTasks []*types.TransactionTask
	nonce := getNonce()
	for _, param := range params {
		if b, err := json.Marshal(param); err != nil {
			return err
		} else {
			baseTask := &types.BaseTask{State: int(SignState)}
			task := &types.TransactionTask{BaseTask: baseTask, Nonce: nonce, Params: string(b), TransferType: AssetTransferIn}
			nonce++
			txTasks = append(txTasks, task)
		}
	}
	if err := t.db.SaveTxTasks(session, txTasks); err != nil {
		logrus.Errorf("createTransferInTx save txTasks:%v", err)
	}
	return
}

func (t *AssetTransfer) createInvestTx(session *xorm.Session, task *types.AssetTransferTask) (err error) {
	params := make([]*types.InvestParam, 0)
	err = json.Unmarshal([]byte(task.Params), params)
	if err != nil {
		return err
	}
	var txTasks []*types.TransactionTask
	nonce := getNonce()
	for _, param := range params {
		if b, err := json.Marshal(param); err != nil {
			return err
		} else {
			baseTask := &types.BaseTask{State: int(SignState)}
			task := &types.TransactionTask{BaseTask: baseTask, Nonce: nonce, Params: string(b), TransferType: Invest}
			nonce++
			txTasks = append(txTasks, task)
		}
	}
	if err := t.db.SaveTxTasks(session, txTasks); err != nil {
		logrus.Errorf("createTransferInTx save txTasks:%v", err)
	}
	return
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
	return t.db.UpdateAssetTransferTask(t.db.GetSession(), task)
}
