package full_rebalance

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-xorm/xorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/tokens"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

const (
	vaultClaimAbi = `[{
		"inputs": [
		  {
			"internalType": "address[]",
			"name": "_strategies",
			"type": "address[]"
		  },
		  {
			"internalType": "uint256[]",
			"name": "_baseTokensAmount",
			"type": "uint256[]"
		  },
		  {
			"internalType": "uint256[]",
			"name": "_counterTokensAmount",
			"type": "uint256[]"
		  },
		  {
			"internalType": "uint256[]",
			"name": "_lpClaimIds",
			"type": "uint256[]"
		  }
		],
		"name": "claimAll",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	  }]`
)

type lpDataGetter interface {
	getLpData(url string) (lpList *types.Data, err error)
}

type getter func(url string) (lpList *types.Data, err error)

func (g getter) getLpData(url string) (lpList *types.Data, err error) {
	return g(url)
}

type claimLPHandler struct {
	token     tokens.Tokener
	db        types.IDB
	abi       abi.ABI
	claimFrom string //TODO
	conf      *config.Config
	getter    lpDataGetter
}

func newClaimLpHandler(conf *config.Config, db types.IDB) *claimLPHandler {
	r := strings.NewReader(vaultClaimAbi)
	abi, err := abi.JSON(r)
	if err != nil {
		logrus.Fatalf("claim abi err:%v", err)
	}
	return &claimLPHandler{
		db:     db,
		abi:    abi,
		getter: getter(getLpData),
	}
}

func (w *claimLPHandler) Name() string {
	return "full_rebalance_claim"
}

type claimParam struct {
	ChainId    int
	ChainName  string
	VaultAddr  string
	Strategies []*strategy
}

type strategy struct {
	StrategyAddr string
	BaseSymbol   string
	QuoteSymbol  string
	BaseAmount   decimal.Decimal
	QuoteAmount  decimal.Decimal
}

func strMustToDecimal(v string) decimal.Decimal {
	num, err := decimal.NewFromString(v)
	if err != nil {
		logrus.Fatalf("str to decimal err v:%s", v)
	}
	return num
}

func findParams(params []*claimParam, vaultAddr string) *claimParam {
	for _, param := range params {
		if param.VaultAddr == vaultAddr {
			return param
		}
	}
	return nil
}

func (w *claimLPHandler) getClaimParams(lps []*types.LiquidityProvider, valuts []*types.VaultInfo) (params []*claimParam) {
	params = make([]*claimParam, 0)
	for _, lp := range lps {
		strategiesM := make(map[string]*strategy)
		for _, info := range lp.LpInfoList {
			base := strMustToDecimal(info.BaseTokenAmount)
			quote := strMustToDecimal(info.QuoteTokenAmount)

			if s, ok := strategiesM[info.StrategyAddress]; ok {
				s.BaseAmount = s.BaseAmount.Add(base)
				s.QuoteAmount = s.QuoteAmount.Add(quote)
			} else {
				s = &strategy{
					StrategyAddr: info.StrategyAddress,
					BaseSymbol:   info.BaseTokenSymbol,
					BaseAmount:   base,
					QuoteSymbol:  info.QuoteTokenSymbol,
					QuoteAmount:  quote,
				}
				strategiesM[s.StrategyAddr] = s

				addr, ok := w.getVaultAddr(s.BaseSymbol, lp.Chain, valuts)
				if !ok {
					return
				}

				param := findParams(params, addr)
				if param == nil {
					param := &claimParam{
						ChainId:    lp.ChainId,
						ChainName:  lp.Chain,
						VaultAddr:  addr,
						Strategies: []*strategy{s},
					}
					params = append(params, param)
				} else {
					param.Strategies = append(param.Strategies, s)
				}
			}
		}
	}

	return params
}

func powN(num decimal.Decimal, n int) decimal.Decimal {
	//10^n
	ten := decimal.NewFromFloat(10).Pow(decimal.NewFromFloat(float64(n)))
	return num.Mul(ten)
}

func decimalToBigInt(num decimal.Decimal) *big.Int {
	ret, ok := new(big.Int).SetString(num.String(), 10)
	if !ok {
		logrus.Fatalf("decimal to big.Int err num:%s", num.String())
	}
	return ret
}

