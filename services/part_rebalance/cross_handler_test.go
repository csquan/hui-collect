package part_rebalance

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/clients"
	"math/big"
	"net/http"
	"testing"

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

func init() {
	flag.StringVar(&confFile, "conf", "config.yaml", "conf file")
}
func TestCreateTreansfer(t *testing.T) {
	flag.Parse()
	logrus.Info(confFile)
	conf, err := config.LoadConf("../../" + confFile)
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
	var tasks []*types.TransactionTask
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{
			ChainId:           1,
			ChainName:         "heco",
			From:              "606288c605942f3c84a7794c0b3257b56487263c",
			To:                "0xC7c38F93036BC13168B4f657296753568f49ef09",
			Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
			Amount:            "1",
			TaskID:            "1",
		},
	}
	SendToBridgeParam := []*types.SendToBridgeParam{
		&types.SendToBridgeParam{
			ChainId:           1,
			ChainName:         "heco",
			From:              "606288c605942f3c84a7794c0b3257b56487263c",
			To:                "0xC7c38F93036BC13168B4f657296753568f49ef09",

			BridgeAddress: common.HexToAddress("606288c605942f3c84a7794c0b3257b56487263c"),
			Amount:            "1",
			TaskID:            "1",
		},
	}
	InvestParams := []*types.InvestParam{
		&types.InvestParam{
			ChainId:            1,
			ChainName:          "heco",
			From:               "606288c605942f3c84a7794c0b3257b56487263c",
			To:                 "0xC7c38F93036BC13168B4f657296753568f49ef09",
			StrategyAddresses:  []common.Address{0xa929022c9107643515f5c777ce9a910f0d1e490c},
			BaseTokenAmount:    []*big.Int{new(big.Int)},
			CounterTokenAmount: []*big.Int{new(big.Int)},
			//Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
		},
	}
	params := &types.Params{
		ReceiveFromBridgeParams: ReceiveFromBridgeParams,
		SendToBridgeParams: SendToBridgeParam,
		InvestParams: InvestParams,
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
	tasks, err = CreateTransactionTask(task, types.SendToBridge)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}
	err = dbConnection.SaveTxTasks(dbConnection.GetSession(), tasks)
}

