package services

import (
	"encoding/json"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/services/part_rebalance"
	"github.com/starslabhq/hermes-rebalance/types"
	"math/big"
	"testing"
)

func TestCreateTreansfer(t *testing.T) {
	var err error
	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "root:123456sj@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		panic(fmt.Sprintf("c mysql cli err:%v", err))
	}
	params := &types.Params{
		ReceiveFromBridgeParams: []*types.ReceiveFromBridgeParam{
			&types.ReceiveFromBridgeParam{
				ChainId:   1,
				ChainName: "bsc",
				From:      "0x0000000000000",
				To:        "0x0000000000000",
				Amount:    new(big.Int).SetInt64(100),
				TaskID:    new(big.Int).SetUint64(1),
			},
			&types.ReceiveFromBridgeParam{
				ChainId:   2,
				ChainName: "poly",
				From:      "0x0000000000001",
				To:        "0x0000000000001",
				Amount:    new(big.Int).SetInt64(100),
				TaskID:    new(big.Int).SetUint64(2),
			},
		},
	}
	data, err := json.Marshal(params)

	task := &types.PartReBalanceTask{
		BaseTask: &types.BaseTask{State: types.PartReBalanceCross, Message: ""},
		Base:     &types.Base{ID: 1},
		Params:   string(data),
	}
	part_rebalance.CreateReceiveFromBridgeTask(task, dbtest)
}
