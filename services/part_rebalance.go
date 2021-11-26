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

type PartReBalance struct {
	db     types.IDB
	config *config.Config
}

func NewPartReBalanceService(db types.IDB, conf *config.Config) (p *PartReBalance, err error) {
	p = &PartReBalance{
		db:     db,
		config: conf,
	}

	return
}

func (p *PartReBalance) Name() string {
	return "part_rebalance"
}

func (p *PartReBalance) Run() (err error) {
	tasks, err := p.db.GetOpenedPartReBalanceTasks()
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		logrus.Infof("no available part rebalance task.")
		return
	}

	if len(tasks) > 1 {
		logrus.Errorf("more than one rebalance tasks are being processed. tasks:%v", tasks)
	}

	switch tasks[0].State {
	case types.PartReBalanceInit:
		return p.handleInit(tasks[0])
	case types.PartReBalanceCross:
		return p.handleCross(tasks[0])
	case types.PartReBalanceTransferIn:
		return p.handleTransferIn(tasks[0])
	case types.PartReBalanceInvest:
		return p.handleInvest(tasks[0])
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
	}

	return
}

//handleInit find out init state and create cross task
func (p *PartReBalance) handleInit(task *types.PartReBalanceTask) (err error) {
	crossBalances := make([]*types.CrossBalanceItem, 0)
	err = json.Unmarshal([]byte(task.Params), &crossBalances)
	if err != nil {
		logrus.Errorf("read task [%v] cross params error:%v", task, err)
		return
	}

	if len(crossBalances) == 0 {
		logrus.Errorf("no cross balance is found for rebalance task: [%v]", task)
		return
	}

	crossTasks := make([]*types.CrossTask, 0, len(crossBalances))
	for _, param := range crossBalances {
		crossTasks = append(crossTasks, &types.CrossTask{
			BaseTask:      &types.BaseTask{State: types.ToCreateSubTask},
			RebalanceId:   task.ID,
			ChainFrom:     param.FromChain,
			ChainTo:       param.ToChain,
			ChainFromAddr: param.FromAddr,
			ChainToAddr:   param.ToAddr,
			CurrencyFrom:  param.FromCurrency,
			CurrencyTo:    param.ToCurrency,
			Amount:        param.Amount,
		})
	}

	err = utils.CommitWithSession(p.db, func(session *xorm.Session) (execErr error) {
		execErr = p.db.SaveCrossTasks(session, crossTasks)
		if execErr != nil {
			logrus.Errorf("save cross task error:%v task:[%v]", execErr, task)
			return
		}

		task.State = types.PartReBalanceCross
		execErr = p.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}

		return
	})

	return
}

