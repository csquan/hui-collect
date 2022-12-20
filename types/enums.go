package types

type TransactionState int

const (
	TxInitState TransactionState = iota
	TxAssmblyState
	TxSignState
	TxBroadcastState
	TxSuccessState
	TxFailedState
)

var TransactionStateName = map[TransactionState]string{
	TxInitState:      "Init",
	TxAssmblyState:   "Assmbly",
	TxSignState:      "Sign",
	TxBroadcastState: "Broadcast",
	TxSuccessState:   "Success",
	TxFailedState:    "Failed",
}
