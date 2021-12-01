package part_rebalance

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/starslabhq/hermes-rebalance/clients"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/types"
)

var (
	confFile string
)

func TestCreateTreansfer(t *testing.T) {
	conf, err := config.LoadConf("../../config.yaml")
	if err != nil {
		logrus.Errorf("load config error:%v", err)
		return
	}

	if conf.ProfPort != 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", conf.ProfPort), nil)
			if err != nil {
				panic(fmt.Sprintf("start pprof server err:%v", err))
			}
		}()
	}

	//setup log print
	err = log.Init(conf.AppName, conf.LogConf)
	if err != nil {
		log.Fatal(err)
	}
	createTask(conf)

}
func createTask(conf *config.Config) {
	//var tasks []*types.TransactionTask
	SendToBridgeParam := []*types.SendToBridgeParam{
		&types.SendToBridgeParam{
			ChainId:   128,
			ChainName: "heco",
			From:      "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			To:        "0xD95Cbc6907134b2C9F3Ba6c424D7d69d493D3014",

			BridgeAddress: common.HexToAddress("0x9f0583a209fedbc404c4968e2157c2e7d4359803"),
			Amount:        big.NewInt(0).Exp(big.NewInt(10), big.NewInt(20), big.NewInt(0)).String(),
			TaskID:        "1",
		},
	}
	CrossBalances := []*types.CrossBalanceItem{
		&types.CrossBalanceItem{
			FromChain:    "HECO",
			ToChain:      "BSC",
			FromAddr:     "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			ToAddr:       "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			FromCurrency: "USDT",
			ToCurrency:   "USDT",
			Amount:       "100",
		},
	}
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{
			ChainId:           56,
			ChainName:         "bsc",
			From:              "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			To:                "0xbFc4c5c1Bb5e9B806899eaAef7f04E278b59198A",
			Erc20ContractAddr: common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"),
			Amount:            big.NewInt(0).Exp(big.NewInt(10), big.NewInt(20), big.NewInt(0)).String(),
			TaskID:            "1",
		},
	}

	InvestParams := []*types.InvestParam{
		//&types.InvestParam{
		//	ChainId:            1,
		//	ChainName:          "heco",
		//	From:               "606288c605942f3c84a7794c0b3257b56487263c",
		//	To:                 "0xC7c38F93036BC13168B4f657296753568f49ef09",
		//	StrategyAddresses:  []common.Address{common.HexToAddress("0xa929022c9107643515f5c777ce9a910f0d1e490c")},
		//	BaseTokenAmount:    []*big.Int{new(big.Int)},
		//	CounterTokenAmount: []*big.Int{new(big.Int)},
		//	//Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
		//},
	}
	params := &types.Params{
		CrossBalances:           CrossBalances,
		ReceiveFromBridgeParams: ReceiveFromBridgeParams,
		SendToBridgeParams:      SendToBridgeParam,
		InvestParams:            InvestParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base:   &types.Base{ID: 10},
		Params: string(data),
	}
	clients.Init(conf)
	dbConnection, err := db.NewMysql(&conf.DataBase)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}

	err = dbConnection.SaveRebalanceTask(dbConnection.GetSession(), task)
}
