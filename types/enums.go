package types

type TransactionState int

const (
	TxInitState TransactionState = iota
	TxAssmblyState
	TxSignState
	TxBroadcastState
	TxCheckState
	TxSuccessState
	TxFailedState
)

var TransactionStateName = map[TransactionState]string{
	TxInitState:      "Init",
	TxAssmblyState:   "Assmbly",
	TxSignState:      "Sign",
	TxBroadcastState: "broadcast",
	TxCheckState:     "check",
	TxSuccessState:   "Success",
	TxFailedState:    "Failed",
}
