package types

type FullReBalanceState = int

const (
	FullReBalanceInit                     FullReBalanceState = iota
	FullReBalanceMarginIn                                    //平无常 http请求
	FullReBalanceClaimLP                                     //拆LP 合约调用
	FullReBalanceMarginBalanceTransferOut                    //保证金转出至对冲账户
	FullReBalanceRecycling                                   //资金跨回
	FullReBalanceParamsCalc                                  // python 计算并创建partRebalanceTask
	FullReBalanceOndoing                                     // 检查partRebalanceTask状态
	FullReBalanceSuccess
	FullReBalanceFailed
)

var FullReBalanceStateName = map[FullReBalanceState]string{
	FullReBalanceInit:                     "Init",
	FullReBalanceMarginIn:                 "MarginIn",
	FullReBalanceClaimLP:                  "ClaimLP",
	FullReBalanceMarginBalanceTransferOut: "MarginOut",
	FullReBalanceRecycling:                "Recycling",
	FullReBalanceParamsCalc:               "ParamsCalc",
	FullReBalanceOndoing:                  "PartOndoing",
	FullReBalanceSuccess:                  "Success",
	FullReBalanceFailed:                   "Failed",
}

type PartReBalanceState = int

const (
	PartReBalanceInit PartReBalanceState = iota
	PartReBalanceTransferOut
	PartReBalanceCross
	PartReBalanceTransferIn
	PartReBalanceInvest
	PartReBalanceSuccess
	PartReBalanceFailed
)

var PartReBalanceStateName = map[PartReBalanceState]string{
	PartReBalanceInit:        "Init",
	PartReBalanceTransferOut: "SendToBridge",
	PartReBalanceCross:       "Cross",
	PartReBalanceTransferIn:  "ReceiveFromBridge",
	PartReBalanceInvest:      "Invest",
	PartReBalanceSuccess:     "Success",
	PartReBalanceFailed:      "Failed",
}

type TransactionState int

const (
	TxUnInitState TransactionState = iota
	TxAuditState
	TxValidatorState
	TxCheckReceiptState
	TxSuccessState
	TxFailedState
)

var TransactionStateName = map[TransactionState]string{
	TxUnInitState:       "Init",
	TxAuditState:        "Audit",
	TxValidatorState:    "Validator",
	TxCheckReceiptState: "CheckReceipt",
	TxSuccessState:      "Success",
	TxFailedState:       "Failed",
}

type CrossState = int

const (
	ToCreateSubTask CrossState = iota
	SubTaskCreated
	TaskSuc //all sub task suc
)

type CrossSubState int

const (
	ToCross CrossSubState = iota
	Crossing
	Crossed
)

type TransactionType int

const (
	SendToBridge TransactionType = iota
	ReceiveFromBridge
	Invest
	Approve
	ClaimFromVault
)

type TaskState int

const (
	StateSuccess TaskState = iota
	StateOngoing
	StateFailed
)
