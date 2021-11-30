package part_rebalance

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
	"math/big"
	"net/http"
	"testing"
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
	conf, err := config.LoadConf("../../"+confFile)
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

	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "root:123456sj@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4&parseTime=true",
	})
	var tasks []*types.TransactionTask
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{
			ChainId:   1,
			ChainName: "heco",
			From:      "606288c605942f3c84a7794c0b3257b56487263c",
			To:        "0x882d0c2435CBB8A0E774b674a5a7e64ea6789fe0",
			Erc20ContractAddr: common.HexToAddress("0x6D2dbA4F00e0Bbc2F93eb43B79ddd00f65fB6bEc"),
			Amount:    new(big.Int).SetInt64(1),
			TaskID:    new(big.Int).SetUint64(1),
		},
	}
	params := &types.Params{
		ReceiveFromBridgeParams:ReceiveFromBridgeParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base: &types.Base{ID: 10},
		Params: string(data),
	}
	utils.Init(conf)
	c := &crossHandler{db:dbtest, clientMap: utils.ClientMap}
	tasks, err = c.CreateReceiveFromBridgeTask(task)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}
	if tasks, err = c.SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	err = c.db.SaveTxTasks(dbtest.GetSession(), tasks)
}