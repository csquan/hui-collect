package types

type TransactionState int

const (
	TxCollectingState TransactionState = iota
	TxInitState
	TxAssmblyState
	TxSignState
	TxBroadcastState
	TxCheckState
	TxCollectedState
)

var TransactionStateName = map[TransactionState]string{
	TxCollectingState: "collecting",
	TxInitState:       "Init",
	TxAssmblyState:    "Assmbly",
	TxSignState:       "Sign",
	TxBroadcastState:  "broadcast",
	TxCheckState:      "check",
	TxCollectedState:  "collected",
}
