package services

import (
	"bytes"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/go-resty/resty/v2"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"net/http"
)

type FailCallBackService struct {
	db     types.IDB
	config *config.Config
}

func NewFailCallBackService(db types.IDB, c *config.Config) *FailCallBackService {
	return &FailCallBackService{
		db:     db,
		config: c,
	}
}

func (c *FailCallBackService) CallBack(task *types.TransactionTask) (finished bool, err error) {
	//这里回掉
	err = c.handleCallBack(task)
	if err != nil {
		return false, err
	}
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		task.State = int(types.TxFailedState)
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
func (c *FailCallBackService) handleCallBack(task *types.TransactionTask) error {
	//定义相关参数
	url := c.config.CallBack.URL

	cli := resty.New()
	cli.SetBaseURL(url)

	data := types.HttpRes{
		RequestId: task.RequestId,
		Hash:      task.TxHash,
		Code:      400,
		Status:    0,
		Message:   task.Error,
	}

	var result types.HttpRes
	resp, err := cli.R().SetBody(data).SetResult(&result).Post("/v1/callback")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return err
	}
	if result.Code != 0 {
		return err
	}

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func createFailCallBackMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->失败回调完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))
	buffer.WriteString(fmt.Sprintf("TxHash: %v\n\n", task.TxHash))

	return buffer.String(), nil
}

func (c *FailCallBackService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createFailCallBackMsg(task)
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

func (c *FailCallBackService) Run() error {
	tasks, err := c.db.GetOpenedFailCallBackTasks()
	if err != nil {
		return fmt.Errorf("get tasks for assembly err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.CallBack(task)
		if err == nil {
			c.tgAlert(task)
		}
	}
	return nil
}

func (c FailCallBackService) Name() string {
	return "FailCallBack"
}
