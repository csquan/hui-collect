package services

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

var c *ethclient.Client

func init() {
	client, err := rpc.DialHTTPWithClient("https://polygon-rpc.com", &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Fatalf("init rpc client err:%v", err)
	}
	c = ethclient.NewClient(client)

}
func TestBlockSafe(t *testing.T) {
	tx := &Transaction{
		config: &config.Config{
			Chains: map[string]*config.ChainInfo{
				"bsc": &config.ChainInfo{
					BlockSafe: 5,
				},
			},
		},
	}
	type input struct {
		chain string
		txh   uint64
		curh  uint64
	}
	tests := []struct {
		input input
		want  bool
	}{
		{
			input: input{
				"bsc",
				1,
				6,
			},
			want: false,
		},
		{
			input: input{
				"bsc",
				1,
				7,
			},
			want: true,
		},
	}
	for i, test := range tests {
		ret := tx.isTxBlockSafe(test.input.chain, test.input.txh, test.input.curh)
		if ret != test.want {
			t.Errorf("unexpected i:%d", i)
		}
	}
}

func TestTransactionFaiMsg(t *testing.T) {
	task := &types.TransactionTask{
		ChainName:       "bsc",
		Hash:            "hash",
		TransactionType: 2,
	}
	err := fmt.Errorf(txErrFormat, task.ChainName, task.Hash, task.TransactionType)
	t.Logf("errmsg:%s", err.Error())
}
