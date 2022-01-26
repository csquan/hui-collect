package services

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/bridge"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type CrossSubTaskService struct {
	db        types.IDB
	bridgeCli bridge.IBridge
	config    *config.Config
}

func NewCrossSubTaskService(db types.IDB, bCli bridge.IBridge, c *config.Config) *CrossSubTaskService {
	return &CrossSubTaskService{
		db:        db,
		bridgeCli: bCli,
		config:    c,
	}
}

func (c *CrossSubTaskService) Run() error {
	parentTasks, err := c.db.GetOpenedCrossTasks()
	if err != nil {
		return err
	}

	for _, pt := range parentTasks {
		childTasks, err := c.db.GetOpenedCrossSubTasks(pt.ID)
		if err != nil {
			logrus.Errorf("get opened sub tasks err:%v,parent:%d", err, pt.ID)
			continue
		}
		if len(childTasks) == 0 {
			logrus.Infof("get opened sub tasks empty parent:%d", pt.ID)
			continue
		}
		bridgeId, err := getBridgeID(c.bridgeCli, pt)
		if err != nil {
			logrus.Errorf("get bridgeId err:%v,taskId:%d", bridgeId, pt.ID)
			continue
		}
		for _, subTask := range childTasks {
			switch types.CrossSubState(subTask.State) {
			case types.ToCross:
				//do cross
				t := &bridge.Task{
					TaskNo:         subTask.TaskNo,
					FromAccountId:  bridgeId.fromAccountId,
					ToAccountId:    bridgeId.toAccountId,
					FromCurrencyId: bridgeId.fromCurrencyId,
					ToCurrencyId:   bridgeId.toCurrencyId,
					Amount:         subTask.Amount,
				}
				taskId, err := c.bridgeCli.AddTask(t)
				if err == nil && taskId != 0 {
					//update bridge taskId
					err1 := c.db.UpdateCrossSubTaskBridgeIDAndState(subTask.ID, taskId, int(types.Crossing))
					if err1 != nil {
						logrus.Errorf("sub task state:%d,id:%d,err1:%v", subTask.State, subTask.ID, err1)
						return fmt.Errorf("sub task state:%d,id:%d,err1:%v", subTask.State, subTask.ID, err1)
					}

					subTask.BridgeTaskId = taskId
					c.stateChanged(types.Crossing, &CrossSubInfo{
						Parent:         pt,
						Sub:            subTask,
						FromAccountId:  t.FromAccountId,
						ToAccountId:    t.ToAccountId,
						FromCurrencyId: t.FromCurrencyId,
						ToCurrencyId:   t.ToCurrencyId,
						Amount:         t.Amount,
					})

				} else {
					logrus.Warnf("add task fail err:%v,taskNo:%d,subTaskId:%d,parent:%d", err, t.TaskNo, subTask.ID, pt.ID)
				}
			case types.Crossing:
				//watch
				bridgeTask, err := c.bridgeCli.GetTaskDetail(subTask.BridgeTaskId)
				if err != nil {
					logrus.Errorf("get task detail err:%v,bridgeId:%d,sub:%d", err, subTask.BridgeTaskId, subTask.ID)
					continue
				}
				switch bridgeTask.Status {
				case 0:
					logrus.Infof("bridge task wait to start bridgeTaskId:%d", subTask.BridgeTaskId)
				case 1:
					logrus.Infof("bridge task crossing bridgeTaskId:%d", subTask.BridgeTaskId)
				case 2:
					logrus.Infof("bridge task crossed bridgeTaskId:%d", subTask.BridgeTaskId)
					err = c.db.UpdateCrossSubTaskState(subTask.ID, int(types.Crossed))
					if err != nil {
						continue
					}
					c.stateChanged(types.Crossing, &CrossSubInfo{
						Parent:         pt,
						Sub:            subTask,
						FromAccountId:  bridgeTask.FromAccountId,
						ToAccountId:    bridgeTask.ToAccountId,
						FromCurrencyId: bridgeId.fromCurrencyId,
						ToCurrencyId:   bridgeId.toCurrencyId,
					})
				default:
					logrus.Fatalf("unexpected bridge task state subtaskId:%d,bridge:%d,state:%d,parent:%d",
						subTask.ID, subTask.BridgeTaskId, subTask.State, pt.ID)
				}
			default:
				logrus.Fatalf("unexpected sub task state sub_task id:%d,state:%d", subTask.ID, subTask.State)
			}
		}
	}
	return nil
}

func (c *CrossSubTaskService) Name() string {
	return "cross_sub"
}

type CrossSubInfo struct {
	Stage          string
	Parent         *types.CrossTask
	Sub            *types.CrossSubTask
	FromAccountId  uint64
	ToAccountId    uint64
	FromCurrencyId int
	ToCurrencyId   int
	Amount         string
}

const crossSubTemp = `# stage: {{.Stage}}
## parent
{{with .Parent}}
- chain: {{.ChainFrom}} -> {{.ChainTo}}
- addr: {{.ChainFromAddr}}->{{.ChainToAddr}}
- currency: {{.CurrencyFrom}}->{{.CurrencyTo}}
- amount: {{.Amount}}
{{end}}
## sub
{{with .Sub}}
- bridge_task_id: {{.BridgeTaskId}}
- task_no: {{.TaskNo}}
- amount: {{.Amount}}
- state: {{.State}}
{{end}}
## bridge

- from_account: {{.FromAccountId}}
- to_account: {{.ToAccountId}}
- from_currency: {{.FromCurrencyId}}
- to_currency: {{.ToCurrencyId}}
`

func createCrossSubMsg(stage string, info *CrossSubInfo) (string, error) {
	info.Stage = stage
	t := template.New("cross_sub_msg")
	temp := template.Must(t.Parse(crossSubTemp))
	buf := &bytes.Buffer{}
	err := temp.Execute(buf, info)
	if err != nil {
		return "", fmt.Errorf("excute temp err:%v", err)
	}
	return buf.String(), nil
}

func (c *CrossSubTaskService) stateChanged(next types.CrossSubState, info *CrossSubInfo) {
	var (
		msg string
		err error
	)
	switch next {
	case types.Crossing:
		msg, err = createCrossSubMsg("sub_crossing", info)
	case types.Crossed:
		msg, err = createCrossSubMsg("sub_crossed", info)
	}
	if err != nil {
		logrus.Errorf("create cross sub msg err:%v,msg:%s", err, msg)
		return
	}
	err = alert.Dingding.SendMessage("cross_sub", msg)
	if err != nil {
		logrus.Errorf("send cross_sub msg err:%v,msg:%s", err, msg)
	}
}
