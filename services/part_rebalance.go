package services

import (
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
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
	case types.Init:
		return p.handleInit(tasks[0])
	case types.Cross:
		return p.handleCross(tasks[0])
	case types.TransferIn:
		return p.handleTransferIn(tasks[0])
	case types.Farm:
		return p.handleFarm(tasks[0])
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
	}

	return
}

func (p *PartReBalance) handleInit(task *types.PartReBalanceTask) (err error) {
	return
}

func (p *PartReBalance) handleCross(task *types.PartReBalanceTask) (err error) {
	return
}

func (p *PartReBalance) handleTransferIn(task *types.PartReBalanceTask) (err error) {
	return
}

func (p *PartReBalance) handleFarm(task *types.PartReBalanceTask) (err error) {
	return
}
