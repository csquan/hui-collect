package tasks

import (
	"github.com/starslabhq/chainmonitor/config"
	"github.com/starslabhq/chainmonitor/types"
)

type PartReBalanceState int

const (
	Init PartReBalanceState = iota
	Cross
	CrossDone
	TransferIn
	TransferInDone
	Farm
	Success
	Failed
)

var partReBalanceNames = map[PartReBalanceState]string{
	Init:           "init",
	Cross:          "cross",
	CrossDone:      "crossDone",
	TransferIn:     "transferIn",
	TransferInDone: "transferInDone",
	Farm:           "farm",
	Success:        "success",
	Failed:         "failed",
}

type PartReBalance struct {
	db      types.IDB
	config  *config.Config
}

func NewPartReBalanceService(db types.IDB, conf *config.Config) (p *PartReBalance, err error) {
	p = &PartReBalance{
		db:      db,
		config:  conf,
	}

	return
}


func (p *PartReBalance) Create() (err error) {

	//// already have a task in processing
	//if p.fsm != nil {
	//	return
	//}
	//check db to get task info
	//tasks, err := p.db.GetOpenedPartReBalanceTasks()
	//if err != nil {
	//	return err
	//}
	//
	//if len(tasks) == 0 {
	//	return
	//}
	////TODO make sure which to pick out
	//sort.SliceStable(tasks, func(i, j int) bool {
	//	return tasks[i].State <= tasks[j].State
	//})

	//tasks := []*types.PartReBalanceTask {
	//	&types.PartReBalanceTask{
	//		Base: &types.Base{
	//
	//		},
	//		BaseTask: &types.BaseTask{
	//
	//		},
	//	},
	//}

	//pick one and create fsm

	//p.fsm = fsm.NewFSM(partReBalanceNames[PartReBalanceState(tasks[0].State)],
	//	fsm.Events{
	//		{Name: "crossing", Src: []string{partReBalanceNames[Init]}, Dst: partReBalanceNames[Cross]},
	//		{Name: "crossed", Src: []string{partReBalanceNames[Cross]}, Dst: partReBalanceNames[CrossDone]},
	//		{Name: "transferOuting", Src: []string{partReBalanceNames[CrossDone]}, Dst: partReBalanceNames[TransferIn]},
	//		{Name: "transferOuted", Src: []string{partReBalanceNames[TransferIn]}, Dst: partReBalanceNames[TransferInDone]},
	//		{Name: "farming", Src: []string{partReBalanceNames[TransferInDone]}, Dst: partReBalanceNames[Farm]},
	//		{Name: "success", Src: []string{partReBalanceNames[Farm]}, Dst: partReBalanceNames[Success]},
	//		{Name: "failed", Src: []string{partReBalanceNames[Cross], partReBalanceNames[TransferIn], partReBalanceNames[Farm]}, Dst: partReBalanceNames[Failed]},
	//	},
	//	fsm.Callbacks{
	//		"leave_state": func(event *fsm.Event) {
	//			logrus.Infof("part rebalance leave state:%v", event.Src)
	//		},
	//		"enter_state": func(event *fsm.Event) {
	//			logrus.Infof("part rebalance enter state:%v", event.Dst)
	//		},
	//		"leave_init": func(event *fsm.Event) {
	//			logrus.Infof("leave init state:%v", event.Event)
	//			//create cross task and change to next state
	//			//event.Cancel(err)
	//			event.Async()
	//		},
	//	})

	return
}

//Push push state machine to next state
func (p *PartReBalance) Push() (err error) {
	//if p.fsm == nil {
	//	return
	//}

	//switch p.fsm.Current() {
	//case partReBalanceNames[Init]:
	//	//TODO create cross task and change to next state
	//	err = p.fsm.Event("crossing")
	//	if err != nil {
	//		logrus.Errorf("init  to crossing failed. err:%v", err)
	//	}
	//	err = p.fsm.Transition()
	//	if err != nil {
	//		logrus.Errorf("init  to crossing failed. err:%v", err)
	//	}
	//	return
	//case partReBalanceNames[Cross]:
	//	//TODO query cross task state
	//case partReBalanceNames[CrossDone]:
	//case partReBalanceNames[TransferIn]:
	//
	//case partReBalanceNames[TransferInDone]:
	//
	//case partReBalanceNames[Farm]:
	//case partReBalanceNames[Success]:
	//
	//case partReBalanceNames[Failed]:
	//}

	return
}
