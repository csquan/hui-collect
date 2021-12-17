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

func GenFullRebalanceMessage(state int, content string) string {
	var status string
	switch state {
	case types.FullReBalanceInit:
		status = "Init"
	case types.FullReBalanceMarginIn:
		status = "MarginIn"
	case types.FullReBalanceClaimLP:
		status = "ClaimLP"
	case types.FullReBalanceMarginBalanceTransferOut:
		status = "MarginOut"
	case types.FullReBalanceRecycling:
		status = "Recycling"
	case types.FullReBalanceParamsCalc:
		status = "ParamsCalc"
	case types.FullReBalanceOndoing:
		status = "PartOndoing"
	case types.FullReBalanceSuccess:
		status = "Success"
	case types.FullReBalanceFailed:
		status = "Failed"
	default:
		status = ""
	}
	return genMessage(status, content)
}

func GenPartRebalanceMessage(state int, content string) string {
	var status string
	switch state {
	case types.PartReBalanceInit:
		status = "Init "
	case types.PartReBalanceTransferOut:
		status = "SendToBridge"
	case types.PartReBalanceCross:
		status = "Cross"
	case types.PartReBalanceTransferIn:
		status = "ReceiveFromBridge"
	case types.PartReBalanceInvest:
		status = "Invest"
	case types.PartReBalanceSuccess:
		status = "Success"
	case types.PartReBalanceFailed:
		status = "Failed"
	default:
		status = ""
	}
	return genMessage(status, content)
}

func GenTxMessage(state int, content string) string {
	var status string
	switch types.TransactionState(state) {
	case types.TxUnInitState:
		status = "Init"
	case types.TxAuditState:
		status = "Audit"
	case types.TxValidatorState:
		status = "Validator"
	case types.TxCheckReceiptState:
		status = "CheckReceipt"
	case types.TxSuccessState:
		status = "Success"
	case types.TxFailedState:
		status = "Failed"
	default:
		status = ""
	}
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
