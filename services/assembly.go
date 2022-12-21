package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
)

type AssemblyMsg struct {
	Stage string
	Task  *types.TransactionTask
}

type AssemblyService struct {
	db     types.IDB
	config *config.Config
}

func NewAssemblyService(db types.IDB, c *config.Config) *AssemblyService {
	return &AssemblyService{
		db:     db,
		config: c,
	}
}

func (c *AssemblyService) AssemblyTx(task *types.TransactionTask) (finished bool, err error) {
	//实际组装tx
	c.handleAssembly(task)

	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		task.State = int(types.TxAssmblyState)
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("Assembly TxCommitWithSession tasks err:%v", err)
	}
	return true, nil
}

func max(nums ...uint64) uint64 {
	var maxNum uint64 = 0
	for _, num := range nums {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}

func (c *AssemblyService) handleAssembly(task *types.TransactionTask) {
	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		logrus.Fatal(err)
	}
	//这里nouce逻辑：1.先查询本地db的nouce，条件为 from ==地址为task.from 2.再从链上取 3.取二者的最大值
	res, err := c.db.GetTaskNonce(task.From)
	if err != nil {
		logrus.Fatal("get tasks for from address:%v err:%v", task.From, err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(task.From))
	if err != nil {
		logrus.Fatal(err)
	}

	task.Nonce = max(nonce, res.Nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
	task.GasPrice = gasPrice.String()
}

func createAssemblyMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("告警:交易组装完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *AssemblyService) tgalert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createAssemblyMsg(task)
	if err != nil {
		logrus.Errorf("create assembly msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
	}

	bot, err := tgbot.NewBot("5985674693:AAF94x_xI2RI69UTP-wt_QThldq-XEKGY8g")
	if err != nil {
		logrus.Fatal(err)
	}
	err = bot.SendMsg(1762573172, msg)
	if err != nil {
		logrus.Fatal(err)
	}
}

func (c *AssemblyService) Run() error {
	tasks, err := c.db.GetOpenedAssemblyTasks()
	if err != nil {
		return fmt.Errorf("get tasks for assembly err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.AssemblyTx(task)
		if err == nil {
			c.tgalert(task)
		}
	}
	return nil
}

func (c AssemblyService) Name() string {
	return "Assembly"
}