func (w *claimLPHandler) createTxTask(tid uint64, params []*claimParam) ([]*types.TransactionTask, error) {
	var tasks []*types.TransactionTask
	for _, param := range params {
		var (
			addrs    []common.Address
			bases    []*big.Int
			quotes   []*big.Int
			claimIds []*big.Int
		)

		for _, s := range param.Strategies {
			addr := common.HexToAddress(s.StrategyAddr)
			addrs = append(addrs, addr)

			//base
			decimal0 := w.token.GetDecimals(s.BaseSymbol)
			if decimal0 == 0 {
				logrus.Fatalf("unexpectd decimal bseSymbol:%s", s.BaseSymbol)
			}
			baseDecimal := powN(s.BaseAmount, decimal0)
			base := decimalToBigInt(baseDecimal)
			bases = append(bases, base)

			//quote
			decimal1 := w.token.GetDecimals(s.QuoteSymbol)
			if decimal1 == 0 {
				logrus.Fatalf("unexpectd decimal quoteSymbol:%s", s.QuoteSymbol)
			}
			quoteDecimal := powN(s.QuoteAmount, decimal1)
			quote := decimalToBigInt(quoteDecimal)
			quotes = append(quotes, quote)
			claimIds = append(claimIds, big.NewInt(0))
		}
		logrus.Infof("claimAll tid:%d,addrs:%v,bases:%v,quotes:%v,claimIds:%v", tid, addrs, bases, quotes, claimIds)
		input, err := w.abi.Pack("claimAll", addrs, bases, quotes, claimIds)
		if err != nil {
			return nil, fmt.Errorf("claim pack err:%v", err)
		}
		encoded, _ := json.Marshal(param)
		task := &types.TransactionTask{
			FullRebalanceId: tid,
			BaseTask:        &types.BaseTask{State: int(types.TxUnInitState)},
			TransactionType: int(types.ClaimFromVault),
			ChainId:         param.ChainId,
			ChainName:       param.ChainName,
			From:            w.claimFrom,
			To:              param.VaultAddr,
			Params:          string(encoded),
			InputData:       hexutil.Encode(input),
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (w *claimLPHandler) insertTxTasksAndUpdateState(txTasks []*types.TransactionTask,
	fullTask *types.FullReBalanceTask, state types.FullReBalanceState) error {
	utils.CommitWithSession(w.db, func(s *xorm.Session) error {
		err := w.db.SaveTxTasks(s, txTasks)
		if err != nil {
			return fmt.Errorf("claim save txtasks err:%v,tid:%d", err, fullTask.ID)
		}
		fullTask.State = state
		err = w.db.UpdateFullReBalanceTask(s, fullTask)
		if err != nil {
			return fmt.Errorf("update claim task err:%v,tid:%d", err, fullTask.ID)
		}
		return nil
	})
	return nil
}

func (w *claimLPHandler) getVaultAddr(tokenSymbol, chain string, vaults []*types.VaultInfo) (string, bool) {
	currency := w.token.GetCurrency(tokenSymbol, chain)
	for _, valut := range vaults {
		if valut.Currency == currency {
			c, ok := valut.ActiveAmount[chain]
			if !ok {
				b, _ := json.Marshal(valut)
				logrus.Fatalf("valut activeAmount not found chain:%s,valut:%s", chain, b)
			}
			return c.ControllerAddress, true
		}
	}
	return "", false
}

func (w *claimLPHandler) Do(task *types.FullReBalanceTask) error {

	data, err := w.getter.getLpData(w.conf.ApiConf.LpUrl)
	if err != nil {
		return fmt.Errorf("claim get lpData err:%v", err)
	}

	var lps = data.LiquidityProviderList
	params := w.getClaimParams(lps, data.VaultInfoList)

	txTasks, err := w.createTxTask(task.ID, params)
	if err != nil {
		return err
	}
	return w.insertTxTasksAndUpdateState(txTasks, task, types.FullReBalanceClaimLP)
}

func (w *claimLPHandler) getTxTasks(fullRebalanceId uint64) ([]*types.TransactionTask, error) {
	tasks, err := w.db.GetTransactionTasksWithFullRebalanceId(fullRebalanceId, types.ClaimFromVault)
	return tasks, err
}

func (w *claimLPHandler) CheckFinished(task *types.FullReBalanceTask) (finished bool, nextState types.FullReBalanceState, err error) {
	txTasks, err := w.getTxTasks(task.ID)
	if err != nil {
		return false, types.FullReBalanceClaimLP, fmt.Errorf("full_r get txTasks err:%v", err)
	}
	taskCnt := len(txTasks)
	if taskCnt == 0 {
		logrus.Fatalf("unexpected claim txTasks size tid:%d", task.ID)
	}
	var (
		sucCnt  int
		failCnt int
	)
	for _, task := range txTasks {
		if task.State == int(types.TxSuccessState) {
			sucCnt++
		}
		if task.State == int(types.TxFailedState) {
			failCnt++
		}
	}
	if sucCnt == taskCnt {
		return true, types.FullReBalanceMarginBalanceTransferOut, nil
	}
	if failCnt != 0 {
		logrus.Warnf("claim lp handler failed tid:%d", task.ID)
		return false, types.FullReBalanceFailed, nil
	}
	return false, types.FullReBalanceClaimLP, nil
}
