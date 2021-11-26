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
		AssetTransferIn: []*types.AssetTransferInParam{
		&types.AssetTransferInParam{
			Amount:    new(big.Int).SetInt64(1000),
			TaskId:    new(big.Int).SetInt64(1),
			ChainId:   1,
			ChainName: "bsc",
			From:      "0000000000000000000000",
			To:        "1111111111111111111111",
		},
		&types.AssetTransferInParam{
			Amount:    new(big.Int).SetInt64(1000),
			TaskId:    new(big.Int).SetInt64(1),
			ChainId:   2,
			ChainName: "poly",
			From:      "0000000000000000000000",
			To:        "1111111111111111111111",
		},
	},
	}
	data, err := json.Marshal(params)

	task := &types.PartReBalanceTask{
		BaseTask: &types.BaseTask{State: types.PartReBalanceCross, Message: ""},
		Base:     &types.Base{ID: 1},
		Params:   string(data),
	}
	part_rebalance.CreateTransferInTask(task, dbtest)
}
