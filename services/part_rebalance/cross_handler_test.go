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
	err = log.Init(conf.AppName, conf.LogConf, "dev")
	if err != nil {
		log.Fatal(err)
	}
	err = createTask(conf)
	if err != nil {
		t.Fatalf("create task err:%v", err)
	}
}
func createTask(conf *config.Config) error {

	var (
		ethAmount        string = powN(big.NewInt(1), 17)
		usdtAmount       string = powN(big.NewInt(320), 18)
		ethAmountBridge  string = "0.1"
		usdtAmountBridge string = "320"
	)
	//var tasks []*types.TransactionTask
	SendToBridgeParam := []*types.SendToBridgeParam{
		&types.SendToBridgeParam{ //usdt
			ChainId:   128,
			ChainName: "heco",
			From:      "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			To:        "0xa82Aa96714bd30EaE09BeB3291834A845B3E5B72",

			BridgeAddress: common.HexToAddress("0x9f0583a209fedbc404c4968e2157c2e7d4359803"),
			Amount:        usdtAmount,
			TaskID:        "1",
		},
		&types.SendToBridgeParam{
			ChainId:   128, //eth
			ChainName: "heco",
			From:      "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			To:        "0x497DeF83FFA5d6C42B39Acd49C292EC49EaD496E",

			BridgeAddress: common.HexToAddress("0x9f0583a209fedbc404c4968e2157c2e7d4359803"),
			Amount:        ethAmount,
			TaskID:        "2",
		},
	}
	CrossBalances := []*types.CrossBalanceItem{
		&types.CrossBalanceItem{ //USDT
			FromChain:    "HECO",
			ToChain:      "BSC",
			FromAddr:     "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			ToAddr:       "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			FromCurrency: "USDT",
			ToCurrency:   "USDT",
			Amount:       usdtAmountBridge,
		},
		&types.CrossBalanceItem{ //ETH
			FromChain:    "HECO",
			ToChain:      "BSC",
			FromAddr:     "0x9f0583a209fedbc404c4968e2157c2e7d4359803",
			ToAddr:       "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			FromCurrency: "ETH",
			ToCurrency:   "ETH",
			Amount:       ethAmountBridge,
		},
	}
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{ //USDT
			ChainId:           56,
			ChainName:         "bsc",
			From:              "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			To:                "0x7867226d16440FFbEAb39225E7a137CA9ba98501",
			Erc20ContractAddr: common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"),
			Amount:            usdtAmount,
			TaskID:            "1",
		},
		&types.ReceiveFromBridgeParam{ //ETH
			ChainId:           56,
			ChainName:         "bsc",
			From:              "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			To:                "0x36Bdee19a991dB559F3072a7974a85759BeE1224",
			Erc20ContractAddr: common.HexToAddress("0x2170ed0880ac9a755fd29b2688956bd959f933f8"),
			Amount:            ethAmount,
			TaskID:            "2",
		},
	}

	InvestParams := []*types.InvestParam{
		// &types.InvestParam{ //USDT
		// 	ChainId:   56,
		// 	ChainName: "bsc",
		// 	From:      "0x74938228ae77e5fcc3504ad46fac4a965d210761",
		// 	To:        "0x36Bdee19a991dB559F3072a7974a85759BeE1224",
		// 	StrategyAddresses: []common.Address{
		// 		common.HexToAddress("0x32944aA3716E3B25e03c80De0A0b4c301CaccDC7"),
		// 	},
		// 	BaseTokenAmount: []string{
		// 		powN(big.NewInt(100), 18),
		// 	},
		// 	CounterTokenAmount: []string{
		// 		"0",
		// 	},
		// },
		&types.InvestParam{ //ETH
			ChainId:   56,
			ChainName: "bsc",
			From:      "0x74938228ae77e5fcc3504ad46fac4a965d210761",
			To:        "0x36Bdee19a991dB559F3072a7974a85759BeE1224", //TODO
			StrategyAddresses: []common.Address{
				common.HexToAddress("0x4956C5835eDBD358A1D51D6DD8B4E0C0665Fb640"), //solo
				common.HexToAddress("0xa71c55cB4A091c7Fb676C29913F74B51FFb6e981"), //bisSwap
			},
			BaseTokenAmount: []string{
				powN(big.NewInt(1), 13), //0.00001
				powN(big.NewInt(5), 13), //0.00005ETH
			},
			CounterTokenAmount: []string{
				"0",
				powN(big.NewInt(22), 16), //0.22U
			},
		},
	}
	params := &types.Params{
		CrossBalances:           CrossBalances,
		ReceiveFromBridgeParams: ReceiveFromBridgeParams,
		SendToBridgeParams:      SendToBridgeParam,
		InvestParams:            InvestParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base:   &types.Base{},
		Params: string(data),
	}
	clients.Init(conf)
	dbConnection, err := db.NewMysql(&conf.DataBase)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return err
	}

	err = dbConnection.SaveRebalanceTask(dbConnection.GetSession(), task)
	if err != nil {
		return fmt.Errorf("save rebalance err:%v", err)
	}
	return nil
}

func TestPowN(t *testing.T) {
	ret := powN(big.NewInt(2), 17)
	t.Logf("%s,%d", ret, len(ret))
}

func powN(num *big.Int, n int64) string {
	var a = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(n), big.NewInt(0))
	return num.Mul(num, a).String()
}

func TestGetOpened(t *testing.T) {
	opened := []*types.CrossTask{
		&types.CrossTask{
			Base: &types.Base{
				ID: 22,
			},
			RebalanceId:   1,
			ChainFrom:     "from",
			ChainTo:       "to",
			ChainFromAddr: "addr_from",
			ChainToAddr:   "addr_to",
			CurrencyFrom:  "c_from",
			CurrencyTo:    "c_to",
			Amount:        "12",
			State:         1,
		},
	}
	msg := getOpenedTaskMsg(opened)
	t.Logf("msg:%s", msg)
}
