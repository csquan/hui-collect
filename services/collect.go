package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/pkg/util/ecies"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"net/http"
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

func (c *CollectService) getBalance(addr string, chainName string) (string, error) {
	client, err := ethclient.Dial(c.config.Chains[chainName].RpcUrl)
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

func (c *CollectService) InsertCollectSubTx(parentIDs string, from string, to string, userID string, requestID string, chainName string, inputdata string, value string, tx_type int, receiver string, amount string, contractAddr string) error {
	//插入sub task
	task := types.TransactionTask{
		ParentIDs:    parentIDs,
		UUID:         time.Now().Unix(),
		UserID:       userID,
		From:         from,
		To:           to,
		ContractAddr: contractAddr,
		Value:        value,
		InputData:    inputdata,
		Chain:        chainName,
		RequestId:    requestID,
		Tx_type:      tx_type,
		Receiver:     receiver,
		Amount:       amount,
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

func (c *CollectService) getUidFromAddr(address string) (uid string, err error) {
	pubKey, err1 := ecies.PublicFromString(c.config.UserInfo.KycPubKey)
	if err1 != nil {
		logrus.Println(err)
	}

	cli := resty.New()
	cli.SetBaseURL(c.config.UserInfo.URL)

	nowStr := time.Now().UTC().Format(http.TimeFormat)
	ct, err1 := ecies.Encrypt(rand.Reader, pubKey, []byte(nowStr), nil, nil)
	if err1 != nil {
		logrus.Println(err1)
	}
	data := map[string]interface{}{
		"verified": hex.EncodeToString(ct),
		"addr":     address,
	}
	var result types.HttpData
	resp, er := cli.R().SetBody(data).SetResult(&result).Post("/api/v1/pub/i-q-user-by-addr")
	if er != nil {
		logrus.Println(err)
	}
	if resp.StatusCode() != http.StatusOK {
		logrus.Println(err)
	}
	if result.Code != 0 {
		logrus.Println(err)
	}

	return result.Data.UID, nil
}

func (c *CollectService) handleAddTx(parentIDs string, from string, to string, userID string, requestID string, chainName string, tokencnt string, contractAddr string) error {
	balance, err := c.getBalance(to, chainName)
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
	receiver := ""
	amount := ""
	if b >= max_tx_fee { //插入一笔归集子交易
		uid, err := c.getUidFromAddr(to)
		if err != nil {
			logrus.Info("empty get!")
		}
		userID = uid //"817583340974" // 0x206beddf4f9fc55a116890bb74c6b79999b14eb1
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

		receiver = c.config.Collect.Addr //receiver 就是归集地址
		amount = Amount.String()

		b, err := erc20ABI.Pack("transfer", common.HexToAddress(receiver), Amount)
		if err != nil {
			return err
		}
		inputdata = hex.EncodeToString(b)

	} else { //不足以支付一笔交易
		//userID = "545950000830"
		value = "0x246139CA8000"
		from = c.config.Gas.Addr //"0x32755f0c070811cdd0b00b059e94593fae9835d9"
		receiver = to            //receiver 就是源to地址，这里先给它打gas
		amount = value
		userID, err = c.getUidFromAddr(from)
		if err != nil {

		}
		tx_type = 0
	}
	c.InsertCollectSubTx(parentIDs, from, to, userID, requestID, chainName, "0x"+inputdata, value, tx_type, receiver, amount, contractAddr)
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

	merge_tasks := make([]*types.CollectTxDB, 0)     //多条相同的交易合并（相同的接收地址和相同的合约地址）
	threshold_tasks := make([]*types.CollectTxDB, 0) //交易是否满足门槛

	//这里如果有多条collectTask，那么需要归并到一起，依据规则：将相同合约地址,相同receiver,相同chain的 tokencnt累加
	for _, task := range collectTasks {
		found := false
		for _, filter_task := range merge_tasks {
			if filter_task.Addr == task.Addr && filter_task.Receiver == task.Receiver && filter_task.Chain == task.Chain {
				cnt1, _ := big.NewInt(0).SetString(task.TokenCnt, 10)
				cnt2, _ := big.NewInt(0).SetString(filter_task.TokenCnt, 10)

				res := big.NewInt(0).Add(cnt1, cnt2)
				filter_task.TokenCnt = res.String()

				found = true
			}
		}
		if found == false {
			merge_tasks = append(merge_tasks, task)
		}
	}

	//这里归并后，应该看相同地址的是否大于对应币种的门槛
	for _, merge_task := range merge_tasks {
		token, err := c.db.GetTokenInfo(merge_task.Addr, merge_task.Chain)

		if err != nil {
			logrus.Fatal(err)
		}

		cnt1, _ := big.NewInt(0).SetString(merge_task.TokenCnt, 10)
		cnt2, _ := big.NewInt(0).SetString(token.Threshold, 10)

		enough := cnt1.Cmp(cnt2)

		if enough >= 0 {
			threshold_tasks = append(threshold_tasks, merge_task)
		}
	}

	parentIDs := ""

	for _, threshold_task := range threshold_tasks {
		parentIDs = parentIDs + "," + strconv.Itoa(int(threshold_task.ID))
	}
	logrus.Info("\n parentIDs:", parentIDs)

	if len(parentIDs) > 1 {
		if parentIDs[0] == 44 { //去除前面的逗号，ASCII值为44
			parentIDs = parentIDs[1:]
		}
	}

	for _, collectTask := range threshold_tasks {
		uid := "" //这个后面填入，根据不同的交易
		requestID := ""
		err = c.handleAddTx(parentIDs, collectTask.Sender, collectTask.Receiver, uid, requestID, collectTask.Chain, collectTask.TokenCnt, collectTask.Addr)

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
