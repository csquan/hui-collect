package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/pkg/util/ecies"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"github.com/tidwall/gjson"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const max_tx_fee = 0 //400000000000000 //4*10 14 认为是一笔交易的费用

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

func (c *CollectService) GetTokenInfo(symbol string, chain string) (string, error) {
	tokenParam := types.TokenParam{
		Chain:  chain,
		Symbol: symbol,
	}
	msg, err1 := json.Marshal(tokenParam)
	if err1 != nil {
		logrus.Error(err1)
	}
	url := c.config.Token.Url + "/" + "getToken"

	str, err := utils.Post(url, msg)
	if err != nil {
		logrus.Error(err1)
		return "", err
	}
	return str, nil
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
			if filter_task.Address == task.Address && filter_task.Symbol == task.Symbol && filter_task.Chain == task.Chain {
				cnt1, _ := big.NewInt(0).SetString(task.Balance, 10)
				cnt2, _ := big.NewInt(0).SetString(filter_task.Balance, 10)

				res := big.NewInt(0).Add(cnt1, cnt2)
				filter_task.Balance = res.String()

				found = true
			}
		}
		if found == false {
			merge_tasks = append(merge_tasks, task)
		}
	}

	//这里归并后，应该看相同地址的是否大于对应币种的门槛--只看本币
	for _, merge_task := range merge_tasks {
		str, err := c.GetTokenInfo(merge_task.Symbol, merge_task.Chain)

		if err != nil {
			logrus.Fatal(err)
		}

		collect_threshold := gjson.Get(str, "collect_threshold")

		cnt1, _ := big.NewFloat(0).SetString(merge_task.Balance)
		cnt2, _ := big.NewFloat(0).SetString(collect_threshold.String())

		logrus.Info(cnt1.String(), cnt2.String())

		enough := cnt1.Cmp(cnt2)

		if enough >= 0 {
			threshold_tasks = append(threshold_tasks, merge_task)
		}
	}

	for _, collectTask := range threshold_tasks {
		balance, err := strconv.Atoi(collectTask.Balance)
		if err != nil {
			logrus.Error(err)
		}
		tokenStr, err := c.GetTokenInfo(collectTask.Symbol, collectTask.Chain)
		if err != nil {
			logrus.Error(err)
		}

		if collectTask.Symbol == "hui" && balance < max_tx_fee { //反向打gas--fundFee 钱包模块
			//gas--getToken token模块
			fee_value := gjson.Get(tokenStr, "give_fee_value")

			fund := types.Fund{
				AppId:     "",
				OrderId:   utils.NewIDGenerator().Generate(),
				AccountId: collectTask.Uid,
				Chain:     collectTask.Chain,
				Symbol:    collectTask.Symbol,
				To:        collectTask.Address,
				Amount:    fee_value.String(),
			}
			msg, err := json.Marshal(fund)
			if err != nil {
				logrus.Error(err)
				continue
			}
			url := c.config.Wallet.Url + "/" + "getAsset"
			str, err := utils.Post(url, msg)
			if err != nil {
				logrus.Error(err)
				continue
			}
			logrus.Info(str)
			//返回200
		} else { //直接归集个人地址--订单ID，插入DB中，目前仅仅是查看标志状态用
			err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
				//这里要按照一定策略选择热钱包目标地址
				to, err := utils.GetHotAddress(collectTask, c.config.HotWallet, c.config.Wallet.Url)
				if err != nil {
					logrus.Error(err)
					return err
				}

				collectRemain := gjson.Get(tokenStr, "collect_remain")
				balance, _ := big.NewFloat(0).SetString(collectTask.Balance)
				remain, _ := big.NewFloat(0).SetString(collectRemain.String())

				balance = balance.Sub(balance, remain)

				logrus.Info(balance.String())

				collectTask.OrderId = utils.NewIDGenerator().Generate()
				//这里调用keep的归集交易接口  --collenttohotwallet
				fund := types.Fund{
					AppId:     "",
					OrderId:   collectTask.OrderId,
					AccountId: collectTask.Uid,
					Chain:     collectTask.Chain,
					Symbol:    collectTask.Symbol,
					From:      collectTask.Address,
					To:        to, //这里要按照一定策略选择热钱包
					Amount:    "9",
				}

				msg, err := json.Marshal(fund)
				if err != nil {
					logrus.Error(err)
					return err
				}
				url := c.config.Wallet.Url + "/" + "collectToHotWallet"
				str, err := utils.Post(url, msg)
				if err != nil {
					logrus.Error(err)
					return err
				}
				logrus.Info(str)

				if err := c.db.UpdateCollectTx(s, collectTask); err != nil {
					logrus.Errorf("update colelct transaction task error:%v tasks:[%v]", err, collectTask)
					return err
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("insert colelct sub transaction task error:%v", err)
			}
		}

		if err != nil {
			continue
		}

	}
	return
}

func (c CollectService) Name() string {
	return "Collect"
}
