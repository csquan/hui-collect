package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type CrossBalanceItem struct {
	FromChain    string `json:"from_chain"`
	ToChain      string `json:"to_chain"`
	FromAddr     string `json:"from_addr"`
	ToAddr       string `json:"to_addr"`
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       string `json:"Amount"`
}

type ReceiveFromBridgeParam struct {
	ChainId   int    `json:"chain_id"`
	ChainName string `json:"chain_name"`
	From      string `json:"from"`
	To        string `json:"to"` //合约地址

	Erc20ContractAddr common.Address `json:"erc20_contract_addr"` //erc20 token地址，用于授权

	Amount string `json:"amount"` //链上精度值的amount，需要提前转换
	TaskID string `json:"task_id"`
}

type Params struct {
	CrossBalances           []*CrossBalanceItem       `json:"cross_balances"`
	ReceiveFromBridgeParams []*ReceiveFromBridgeParam `json:"receive_from_bridge_params"`
	InvestParams            []*InvestParam            `json:"invest_params"`
	SendToBridgeParams      []*SendToBridgeParam      `json:"send_to_bridge_params"`
}

type InvestParam struct {
	ChainId   int    `json:"chain_id"`
	ChainName string `json:"chain_name"`
	From      string `json:"from"`
	To        string `json:"to"` //合约地址

	StrategyAddresses  []common.Address `json:"strategy_addresses"`
	BaseTokenAmount    []string         `json:"base_token_amount"`
	CounterTokenAmount []string         `json:"counter_token_amount"`
}

type SendToBridgeParam struct {
	ChainId   int    `json:"chain_id"`
	ChainName string `json:"chain_name"`
	From      string `json:"from"`
	To        string `json:"to"` //合约地址

	BridgeAddress common.Address `json:"bridge_address"`
	Amount        string         `json:"amount"`
	TaskID        string         `json:"task_id"`
}

type ClaimFromVaultParam struct {
	ChainId   int    `json:"chain_id"`
	ChainName string `json:"chain_name"`
	From      string `json:"chain_from"`
	To        string `json:"to"` //合约地址

	StrategyAddresses  []common.Address `json:"strategy_addresses"`
	BaseTokenAmount    []string         `json:"base_token_amount"`
	CounterTokenAmount []string         `json:"counter_token_amount"`
}

type TransactionParamInterface interface {
	CreateTask(rebalanceTaskID uint64) (*TransactionTask, error)
}

func toBigIntAmounts(bases, counters []string) ([]*big.Int, []*big.Int) {
	var baseAmounts, counterAmounts []*big.Int
	for _, v := range bases {
		amount, ok := new(big.Int).SetString(v, 10)
		if !ok {
			logrus.Fatalf("base amount to big.Int err,v:%s", v)
		}
		baseAmounts = append(baseAmounts, amount)
	}
	for _, v := range counters {
		amount, ok := new(big.Int).SetString(v, 10)
		if !ok {
			logrus.Fatalf("counter amount to big.Int err,v:%s", v)
		}
		counterAmounts = append(counterAmounts, amount)
	}
	return baseAmounts, counterAmounts
}

func (p *InvestParam) CreateTask(rebalanceTaskID uint64) (*TransactionTask, error) {
	var baseAmounts, counterAmounts = toBigIntAmounts(p.BaseTokenAmount, p.CounterTokenAmount)
	inputData, err := InvestInput(p.StrategyAddresses, baseAmounts, counterAmounts)
	if err != nil {
		return nil, err
	}
	paramData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	task := &TransactionTask{
		BaseTask:        &BaseTask{State: int(TxUnInitState)},
		RebalanceId:     rebalanceTaskID,
		TransactionType: int(Invest),
		ChainId:         p.ChainId,
		ChainName:       p.ChainName,
		From:            p.From,
		To:              p.To,
		Params:          string(paramData),
		InputData:       hexutil.Encode(inputData),
	}
	return task, nil
}

func (p *ClaimFromVaultParam) CreateTask(globalTaskID uint64) (*TransactionTask, error) {
	var bases, counters = toBigIntAmounts(p.BaseTokenAmount, p.CounterTokenAmount)
	inputData, err := InvestInput(p.StrategyAddresses, bases, counters)
	if err != nil {
		return nil, err
	}
	paramData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	task := &TransactionTask{
		BaseTask:        &BaseTask{State: int(TxUnInitState)},
		RebalanceId:     globalTaskID,
		TransactionType: int(ClaimFromVault),
		ChainId:         p.ChainId,
		ChainName:       p.ChainName,
		From:            p.From,
		To:              p.To,
		Params:          string(paramData),
		InputData:       hexutil.Encode(inputData),
	}
	return task, nil
}

func (p *ReceiveFromBridgeParam) CreateTask(rebalanceTaskID uint64) (task *TransactionTask, err error) {
	var amount, taskID *big.Int
	var ok bool
	if amount, ok = new(big.Int).SetString(p.Amount, 10); !ok {
		err = fmt.Errorf("CreateTask param error")
		return
	}
	if taskID, ok = new(big.Int).SetString(p.TaskID, 10); !ok {
		err = fmt.Errorf("CreateTask param error")
		return
	}
	inputData, err := ReceiveFromBridgeInput(amount, taskID)
	if err != nil {
		return nil, err
	}
	paramData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	task = &TransactionTask{
		BaseTask:        &BaseTask{State: int(TxUnInitState)},
		RebalanceId:     rebalanceTaskID,
		TransactionType: int(ReceiveFromBridge),
		ChainId:         p.ChainId,
		ChainName:       p.ChainName,
		From:            p.From,
		To:              p.To,
		Params:          string(paramData),
		InputData:       hexutil.Encode(inputData),
	}
	return task, nil
}

func (p *SendToBridgeParam) CreateTask(rebalanceTaskID uint64) (task *TransactionTask, err error) {
	var amount, taskID *big.Int
	var ok bool
	if amount, ok = new(big.Int).SetString(p.Amount, 10); !ok {
		err = fmt.Errorf("CreateTask param error")
		return
	}
	if taskID, ok = new(big.Int).SetString(p.TaskID, 10); !ok {
		err = fmt.Errorf("CreateTask param error")
		return
	}
	inputData, err := SendToBridgeInput(p.BridgeAddress, amount, taskID)
	if err != nil {
		return nil, err
	}
	paramData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	task = &TransactionTask{
		BaseTask:        &BaseTask{State: int(TxUnInitState)},
		RebalanceId:     rebalanceTaskID,
		TransactionType: int(SendToBridge),
		ChainId:         p.ChainId,
		ChainName:       p.ChainName,
		From:            p.From,
		To:              p.To,
		Params:          string(paramData),
		InputData:       hexutil.Encode(inputData),
	}
	return task, nil
}
