package part_rebalance

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/types"
)

func TestCheckEvents(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	checker := NewMockEventChecker(ctrl)
	var params []*checkEventParam = []*checkEventParam{
		&checkEventParam{
			ChainID: 56,
			Hash:    "1",
		},
		&checkEventParam{
			ChainID: 128,
			Hash:    "2",
		},
	}
	checker.EXPECT().checkEventHandled(params[0]).Return(true, nil)
	checker.EXPECT().checkEventHandled(params[1]).Return(true, nil)
	i := &investHandler{
		eChecker: checker,
	}
	ok, err := i.checkEventsHandled(params)
	t.Logf("suc ret:%t,err:%v", ok, err)
	if !ok || err != nil {
		t.Fatalf("all suc fail")
	}

	checker.EXPECT().checkEventHandled(params[0]).Return(true, nil)
	checker.EXPECT().checkEventHandled(params[1]).Return(false, nil)
	ok, err = i.checkEventsHandled(params)
	if ok || err != nil {
		t.Fatalf("part suc fail")
	}
	checker.EXPECT().checkEventHandled(params[0]).Return(false, nil)
	ok, err = i.checkEventsHandled(params)
	if ok || err != nil {
		t.Fatalf("all fail err:%v", err)
	}
	checker.EXPECT().checkEventHandled(params[0]).Return(true, nil)
	checker.EXPECT().checkEventHandled(params[1]).Return(true, errors.New("rpc err"))
	ok, err = i.checkEventsHandled(params)
	t.Logf("ret:%t,err:%v", ok, err)
	if ok || err == nil {
		t.Fatalf("all fail fail")
	}
}

func TestInvestMoveToNext(t *testing.T) {
	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		t.Fatalf("new mysql err:%v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	checker := NewMockEventChecker(ctrl)

	checker.EXPECT().checkEventHandled(gomock.Any()).Return(true, nil).AnyTimes()

	i := &investHandler{
		db:       dbtest,
		eChecker: checker,
	}
	err = i.MoveToNextState(&types.PartReBalanceTask{
		Base: &types.Base{
			ID: 2,
		},
		BaseTask: &types.BaseTask{},
	}, types.PartReBalanceSuccess)
	if err != nil {
		t.Fatalf("move to next state err:%v", err)
	}
}
