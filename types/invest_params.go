package types

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/starslabhq/hermes-rebalance/utils"
)

type TransactionParamInterface interface {
	EncodeInput() (string, error)  //签名使用的input数据
	EncodeParam() (string, error)  //原生的参数，目的是增加db中数据的可读性。
	GetBase() (int, string, string, string)
}

func (p *InvestParam) GetBase() (int, string, string, string){
	return p.ChainId, p.ChainName, p.From, p.To
}

func (p *InvestParam) EncodeInput() (string, error){
	inputData, err := utils.InvestInput(p.Address, p.BaseTokenAmount, p.CounterTokenAmount)
	if err != nil{
		return "", err
	}
	return hexutil.Encode(inputData), nil
}

func (p *InvestParam) EncodeParam() (string, error){
	paramData, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(paramData), nil
}