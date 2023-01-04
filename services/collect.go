package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"strconv"
	"time"
)

type CollectService struct {
	db     types.IDB
	config *config.Config
}

func NewCollectService(db types.IDB, c *config.Config) *CollectService {
	return &CollectService{
		db:     db,
		config: c,
	}
}

func getBalance(addr string) (string, error) {
	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return "", err
	}

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), big.NewInt(786760))
	if err != nil {
		return "", err
	}
	return balance.String(), nil
}

func (c *CollectService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createInitMsg(task)
	if err != nil {
		logrus.Errorf("create init msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
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

func createInitMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("检测到待归集交易:->交易初始\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *CollectService) InsertCollectSubTx(from string, to string, userID string, requestID string, chainId string, value string) error {
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

func (c *CollectService) handleAddTx(from string, to string, userID string, requestID string, chainId string, value string) error {
	balance, err := getBalance(from)
	if err != nil {
		return err
	}

	b, err := strconv.ParseFloat(balance, 10)
	if err != nil {
		return err
	}

	if b >= 0.00004 {
		to = "0x32755f0c070811cdd0b00b059e94593fae9835d9" //选择的一个热钱包地址
	} else { //不足以支付一笔交易
		to = from
		userID = "545950000830"
		value = "0x246139CA8000"
		from = "0x32755f0c070811cdd0b00b059e94593fae9835d9" //选择的一个热钱包地址
	}

	c.InsertCollectSubTx(from, to, userID, requestID, chainId, value)
	return nil
}

func (c *CollectService) Run() (err error) {
	collectTasks, err := c.db.GetOpenedCollectTask()
	if err != nil {
		return
	}
	if len(collectTasks) == 0 {
		logrus.Infof("no available collect Transaction task.")
		return
	}

	for _, collectTask := range collectTasks {
		uid := ""
		requestID := ""
		err = c.handleAddTx(collectTask.AddrFrom, collectTask.AddrTo, uid, requestID, "8888", collectTask.Value)

		if err != nil {
			continue
		}

		err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
			collectTask.State = int(types.TxCollectingState)
			if err := c.db.UpdateCollectTx(s, collectTask); err != nil {
				logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("insert colelct sub transactidon task error:%v", err)
		}

	}
	return
}

func (c CollectService) Name() string {
	return "Collect"
}
