package services

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"strings"
	"time"
)

const erc20abi = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

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

func (c *CheckService) InsertCollectSubTx(parentID uint64, from string, to string, userID string, requestID string, chainId int, inputdata string, value string, tx_type int) error {
	//插入sub task
	task := types.TransactionTask{
		ParentID:  parentID,
		UUID:      time.Now().Unix(),
		UserID:    userID,
		From:      from,
		To:        to,
		InputData: inputdata,
		ChainId:   chainId,
		RequestId: requestID,
		Tx_type:   tx_type,
		Value:     value,
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
	if task.Tx_type == 0 { //说明是打gas交易，需要在交易表中插入一条归集交易
		dest := c.config.Collect.Addr //"0x32f3323a268155160546504c45d0c4a832567159"
		src_task, err := c.db.GetCollectTask(task.ParentID)
		if err != nil {
			return false, err
		}
		to := src_task.Addr // "0x99ac689fd1f09ada4c0365e6497b2a824af68557" 这笔源交易对应的合约地址
		UID := "817583340974"

		r := strings.NewReader(erc20abi)
		erc20ABI, err := abi.JSON(r)
		if err != nil {
			return false, err
		}
		//src_task.TokenCnt = "1000000000000000000000"
		//得到关联交易的value
		Amount := &big.Int{}
		Amount.SetString(src_task.TokenCnt, 10)

		b, err := erc20ABI.Pack("transfer", common.HexToAddress(dest), Amount)
		if err != nil {
			return false, err
		}
		inputdata := hex.EncodeToString(b)

		value := "0x0" //这里应该查询这笔gas交易对应的源交易value是多少
		c.InsertCollectSubTx(task.ParentID, task.To, to, UID, "", 8888, "0x"+inputdata, value, 1)
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
