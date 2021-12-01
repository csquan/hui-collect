package types

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	From      string `json:"chain_from"`
	To        string `json:"to"` //合约地址

	StrategyAddresses  []common.Address `json:"strategy_addresses"`
	BaseTokenAmount    []*big.Int       `json:"base_token_amount"`
	CounterTokenAmount []*big.Int       `json:"counter_token_amount"`
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

type TransactionParamInterface interface {
	CreateTask(rebalanceTaskID uint64) (*TransactionTask, error)
}

func (p *InvestParam) CreateTask(rebalanceTaskID uint64) (*TransactionTask, error) {
	inputData, err := InvestInput(p.StrategyAddresses, p.BaseTokenAmount, p.CounterTokenAmount)
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
