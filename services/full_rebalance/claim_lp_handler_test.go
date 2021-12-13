package full_rebalance

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/tokens/mock_tokens"
	"github.com/starslabhq/hermes-rebalance/types"
)

var h *claimLPHandler

func init() {
	r := strings.NewReader(vaultClaimAbi)
	abi, err := abi.JSON(r)
	if err != nil {
		logrus.Fatalf("claim abi err:%v", err)
	}
	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "test:123@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	if err != nil {
		logrus.Fatalf("create mysql cli err:%v", err)
	}
	h = &claimLPHandler{
		abi: abi,
		db:  dbtest,
	}
}
func TestNumber(t *testing.T) {
	num, _ := decimal.NewFromString("1.2")
	ten, _ := decimal.NewFromString("10")
	ten = ten.Pow(decimal.NewFromFloat(18))
	num = num.Mul(ten)

	t.Logf("num:%s", num.String())
}

func TestGetClaimParamsAndCreateTask(t *testing.T) {
	var (
		strategyAddr0 string = "addr0"
		strategyAddr1 string = "addr1"
		strategyAddr2 string = "addr2"
	)
	lps := []*types.LiquidityProvider{
		&types.LiquidityProvider{
			Chain: "BSC",
			LpInfoList: []*types.LpInfo{
				&types.LpInfo{
					StrategyAddress:  strategyAddr0,
					BaseTokenSymbol:  "BTCB",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1.5",
					QuoteTokenAmount: "2",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr0,
					BaseTokenSymbol:  "BTCB",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1",
					QuoteTokenAmount: "3",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr1,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "1.2",
					QuoteTokenAmount: "1",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr1,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDT",
					BaseTokenAmount:  "2",
					QuoteTokenAmount: "2",
				},
				&types.LpInfo{
					StrategyAddress:  strategyAddr2,
					BaseTokenSymbol:  "ETH",
					QuoteTokenSymbol: "USDC",
					BaseTokenAmount:  "2",
					QuoteTokenAmount: "2",
				},
			},
		},
	}
	valuts := []*types.VaultInfo{
		&types.VaultInfo{
			Currency: "btc",
			ActiveAmount: map[string]*types.ControllerInfo{
				"BSC": &types.ControllerInfo{
					ControllerAddress: "vault0",
				},
			},
		},
		&types.VaultInfo{
			Currency: "eth",
			ActiveAmount: map[string]*types.ControllerInfo{
				"BSC": &types.ControllerInfo{
					ControllerAddress: "vault1",
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := mock_tokens.NewMockTokener(ctrl)
	token.EXPECT().GetCurrency(gomock.Any(), "ETH").Return("eth").AnyTimes()
	token.EXPECT().GetCurrency(gomock.Any(), "BTCB").Return("btc")
	h.token = token
	params, err := h.getClaimParams(lps, valuts)
	if err != nil {
		t.Fatalf("get claim params err:%v", err)
	}
	b, _ := json.Marshal(params)
	t.Logf("claim params:%s", b)

	token.EXPECT().GetDecimals(gomock.Any(), gomock.Any()).Return(18, true).AnyTimes()
	tasks, err := h.createTxTask(1, params)
	if err != nil {
		t.Fatalf("create tx task err:%v", err)
	}
	b, _ = json.Marshal(tasks)
	t.Logf("tasks:%s", b)

	err = h.insertTxTasksAndUpdateState(tasks, &types.FullReBalanceTask{
		Base: &types.Base{
			ID: 1,
		},
		BaseTask: &types.BaseTask{
			State: 0,
		},
	}, types.FullReBalanceClaimLP)
	if err != nil {
		t.Fatalf("insert tx tasks and update state err:%v", err)
	}
}

func TestGetTasks(t *testing.T) {
	tasks, err := h.getTxTasks(1)
	if err != nil {
		t.Fatalf("get tasks err:%v", err)
	}
	b, _ := json.Marshal(tasks)
	t.Logf("tasks:%s", b)
}
