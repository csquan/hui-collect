package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"strconv"
)

type SignService struct {
	db     types.IDB
	config *config.Config
}

func NewSignService(db types.IDB, c *config.Config) *SignService {
	return &SignService{
		db:     db,
		config: c,
	}
}

func (c *SignService) SignTx(task *types.TransactionTask) (finished bool, err error) {
	gasLimit, err := strconv.ParseUint(task.GasLimit, 10, 64)
	if err != nil {
		return false, err
	}
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return false, err
	}

	tx := ethtypes.NewTransaction(task.Nonce, common.HexToAddress(task.To), big.NewInt(0), gasLimit, big.NewInt(gasPrice), []byte(task.InputData))

	walletPrivateKey, err := crypto.HexToECDSA("35f47c090146782715561fc6df6033167a72660ae1386fab0d0a6f5a1db3d18f")

	task.Hash = tx.Hash().Hex() //这里存储的是计算出来的hash，而不是广播的hash，广播后会有比对

	signData := types.SignData{
		UID:     task.UserID,
		Address: task.From,
		Hash:    tx.Hash().String(),
	}

	msg, err := json.Marshal(signData)
	cipherPubKey := ecies.ImportECDSAPublic(&walletPrivateKey.PublicKey)

	_, err = ecies.Encrypt(rand.Reader, cipherPubKey, msg, nil, nil)

	if err != nil {
		return false, err
	}

	//将ct发送给签名接口

	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		// 依照结果更新task状态
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("add sign tasks err:%v", err)
	}
	return true, nil
}

func (c *SignService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createSignMsg(task)
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
func createSignMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易签名完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("Hash: %v\n\n", task.Hash))
	buffer.WriteString(fmt.Sprintf("SignData: %v\n\n", task.SignData))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *SignService) Run() error {
	tasks, err := c.db.GetOpenedSignTasks()
	if err != nil {
		return fmt.Errorf("get tasks for sign err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.SignTx(task)
		if err == nil {
			c.tgAlert(task)
		}
	}
	return nil
}

func (c SignService) Name() string {
	return "Sign"
}
