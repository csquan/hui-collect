package services

import (
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"unsafe"
)

const SigLen = 65

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

func UnmarshalP256(hexkey string) (*ecies.PrivateKey, error) {
	priv, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.P256().ScalarBaseMult(priv)

	return &ecies.PrivateKey{
		PublicKey: ecies.PublicKey{
			X:      x,
			Y:      y,
			Curve:  elliptic.P256(),
			Params: ecies.ParamsFromCurve(elliptic.P256()),
		},
		D: big.NewInt(0).SetBytes(priv),
	}, nil
}

func UnmarshalP256CompressedPub(hexkey string) (*ecies.PublicKey, error) {
	pb, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pb)
	return &ecies.PublicKey{
		X:      x,
		Y:      y,
		Curve:  elliptic.P256(),
		Params: ecies.ParamsFromCurve(elliptic.P256()),
	}, nil
}

func Hex2Dec(val string) (int, error) {
	n, err := strconv.ParseUint(val, 16, 32)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func stringTobyteSlice(s string) []byte {
	tmp1 := (*[2]uintptr)(unsafe.Pointer(&s))
	tmp2 := [3]uintptr{tmp1[0], tmp1[1], tmp1[1]}
	return *(*[]byte)(unsafe.Pointer(&tmp2))

}

func stringTobyteSliceOld(s string) []byte {
	return []byte(s)

}

func byteSliceToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func (c *SignService) SignTx(task *types.TransactionTask) (finished bool, err error) {
	//gasLimit, err := strconv.ParseUint(task.GasLimit, 10, 64)
	//if err != nil {
	//	return false, err
	//}
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return false, err
	}

	to := common.HexToAddress(task.To)

	value, err := Hex2Dec(task.Value[2:])
	if err != nil {
		return false, err
	}

	b1 := stringTobyteSlice(task.InputData)
	b2 := stringTobyteSliceOld(task.InputData)
	b, err := hex.DecodeString(task.InputData[2:])

	fmt.Println(b1, b2)
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    task.Nonce,
		GasPrice: big.NewInt(gasPrice),
		Gas:      8000000,
		To:       &to,
		Value:    big.NewInt(int64(value)),
		Data:     b,
	})

	pubKey, err := UnmarshalP256CompressedPub("0209674d59b772b17524ec19bfc407c66547f8ff332c5e0a2097e8a3c36de09814")

	signer := ethTypes.NewEIP155Signer(big.NewInt(int64(task.ChainId)))
	signHash := signer.Hash(tx)

	task.SignHash = signHash.Hex() //这里存储的是计算出来的签名前的hash

	signData := types.SignData{
		UID:     task.UserID,
		Address: task.From,
		Hash:    signHash.Hex(),
	}

	msg, err := json.Marshal(signData)
	ct, err := ecies.Encrypt(rand.Reader, pubKey, msg, nil, nil)

	if err != nil {
		return false, err
	}
	fmt.Printf(hex.EncodeToString(ct))

	signurl := "http://15.152.203.71:8080/sign"

	postValue := url.Values{
		"data": {hex.EncodeToString(ct)},
	}
	resp, err := http.PostForm(signurl, postValue)
	resp.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	sig := string(content)

	isValid := gjson.Valid(sig)
	if isValid == false {
		fmt.Println("Fatal error ", err.Error())
	}
	task.Sig = sig
	//res := gjson.Get(sig, "signature")
	//signature := res.String()
	////32 32 1-->R S V
	//if len(signature)/2 != SigLen { //这里记录错误，更新数据库，也返回true,留待下次扫到这个交易重试签名
	//	task.Error = fmt.Sprintf("signature len :%d is error", len(sig))
	//} else {
	//	signature = "0x" + signature
	//	b, err := hexutil.Decode(signature)
	//	if err != nil {
	//		fmt.Println("Fatal error ", err.Error())
	//	}
	//	task.Signature = b
	//}
	task.State = int(types.TxSignState)
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
	buffer.WriteString(fmt.Sprintf("Signature: %v\n\n", task.Sig))
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
