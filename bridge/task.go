package bridge

import "log"

type taskState int

const (
	toCreate state = iota
	created

	suc
	fail
)

const (
	create eventType = iota
	watch
)

var taskEvents []*event

func init() {
	taskEvents = []*event{
		&event{
			name:      "task_create",
			curState:  toCreate,
			eventType: create,
			action: func(taskID uint64) (state, bool) {
				log.Printf("created taskid:%d", taskID)
				return created, true
			},
		},
		&event{
			name:      "task_watch",
			curState:  created,
			eventType: watch,
			action: func(taskID uint64) (state, bool) {
				log.Printf("suc taskid:%d", taskID)
				return suc, true
			},
		},
	}
}
