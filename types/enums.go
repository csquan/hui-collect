package types

type TransactionState int

const (
	TxReadyCollectState TransactionState = iota
	TxCollectingState
	TxCollectedState
)

const (
	TxInitState TransactionState = iota
	TxAssmblyState
	TxSignState
	TxBroadcastState
	TxCheckState
)

var TransactionStateName = map[TransactionState]string{
	TxInitState:      "Init",
	TxAssmblyState:   "Assmbly",
	TxSignState:      "Sign",
	TxBroadcastState: "broadcast",
	TxCheckState:     "check",
}