//handleCross check cross task state and create transfer in task when cross finished
func (p *PartReBalance) handleCross(task *types.PartReBalanceTask) (err error) {

	crossTasks, err := p.db.GetCrossTasksByReBalanceId(task.ID)
	if err != nil {
		logrus.Errorf("get cross task for rebalance [%v] failed", task)
	}

	if len(crossTasks) == 0 {
		err = fmt.Errorf("part rebalance task [%v] has no cross task", task)
		return
	}

	success := true
	for _, crossTask := range crossTasks {
		if crossTask.State != types.TaskSuc {
			logrus.Debugf("cross task [%v] is not finished", crossTask)

			return
		}

		success = success && crossTask.State == types.TaskSuc
	}

	err = utils.CommitWithSession(p.db, func(session *xorm.Session) (execErr error) {

		//create next state task
		if success {
			var assetTransfer *types.AssetTransferTask
			assetTransfer, execErr = p.createTransferInTask(task)
			if execErr != nil {
				return
			}

			execErr = p.db.SaveAssetTransferTask(session, assetTransfer)
			if execErr != nil {
				logrus.Errorf("save assetTransfer task error:%v task:[%v]", execErr, task)
				return
			}
		}

		//move to next state
		if success {
			task.State = types.PartReBalanceTransferIn
		} else {
			task.State = types.PartReBalanceFailed
		}
		execErr = p.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}

func (p *PartReBalance) createTransferInTask(task *types.PartReBalanceTask) (assetTransfer *types.AssetTransferTask, err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}

	assetTransferParams, err := json.Marshal(params.AssetTransferIn)
	if err != nil {
		logrus.Errorf("marshal AssetTransferInParams params error:%v task:[%v]", err, task)
		return
	}

	assetTransfer = &types.AssetTransferTask{
		BaseTask:     &types.BaseTask{State: types.AssetTransferInit},
		RebalanceId:  task.ID,
		TransferType: types.AssetTransferIn,
		Params:       string(assetTransferParams),
	}

	return
}

func (p *PartReBalance) createInvestTask(task *types.PartReBalanceTask) (assetTransfer *types.AssetTransferTask, err error) {
	params, err := task.ReadParams()
	if err != nil {
		return
	}

	investParams, err := json.Marshal(params.Invest)
	if err != nil {
		logrus.Errorf("marshal AssetTransferInParams params error:%v task:[%v]", err, task)
		return
	}

	assetTransfer = &types.AssetTransferTask{
		BaseTask:     &types.BaseTask{State: types.AssetTransferInit},
		RebalanceId:  task.ID,
		TransferType: types.Invest,
		Params:       string(investParams),
	}

	return
}

// handleTransferIn check transferIn task state and create invest task after transferIn finished
func (p *PartReBalance) handleTransferIn(task *types.PartReBalanceTask) (err error) {

	state, err := p.getTransferState(task, types.AssetTransferIn)
	if err != nil {
		return err
	}

	err = utils.CommitWithSession(p.db, func(session *xorm.Session) (execErr error) {
		if state == types.AssetTransferSuccess {
			var invest *types.AssetTransferTask

			invest, execErr = p.createInvestTask(task)
			if execErr != nil {
				return
			}

			execErr = p.db.SaveAssetTransferTask(session, invest)
			if execErr != nil {
				logrus.Errorf("save invest task error:%v task:[%v]", execErr, task)
				return
			}
			task.State = types.PartReBalanceInvest
		} else if state == types.AssetTransferFailed {
			//TODO
			task.State = types.PartReBalanceFailed
		}
		execErr = p.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}

//getTransferSatae
func (p *PartReBalance) getTransferState(task *types.PartReBalanceTask, transferType int) (state int, err error) {
	atTasks, err := p.db.GetAssetTransferTasksWithReBalanceId(task.ID, transferType)
	if err != nil {
		logrus.Errorf("get asset transfer task error:%v", err)
		return
	}

	if len(atTasks) == 0 {
		err = fmt.Errorf("part rebalance task [%v] has no transfer in task", task)
		return
	}
	success := true
	for _, at := range atTasks {
		if at.State == types.AssetTransferFailed {
			state = types.AssetTransferFailed
			return
		}
		success = success && (at.State == types.AssetTransferSuccess)
	}
	if success {
		state = types.AssetTransferSuccess
	} else {
		state = types.AssetTransferOngoing
	}
	return
}

func (p *PartReBalance) handleInvest(task *types.PartReBalanceTask) (err error) {
	state, err := p.getTransferState(task, types.Invest)
	if err != nil {
		return err
	}

	err = utils.CommitWithSession(p.db, func(session *xorm.Session) (execErr error) {
		if state == types.AssetTransferSuccess {
			var invest *types.AssetTransferTask
			invest, execErr = p.createInvestTask(task)
			if execErr != nil {
				return
			}

			execErr = p.db.SaveAssetTransferTask(session, invest)
			if execErr != nil {
				logrus.Errorf("save invest task error:%v task:[%v]", execErr, task)
				return
			}
			task.State = types.PartReBalanceSuccess
		} else if state == types.AssetTransferFailed {
			//TODO
			task.State = types.PartReBalanceFailed
		}
		execErr = p.db.UpdatePartReBalanceTask(session, task)
		if execErr != nil {
			logrus.Errorf("update part rebalance task error:%v task:[%v]", execErr, task)
			return
		}
		return
	})

	return
}
