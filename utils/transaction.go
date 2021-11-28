package utils

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/starslabhq/hermes-rebalance/types"
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
	return abi.Pack("invest", param)
}

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
