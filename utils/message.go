package utils

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
)

type message struct {
	Status  string
	Content string
}

func GenFullRebalanceMessage(state int, content string) (string, string) {
	status := types.FullReBalanceStateName[state]

	return genMessage(status, content), status
}

func GenPartRebalanceMessage(state int, content string) (string, string) {
	status := types.PartReBalanceStateName[state]
	return genMessage(status, content), status
}

func GenTxMessage(state int, content string) string {
	status := types.TransactionStateName[types.TransactionState(state)]
	return genMessage(status, content)
}

func genMessage(status, content string) string {
	message := &message{status, content}
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Warnf("genMessage marshal err:%v, data:%+v", err, message)
	}
	return string(data)
}
