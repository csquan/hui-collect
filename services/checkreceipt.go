package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
)

type CheckReceiptService struct {
	db     types.IDB
	config *config.Config
}

func NewCheckReceiptService(db types.IDB, c *config.Config) *CheckReceiptService {
	return &CheckReceiptService{
		db:     db,
		config: c,
	}
}

func (c *CheckReceiptService) CheckReceipt(task *types.TransactionTask) (finished bool, err error) {
	receipt, err := c.handleCheckReceipt(task)
	if err != nil {
		return false, err
	}
	b, err := json.Marshal(receipt)
	if err != nil {
		return false, err
	}
	task.Receipt = string(b)
	task.State = int(types.TxCheckState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in CheckReceipt err:%v", err)
	}
	return true, nil
}

func (c *CheckReceiptService) handleCheckReceipt(task *types.TransactionTask) (*ethtypes.Receipt, error) {
	url := ""
	for _, v := range c.config.Chains {
		if v.ID == task.ChainId {
			url = v.RpcUrl
		}
	}
	if url == "" {
		return nil, fmt.Errorf("nor found chain url match to task.ChainId:%d ", task.ChainId)
	}
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.TxHash))
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (c *CheckReceiptService) tgAlert(task *types.TransactionTask) {
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
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易获取收据完成\n\n"))
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

func (c *CheckReceiptService) Run() error {
	tasks, err := c.db.GetOpenedCheckReceiptTasks()
	if err != nil {
		return fmt.Errorf("get tasks for check receipt err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.CheckReceipt(task)
		if err == nil {
			c.tgAlert(task)
		}
	}
	return nil
}

func (c CheckReceiptService) Name() string {
	return "CheckReceipt"
}
