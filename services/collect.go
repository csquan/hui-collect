package services

import (
	"bytes"
	"context"
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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
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

func (c *CollectService) AssemblyTx(task *types.TransactionTask) (finished bool, err error) {
	//实际组装tx
	c.handleAssembly(task)

	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		task.State = int(types.TxAssmblyState)
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

func max(nums ...uint64) uint64 {
	var maxNum uint64 = 0
	for _, num := range nums {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
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
	buffer.WriteString(fmt.Sprintf("交易状态变化:->交易初始\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

// 接收注册过来的消息，存入db作为tx初始状态
func (c *CollectService) AddTask(from string, to string, inputData string, userID string, requestID string, chainId string, value string) {
	//插入task
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

	err := utils.CommitWithSession(c.db, func(session *xorm.Session) (execErr error) {
		if err := c.db.SaveTxTask(session, &task); err != nil {
			logrus.Errorf("save transaction task error:%v tasks:[%v]", err, task)
			return
		}
		c.tgAlert(&task)
		return
	})
	if err != nil {
	}

}
func (c *CollectService) handleAssembly(task *types.TransactionTask) error {
	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		logrus.Fatal(err)
	}

	from := task.From
	//先看这笔交易的from地址有没有钱，没有钱搞钱
	balance, err := getBalance(from)
	if err != nil {

	}

	parseInt, err := strconv.ParseFloat(balance, 10)
	if err != nil {
		return err
	}

	if parseInt <= 0.00004 { //没有足够钱支付一笔基础转账

	} else { //

	}
	//然后构造第二笔交易，

	//这里nouce逻辑：1.先查询本地db的nouce，条件为 from ==地址为task.from 2.再从链上取 3.取二者的最大值
	res, err := c.db.GetTaskNonce(task.From)
	if err != nil {
		logrus.Fatal("get tasks for from address:%v err:%v", task.From, err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(task.From))
	if err != nil {
		logrus.Fatal(err)
	}

	task.Nonce = max(nonce, res.Nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
	task.GasPrice = gasPrice.String()

	//gasPrice, err := client.EstimateGas(context.Background())
	//if err != nil {
	//	logrus.Fatal(err)
	//}
	task.GasLimit = "8000000"
	return nil
}

func createAssemblyMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易组装完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func Hex2Dec(val string) (int, error) {
	n, err := strconv.ParseUint(val, 16, 32)
	if err != nil {
		return 0, err
	}
	return int(n), nil
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

func (c *CollectService) handleSign(task *types.TransactionTask) error {
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return err
	}

	to := common.HexToAddress(task.To)

	value, err := Hex2Dec(task.Value[2:])
	if err != nil {
		return err
	}

	b, err := hex.DecodeString(task.InputData[2:])

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
		return err
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
		return fmt.Errorf("add sign tasks err:%v", err)
	}
	return nil
}

func (c *CollectService) handleBroadcast(task *types.TransactionTask) error {
	hash, err := c.handleBroadcastTx(task)
	if err != nil {
		return err
	}
	task.TxHash = hash //广播的交易hash
	task.State = int(types.TxBroadcastState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf(" CommitWithSession in BroadcastTx err:%v", err)
	}
	return nil
}

func (c *CollectService) handleBroadcastTx(task *types.TransactionTask) (string, error) {
	//将签名数据组装
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return "", err
	}

	to := common.HexToAddress(task.To)

	value, err := Hex2Dec(task.Value[2:])
	if err != nil {
		return "", err
	}

	b, err := hex.DecodeString(task.InputData[2:])
	if err != nil {
		return "", err
	}
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    task.Nonce,
		GasPrice: big.NewInt(gasPrice),
		Gas:      8000000,
		To:       &to,
		Value:    big.NewInt(int64(value)),
		Data:     b,
	})

	var sigData types.SigData
	err = json.Unmarshal([]byte(task.Sig), &sigData)

	remoteSig, err := hex.DecodeString(sigData.Signature)

	signer := ethTypes.NewEIP155Signer(big.NewInt(int64(task.ChainId)))
	sigedTx, err := tx.WithSignature(signer, remoteSig)

	if err != nil {
		return "", err
	}

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return "", err
	}

	from, err := ethTypes.Sender(signer, tx)

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(task.From), big.NewInt(786760))
	if err != nil {
		return "", err
	}
	fmt.Printf("insert tx from balance is %s\n", balance.String())

	err = client.SendTransaction(context.Background(), sigedTx)
	if err != nil {
		fmt.Printf("signature from address  ：" + from.Hex() + " send err:" + err.Error() + "\n")
		task.Error = err.Error()
		return "", err
	}

	fmt.Printf("tx sent: %s", sigedTx.Hash())

	return sigedTx.Hash().Hex(), nil

}

func (c *CollectService) handleInsertOrUpdateAccount(task *types.TransactionTask) error {
	//receipt, err := c.handleCheckReceipt(task)
	//if err != nil {
	//	return err
	//}
	//b, err := json.Marshal(receipt)
	//if err != nil {
	//	return err
	//}
	//task.Receipt = string(b)
	//task.State = int(types.TxCheckState)
	//err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
	//	if err := c.db.UpdateTransactionTask(s, task); err != nil {
	//		logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
	//		return err
	//	}
	//	return nil
	//})
	//if err != nil {
	//	return fmt.Errorf(" CommitWithSession in CheckReceipt err:%v", err)
	//}
	return nil
}

func (c *CollectService) handleTransactionCheck(task *types.TransactionTask) error {
	receipt, err := c.handleCheckReceipt(task)
	if err != nil {
		return err
	}
	b, err := json.Marshal(receipt)
	if err != nil {
		return err
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
		return fmt.Errorf(" CommitWithSession in CheckReceipt err:%v", err)
	}
	return nil
}

func (c *CollectService) handleCheckReceipt(task *types.TransactionTask) (*ethTypes.Receipt, error) {
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

func (c *CollectService) Run() (err error) {
	tasks, err := c.db.GetOpenedTransactionTask()
	if err != nil {
		return
	}
	if len(tasks) == 0 {
		logrus.Infof("no available part Transaction task.")
		return
	}

	for _, task := range tasks {
		err = func(task *types.TransactionTask) error {
			switch types.TransactionState(task.State) {
			case types.TxInitState:
				return c.handleAssembly(task)
			case types.TxAssmblyState:
				return c.handleSign(task)
			case types.TxSignState:
				return c.handleBroadcast(task)
			case types.TxBroadcastState:
				return c.handleTransactionCheck(task)
			case types.TxCheckState:
				return c.handleInsertOrUpdateAccount(task)
			default:
				logrus.Errorf("unkonwn task state [%v] for task [%v]", tasks[0].State, tasks[0].ID)
				return nil
			}
		}(task)

		if err != nil {
			c.db.UpdateTransactionTaskState(task.ID, 6)
			return
		}
	}
	return
}

func (c CollectService) Name() string {
	return "Collect"
}
