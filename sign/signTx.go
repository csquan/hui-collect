package sign

import (
	"github.com/go-delve/delve/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
)

const (
	SignState = iota
	AuditState
	ValidatorState
	FinishState
)

type Signer struct {
	db     types.IDB
	config *config.Config
}

type SignerState = int

func NewTransactionService(db types.IDB, conf *config.Config) (s *Signer, err error) {
	s = &Signer{
		db:     db,
		config: conf,
	}
	return
}

func (signer *Signer) Run() (err error) {
	tasks, err := signer.db.GetOpenedSignTasks()
	if err != nil {
		return
	}
	if len(tasks) == 0 {
		logrus.Infof("no available transfer task.")
		return
	}
	if len(tasks) > 1 {
		logrus.Errorf("more than one sign services are being processed. tasks:%v", tasks)
	}

	switch SignerState(tasks[0].State) {
	case SignState:
		return signer.handleSign(tasks[0])
	case AuditState:
		return signer.handleAudit(tasks[0])
	case ValidatorState:
		return signer.handleValidator(tasks[0])
	default:
		logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
	}
	return
}

func (signer *Signer) handleSign(task *types.SignTask) (err error) {
	//TODO 放在事物中
	input := ""  //temp def
	decimal := 0
	nonce :=0
	from := ""
	to := ""
	GasLimit :=""
	GasPrice :=""
	Amount :=""
	quantity:=""
	receiver:=""

	signRet,err := signer.sign(input, decimal, nonce, from, to, GasLimit, GasPrice, Amount, quantity, receiver)
	if err != nil {
		//err写入db
	}else {
		task.State = int(AuditState)
		task.Cipher = signRet.Data.Extra.Cipher
		task.EncryptData = signRet.Data.EncryptData
		task.TxHash = signRet.Data.Extra.TxHash
		signer.db.UpdateTxTask(task)
	}
	return nil
}

func (signer *Signer) handleAudit(task *types.SignTask) (err error) {
	//TODO 放在事物中
	input := ""  //temp def
	quantity:=""
	receiver:=""
	orderID :=0

	_, err = signer.audit(input,receiver,quantity,orderID)
	if err != nil {
		//写入db
	}else{
		task.State = int(ValidatorState)
		signer.db.UpdateTxTask(task)
	}
	return nil
}

func (signer *Signer) handleValidator(task *types.SignTask) (err error) {
	input := ""  //temp def
	quantity:=""
	orderID :=0
	to := ""

	vRet,err := signer.validator(input, to, quantity,orderID)  //这里检验通过会改写vRet
	if err != nil  {
		//写入db
	}else{
		task.State = int(FinishState)
		task.RawTx = vRet.RawTx
		signer.db.UpdateTxTask(task)
	}
	return
}

