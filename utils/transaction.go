package utils

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/starslabhq/hermes-rebalance/types"
	"math"
	"math/big"
	"strings"
)

func ReceiveFromBridgeInput(param *types.ReceiveFromBridgeParam) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("receiveFromBridge", param.Amount, param.TaskID)
}

func InvestInput(param *types.InvestParam) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("invest", param.Address, param.Token1Amounts, param.Token2Amounts)
}

func ApproveInput(param *types.ReceiveFromBridgeParam) (input []byte, err error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("approve", common.HexToAddress(param.To), new(big.Int).SetInt64(math.MaxInt64))
}

func AllowanceInput(param *types.ReceiveFromBridgeParam) (input []byte, err error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("allowance", common.HexToHash(param.From), common.HexToHash(param.To))
}

func AllowanceOutput(result hexutil.Bytes) ([]interface{}, error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}

	return abi.Unpack("allowance", result)
}

const erc20abi = `[
	{
        "constant":false,
        "inputs":[
            {
                "name":"_spender",
                "type":"address"
            },
            {
                "name":"_value",
                "type":"uint256"
            }
        ],
        "name":"approve",
        "outputs":[
            {
                "name":"",
                "type":"bool"
            }
        ],
        "payable":false,
        "stateMutability":"nonpayable",
        "type":"function"
    },
    {
        "constant": true,
        "inputs": [
            {
                "name": "_owner",
                "type": "address"
            },
            {
                "name": "_spender",
                "type": "address"
            }
        ],
        "name": "allowance",
        "outputs": [
            {
                "name": "",
                "type": "uint256"
            }
        ],
        "payable": false,
        "stateMutability": "view",
        "type": "function"
    }
]`

var content = `[
    {
      "inputs": [
        {
          "internalType": "uint256",
          "name": "amount",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "taskId",
          "type": "uint256"
        }
      ],
      "name": "receiveFromBridge",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
        "inputs":[
            {
                "internalType":"address[]",
                "name":"_strategies",
                "type":"address[]"
            },
            {
                "internalType":"uint256[]",
                "name":"_baseTokensAmount",
                "type":"uint256[]"
            },
            {
                "internalType":"uint256[]",
                "name":"_counterTokensAmount",
                "type":"uint256[]"
            }
        ],
        "name":"invest",
        "outputs":[

        ],
        "stateMutability":"nonpayable",
        "type":"function"
    }
]`
