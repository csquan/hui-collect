package bridge

import (
	"testing"
	"time"
)

func NoTestFsmRun(t *testing.T) {
	fsm := NewFsm(inited)
	fsm.regEvents(events)
	finished := func(curState state) bool {
		if curState == childsSuc || curState == childsFail || curState == childsPart {
			return true
		}
		return false
	}
	fsm.setFinishState(finished)
	fsm.handleEvent(1)
	fsm.handleEvent(1)
}

func TestTaskFsmRun(t *testing.T) {
	fsm := NewFsm(created)
	fsm.regEvents(taskEvents)

	finished := func(curState state) bool {
		if curState == suc || curState == fail {
			return true
		}
		return false
	}
	fsm.setFinishState(finished)
	fsm.handleEvent(2)
	fsm.handleEvent(2)

}

type rootTask func()

func (rt rootTask) getTasks() []uint64 {
	return []uint64{1}
}

type childTask func()

func (ct childTask) getTasks() []uint64 {
	return []uint64{2, 3}
}

func TestRunTaskManager(t *testing.T) {

	fsm1 := NewFsm(toCreate)
	fsm1.regEvents(taskEvents)

	finished1 := func(curState state) bool {
		if curState == suc || curState == fail {
			return true
		}
		return false
	}
	fsm1.setFinishState(finished1)
	mchild := NewManager(fsm1, fsm1)

	//parent manager
	fsm := NewFsm(inited)
	finished := func(curState state) bool {
		if curState == childsSuc || curState == childsFail || curState == childsPart {
			return true
		}
		return false
	}
	fsm.setFinishState(finished)
	m0 := NewManager(fsm, fsm)
	InitParentEvents(m0)
	m0.child = mchild
	m0.Run()
	m0.HandleTask(1)
	time.Sleep(10 * time.Second)

}
