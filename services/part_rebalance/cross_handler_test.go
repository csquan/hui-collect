package part_rebalance

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
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
	createInvestTask(conf)

}
func createReceiveFromBridgeTask(conf *config.Config) {
	var tasks []*types.TransactionTask
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{
			ChainId:           1,
			ChainName:         "heco",
			From:              "606288c605942f3c84a7794c0b3257b56487263c",
			To:                "0xC7c38F93036BC13168B4f657296753568f49ef09",
			Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
			Amount:            new(big.Int).SetInt64(1),
			TaskID:            new(big.Int).SetUint64(1),
		},
	}
	params := &types.Params{
		ReceiveFromBridgeParams: ReceiveFromBridgeParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base:   &types.Base{ID: 10},
		Params: string(data),
	}
	utils.Init(conf)
	dbConnection, err := db.NewMysql(&conf.DataBase)
	c := &crossHandler{db: dbConnection, clientMap: utils.ClientMap}
	tasks, err = c.CreateReceiveFromBridgeTask(task)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}
	if tasks, err = SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	err = c.db.SaveTxTasks(dbConnection.GetSession(), tasks)
}

func createInvestTask(conf *config.Config) {
	var tasks []*types.TransactionTask
	address := common.HexToAddress("0xa929022c9107643515f5c777ce9a910f0d1e490c")
	InvestParams := []*types.InvestParam{
		&types.InvestParam{
			ChainId:            1,
			ChainName:          "heco",
			From:               "606288c605942f3c84a7794c0b3257b56487263c",
			To:                 "0xC7c38F93036BC13168B4f657296753568f49ef09",
			Address:            []common.Address{address},
			BaseTokenAmount:    []*big.Int{new(big.Int)},
			CounterTokenAmount: []*big.Int{new(big.Int)},
			//Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
			//Amount:    new(big.Int).SetInt64(1),
			//TaskID:    new(big.Int).SetUint64(1),
		},
	}
	params := &types.Params{
		InvestParams: InvestParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base:   &types.Base{ID: 10},
		Params: string(data),
	}
	utils.Init(conf)
	dbConnection, err := db.NewMysql(&conf.DataBase)
	c := &transferInHandler{db: dbConnection}
	tasks, err = c.CreateInvestTask(task)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}
	if tasks, err = SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	err = c.db.SaveTxTasks(dbConnection.GetSession(), tasks)
}
