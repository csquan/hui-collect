package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const max_tx_fee = 400000000000000 //4*10 14 认为是一笔交易的费用
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
	client, err := ethclient.Dial("http://54.169.11.46:8545")
	if err != nil {
		return "", err
	}

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
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

func (c *CollectService) InsertCollectSubTx(parentID uint64, from string, to string, userID string, requestID string, chainId string, inputdata string, value string, tx_type int) error {
	//插入sub task
	task := types.TransactionTask{
		ParentID:  parentID,
		UUID:      time.Now().Unix(),
		UserID:    userID,
		From:      from,
		To:        to,
		Value:     value,
		InputData: inputdata,
		ChainId:   8888,
		RequestId: requestID,
		Tx_type:   tx_type,
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

func (c *CollectService) handleAddTx(parentID uint64, from string, to string, userID string, requestID string, chainId string, tokencnt string, contractAddr string) error {
	balance, err := getBalance(to)
	if err != nil {
		return err
	}

	b, err := strconv.ParseFloat(balance, 10)
	if err != nil {
		return err
	}

	tx_type := 0
	inputdata := ""
	value := "0x0"
	if b >= max_tx_fee { //插入一笔归集子交易
		userID = "817583340974" // 0x206beddf4f9fc55a116890bb74c6b79999b14eb1
		from = to
		to = contractAddr
		tx_type = 1

		r := strings.NewReader(erc20abi)
		erc20ABI, err := abi.JSON(r)
		if err != nil {
			return err
		}
		Amount := &big.Int{}
		Amount.SetString(tokencnt, 10)

		dest := c.config.Collect.Addr
		b, err := erc20ABI.Pack("transfer", common.HexToAddress(dest), Amount)
		if err != nil {
			return err
		}
		inputdata = hex.EncodeToString(b)

	} else { //不足以支付一笔交易
		userID = "545950000830"
		value = "0x246139CA8000"
		from = c.config.Gas.Addr //"0x32755f0c070811cdd0b00b059e94593fae9835d9"
		tx_type = 0
	}

	c.InsertCollectSubTx(parentID, from, to, userID, requestID, chainId, "0x"+inputdata, value, tx_type)
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
		parentID := collectTask.Id
		err = c.handleAddTx(parentID, collectTask.Sender, collectTask.Receiver, uid, requestID, "8888", collectTask.TokenCnt, collectTask.Addr)

		if err != nil {
			continue
		}

		err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
			collectTask.CollectState = int(types.TxCollectingState)
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
