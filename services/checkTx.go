package services

import (
	"bytes"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"time"
)

type CheckService struct {
	db     types.IDB
	config *config.Config
}

func NewCheckService(db types.IDB, c *config.Config) *CheckService {
	return &CheckService{
		db:     db,
		config: c,
	}
}

func (c *CheckService) InsertCollectSubTx(from string, to string, userID string, requestID string, chainId string, value string) error {
	//插入sub task
	task := types.TransactionTask{
		UUID:      time.Now().Unix(),
		UserID:    userID,
		From:      from,
		To:        to,
		Value:     value,
		ChainId:   8888,
		RequestId: requestID,
	}
	task.State = int(types.TxInitState)

	err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.InsertCollectSubTx(s, &task); err != nil {
			logrus.Errorf("insert colelct sub transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("insert colelct sub transaction task error:%v", err)
	}
	return nil
}

// 根据传入的交易，如果是打gas类型的交易，那么再生成一笔TxInitState状态的交易，如果是归集到热钱包的交易，那么进入下一状态，表示可以更新账本
func (c *CheckService) Check(task *types.TransactionTask) (finished bool, err error) {
	if task.Type == 0 { //说明是打gas交易，需要在交易表中插入一条交易
		dest := "0x32f3323a268155160546504c45d0c4a832567159"
		UID := ""
		Value := "10000000000000000000"
		c.InsertCollectSubTx(task.To, dest, UID, "", "8888", Value)
		task.State = int(types.TxEndState)
	} else { //说明已经是归集交易，进入下一状态
		task.State = int(types.TxCheckState)
	}

	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in Check err:%v", err)
	}

	return true, nil
}

func (c *CheckService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createCheckMsg(task)
	if err != nil {
		logrus.Errorf("create assembly msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
	}

	bot, err := tgbot.NewBot("5904746042:AAGjBMN_ahQ0uavSCakrEFUN7RV2Q8oDY4I")
	if err != nil {
		logrus.Fatal(err)
	}
	err = bot.SendMsg(-1001731474163, msg)
	if err != nil {
		logrus.Fatal(err)
	}
}
func createCheckMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易广播完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("SignHash: %v\n\n", task.SignHash))
	buffer.WriteString(fmt.Sprintf("TxHash: %v\n\n", task.TxHash))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *CheckService) Run() error {
	tasks, err := c.db.GetOpenedCheckTasks()
	if err != nil {
		return fmt.Errorf("get tasks for check err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.Check(task)
		if err == nil {
			//c.tgAlert(task)
		}
	}
	return nil
}

func (c CheckService) Name() string {
	return "Check"
}
