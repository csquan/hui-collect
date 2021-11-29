package services
//
//import (
//	"encoding/json"
//	"fmt"
//	"math/big"
//	"testing"
//
//	"github.com/starslabhq/hermes-rebalance/config"
//	"github.com/starslabhq/hermes-rebalance/db"
//	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"
//	"github.com/starslabhq/hermes-rebalance/types"
//)
//
//func TestCreateTreansfer(t *testing.T) {
//	var err error
//	dbtest, err := db.NewMysql(&config.DataBaseConf{
//		DB: "root:123456sj@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
//	})
//	if err != nil {
//		panic(fmt.Sprintf("c mysql cli err:%v", err))
//	}
//	params := &types.Params{
//		ReceiveFromBridgeParams: []*types.ReceiveFromBridgeParam{
//			&types.ReceiveFromBridgeParam{
//				ChainId:   1,
//				ChainName: "bsc",
//				From:      "606288c605942f3c84a7794c0b3257b56487263c",
//				To:        "a929022c9107643515f5c777ce9a910f0d1e490c",
//				Amount:    new(big.Int).SetInt64(100),
//				TaskID:    new(big.Int).SetUint64(1),
//			},
//			&types.ReceiveFromBridgeParam{
//				ChainId:   2,
//				ChainName: "poly",
//				From:      "a929022c9107643515f5c777ce9a910f0d1e490c",
//				To:        "a929022c9107643515f5c777ce9a910f0d1e490c",
//				Amount:    new(big.Int).SetInt64(100),
//				TaskID:    new(big.Int).SetUint64(2),
//			},
//		},
//	}
//	data, err := json.Marshal(params)
//
//	task := &types.PartReBalanceTask{
//		BaseTask: &types.BaseTask{State: types.PartReBalanceCross, Message: ""},
//		Base:     &types.Base{ID: 1},
//		Params:   string(data),
//	}
//
//	part_rebalance.CreateReceiveFromBridgeTask(task, dbtest, dbtest.GetSession())

//}
