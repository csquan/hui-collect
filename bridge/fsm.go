package bridge

import (
	"log"
	"time"
)

type state int

const (
	inited state = iota
	childsAdded

	childsSuc
	childsFail
	childsPart
)

type eventType int

const (
	addChild eventType = iota
	watchChilds
)

type event struct {
	name      string
	curState  state
	eventType eventType
	action    func(taskID uint64) (state, bool)
}

type fsm struct {
	initState     state
	storage       map[uint64]state
	handlers      map[state]*event
	isFinishState func(state) bool
}

var events []*event

func getChildTasks(taskId uint64) []uint64 {
	return []uint64{2, 3}
}

func InitParentEvents(m *Manager) {
	events := []*event{
		&event{
			name:      "task1",
			curState:  inited,
			eventType: addChild,
			action: func(taskID uint64) (state, bool) {
				log.Printf("add childs task")
				m.child.fsm.handleEvent(2)
				m.child.fsm.handleEvent(3)
				return childsAdded, true
			},
		},
		&event{
			name:      "task2",
			curState:  childsAdded,
			eventType: watchChilds,
			action: func(taskID uint64) (state, bool) {
				time.Sleep(time.Second)
				childs := getChildTasks(taskID)
				var errCnt int
				var unknownCnt int
				for _, child := range childs {
					if taskState, ok := m.child.fsm.getCurstate(child); ok {
						if taskState == suc {
							log.Printf("child suc taskid:%d,state:%d", child, taskState)
							continue
						} else if taskState == fail {
							log.Printf("child fail taskid:%d,state:%d", child, taskState)
							errCnt++
						} else {
							unknownCnt++
						}
					}
				}
				if unknownCnt == 0 {
					switch errCnt {
					case 0:
						log.Printf("childs all suc parent:%d", taskID)
						return childsSuc, true
					case len(childs):
						return childsFail, true
					default:
						return childsPart, true
					}
				} else {
					return 0, false
				}
			},
		},
	}
	m.fsm.regEvents(events)
}

func NewFsm(initState state) *fsm {
	fsm := &fsm{
		initState: initState,
		storage:   make(map[uint64]state),
		handlers:  make(map[state]*event),
	}

	return fsm
}

func (f *fsm) regEvents(events []*event) {
	for _, e := range events {
		log.Printf("e.name:%s,key:%d", e.name, e.curState)
		f.handlers[e.curState] = e
	}
}
func (f *fsm) setFinishState(isFinished func(state) bool) {
	f.isFinishState = isFinished
}

func (f *fsm) getCurstate(taskid uint64) (state, bool) {
	ret, ok := f.storage[taskid]
	return ret, ok
}

func (f *fsm) getTasks() []uint64 {
	var tasks []uint64
	for k := range f.storage {
		tasks = append(tasks, k)
	}
	return tasks
}

func (f *fsm) delTask(taskID uint64) {
	delete(f.handlers, state(taskID))
}

func (f *fsm) transfer(taskid uint64, state state) {
	log.Printf("transfer taskid:%d,state:%d", taskid, state)
	f.storage[taskid] = state
}

func (f *fsm) handleEvent(taskID uint64) {
	curState, ok := f.getCurstate(taskID)
	if !ok {
		f.storage[taskID] = f.initState
	}
	if f.isFinishState(curState) {
		return
	}
	if len(f.handlers) == 0 {
		panic("handlers empty")
	}
	e, ok := f.handlers[curState]
	log.Printf("run taksid:%d,curstate:%d,e:%v", taskID, curState, e)
	if ok {
		nextState, ok := e.action(taskID)
		if ok {
			f.transfer(taskID, nextState)
		}
	}
}

type taskGetter interface {
	getTasks() []uint64
}

type Manager struct {
	getter  taskGetter
	fsm     *fsm
	closeCh chan struct{}
	child   *Manager
	taskCh  chan uint64
}

func NewManager(getter taskGetter, fsm *fsm) *Manager {
	m := &Manager{
		getter: getter,
		fsm:    fsm,
		taskCh: make(chan uint64, 1),
	}
	return m
}

func (m *Manager) HandleTask(taskID uint64) bool {
	select {
	case m.taskCh <- taskID:
		return true
	default:
		return false
	}
}

func (m *Manager) Run() {
	go func() {
		for {
			select {
			case <-m.closeCh:
				return
			case task := <-m.taskCh:
				m.fsm.storage[task] = inited
				m.fsm.handleEvent(task)
			default:
				tasks := m.getter.getTasks()
				for _, task := range tasks {
					m.fsm.handleEvent(task)
				}
				time.Sleep(time.Second)
			}
		}
	}()
	if m.child != nil {
		go func() {
			m.child.Run()
		}()
	}
}
