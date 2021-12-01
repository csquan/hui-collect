package utils

import (
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
)

var ClientMap map[string]*ethclient.Client

func Init(c *config.Config) {
	ClientMap = make(map[string]*ethclient.Client)
	for k, chain := range c.Chains {
		client, err := rpc.DialHTTPWithClient(chain.RpcUrl, &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
			Timeout: time.Duration(chain.Timeout) * time.Millisecond,
		})
		if err != nil {
			logrus.Fatalf("init rpc client err:%v", err)
		}

		ClientMap[k] = ethclient.NewClient(client)
	}
}
