package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"unsafe"
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

//func (c *SignService) WithSignature(tx *ethTypes.Transaction, signer ethTypes.Signer, sig []byte) (*ethTypes.Transaction, error) {
//	r, s, v, err := signer.SignatureValues(tx, sig)
//	if err != nil {
//		return nil, err
//	}
//	cpy := tx.inner.copy()
//	cpy.setSignatureValues(signer.ChainID(), v, r, s)
//	return &ethTypes.Transaction{inner: cpy, time: tx.time}, nil
//}

func (c *SignService) SignTx(task *types.TransactionTask) (finished bool, err error) {
	gasLimit, err := strconv.ParseUint(task.GasLimit, 10, 64)
	if err != nil {
		return false, err
	}
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return false, err
	}

	tx := ethTypes.NewTransaction(task.Nonce, common.HexToAddress(task.To), big.NewInt(0), gasLimit, big.NewInt(gasPrice), []byte(task.InputData))

	walletPrivateKey, err := crypto.HexToECDSA("35f47c090146782715561fc6df6033167a72660ae1386fab0d0a6f5a1db3d18f")

	task.SignHash = tx.Hash().Hex() //这里存储的是计算出来的签名前的hash

	signData := types.SignData{
		UID:     task.UserID,
		Address: task.From,
		Hash:    tx.Hash().String(),
	}

	msg, err := json.Marshal(signData)
	cipherPubKey := ecies.ImportECDSAPublic(&walletPrivateKey.PublicKey)

	ct, err := ecies.Encrypt(rand.Reader, cipherPubKey, msg, nil, nil)

	if err != nil {
		return false, err
	}

	//将ct发送给签名接口
	res, err := http.Post("http://15.152.203.71:8080/sign",
		"application/json;charset=utf-8", bytes.NewBuffer(ct))
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	sig := (*string)(unsafe.Pointer(&content)) //转化为string,优化内存

	//将sig中rsv三个持久化
	fmt.Println(sig)

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
	buffer.WriteString(fmt.Sprintf("SignHash: %v\n\n", task.SignHash))
	buffer.WriteString(fmt.Sprintf("TxHash: %v\n\n", task.TxHash))
	buffer.WriteString(fmt.Sprintf("Sign_R: %v\n\n", task.R))
	buffer.WriteString(fmt.Sprintf("Sign_S: %v\n\n", task.S))
	buffer.WriteString(fmt.Sprintf("Sign_V: %v\n\n", task.V))
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
