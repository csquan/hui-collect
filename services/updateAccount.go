package services

import (
	"bytes"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
)

type UpdateAccountService struct {
	db     types.IDB
	config *config.Config
}

func NewUpdateAccountService(db types.IDB, c *config.Config) *UpdateAccountService {
	return &UpdateAccountService{
		db:     db,
		config: c,
	}
}

func (c *UpdateAccountService) UpdateAccount(task *types.TransactionTask) (finished bool, err error) {
	task.State = int(types.TxEndState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil { //更新源归集子交易的状态
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		if err := c.db.UpdateCollectTxState(uint64(task.ParentID), int(types.TxCollectedState)); err != nil { //更新归集源交易的状态
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}

		//这里先取出余额，累加task.Amount
		account, err := c.db.GetAccountBalance(task.Receiver, task.ContractAddr)
		if err != nil { //查询账本
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}

		if account == nil { //说明不存在,直接插入
			account1 := types.Account{
				Addr:         task.Receiver,
				ContractAddr: task.ContractAddr,
				Balance:      task.Amount,
			}
			if err := c.db.SaveAccount(s, &account1); err != nil { //更新账本
				logrus.Errorf("save account task error:%v tasks:[%v]", err, account)
				return err
			}
		} else { //更新账本
			cnt1, _ := big.NewInt(0).SetString(account.Balance, 10)
			cnt2, _ := big.NewInt(0).SetString(task.Amount, 10)

			res := big.NewInt(0).Add(cnt1, cnt2)

			if err := c.db.UpdateAccount(res.String(), task.Receiver, task.ContractAddr); err != nil {
				logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in BroadcastTx err:%v", err)
	}
	return true, nil
}

func (c *UpdateAccountService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createUpdateMsg(task)
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
func createUpdateMsg(task *types.TransactionTask) (string, error) {
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

func (c *UpdateAccountService) Run() error {
	tasks, err := c.db.GetOpenedUpdateAccountTasks()
	if err != nil {
		return fmt.Errorf("get tasks for update account err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.UpdateAccount(task)
		if err == nil {
			//c.tgAlert(task)
		}
	}
	return nil
}

func (c UpdateAccountService) Name() string {
	return "UpdateAccount"
}
