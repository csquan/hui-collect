package types

import (
	"context"
	"fmt"
	"github.com/starslabhq/hermes-rebalance/clients"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func SendToBridgeInput(bridgeAddress common.Address, amount *big.Int, taskID *big.Int) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("sendToBridge", bridgeAddress, amount, taskID)
}

func ReceiveFromBridgeInput(amount *big.Int, taskID *big.Int) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("receiveFromBridge", amount, taskID)
}

func ClaimFromVaultInput(address []common.Address, baseTokenAmount []*big.Int, counterTokenAmount []*big.Int) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("claimFromVaultInput", address, baseTokenAmount, counterTokenAmount)
}

func InvestInput(address []common.Address, baseTokenAmount []*big.Int, counterTokenAmount []*big.Int) (input []byte, err error) {
	r := strings.NewReader(content)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("invest", address, baseTokenAmount, counterTokenAmount)
}

func ApproveInput(address string) (input []byte, err error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("approve", common.HexToAddress(address), new(big.Int).SetInt64(math.MaxInt64))
}

func AllowanceInput(from string, to string) (input []byte, err error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}
	return abi.Pack("allowance", common.HexToHash(from), common.HexToHash(to))
}

func AllowanceOutput(result hexutil.Bytes) ([]interface{}, error) {
	r := strings.NewReader(erc20abi)
	abi, err := abi.JSON(r)
	if err != nil {
		return nil, err
	}

	return abi.Unpack("allowance", result)
}

func DecodeTransaction(txRaw string) (transaction *etypes.Transaction, err error) {
	transaction = &etypes.Transaction{}
	b, err := hexutil.Decode(txRaw)
	if err != nil {
		return
	}
	err = rlp.DecodeBytes(b, &transaction)
	return
}

func GetNonce(address string, chainName string) (uint64, error) {
	client, ok := clients.ClientMap[chainName]
	if !ok {
		return 0, fmt.Errorf("not find chain client, chainName:%v", chainName)
	}
	//TODO client.PendingNonceAt() ?
	return client.NonceAt(context.Background(), common.HexToAddress(address), nil)
}

func GetGasPrice(chainName string) (*big.Int, error) {
	client, ok := clients.ClientMap[chainName]
	if !ok {
		return nil, fmt.Errorf("not find chain client, chainName:%v", chainName)
	}
	return client.SuggestGasPrice(context.Background())
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
          "internalType": "address",
          "name": "bridge",
          "type": "address"
        },
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
      "name": "sendToBridge",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
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
