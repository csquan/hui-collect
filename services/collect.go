package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"github.com/tidwall/gjson"
	"math/big"
	"strings"
	"time"
)

const max_tx_fee = "5000000000000000" //4*10 15 认为是一笔交易的费用

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

func (c *CollectService) SubBlack(hotAddrs []string, blackAddrs []string) ([]string, error) {
	var addrs []string

	for _, hotValue := range hotAddrs {
		if len(blackAddrs) == 0 {
			hotValue = hotValue[1 : len(hotValue)-1]
			addrs = append(addrs, hotValue)
		}
		for _, blackValue := range blackAddrs {
			if blackValue != hotValue {
				hotValue = hotValue[1 : len(hotValue)-1]
				addrs = append(addrs, hotValue)
			}
		}
	}
	return addrs, nil
}

func (c *CollectService) GetHotWallet(str string) ([]string, error) {
	str = str[1 : len(str)-1]
	arr := strings.Split(str, ",")
	return arr, nil
}

func (c *CollectService) Run() (err error) {
	srcTasks, err := c.db.GetOpenedCollectTask()
	if err != nil {
		return
	}
	if len(srcTasks) == 0 {
		logrus.Infof("no available collect Transaction task.")
		return
	}
	logrus.Info("开始新一轮归集，得到可供归集的原始交易:")
	for _, tx := range srcTasks {
		str := fmt.Sprintf("ID:%d ", tx.ID)
		logrus.Info(str + "addr:" + tx.Address + "balance:" + tx.Balance)
	}

	logrus.Info("过滤余额为0的源交易:")
	var collectTasks []*types.CollectTxDB
	for _, task := range srcTasks {
		if task.Balance != "0" {
			collectTasks = append(collectTasks, task)
		}
	}

	logrus.Info("排除余额为0的交易后的归集源交易:")
	for _, tx := range collectTasks {
		str := fmt.Sprintf("ID:%d ", tx.ID)
		logrus.Info(str + "addr:" + tx.Address + "balance:" + tx.Balance)
	}

	mergeTasks := make([]*types.CollectTxDB, 0)      //多条相同的交易合并（相同的接收地址和相同的合约地址）
	threshold_tasks := make([]*types.CollectTxDB, 0) //交易是否满足门槛
	hotWallets := make(map[string]map[string][]string)

	//这里如果有多条collectTask，那么需要归并到一起，依据规则：将相同合约地址,相同receiver,相同chain的 tokencnt累加
	for _, task := range collectTasks {
		found := false
		for _, filterTask := range mergeTasks {
			if filterTask.Address == task.Address && filterTask.Symbol == task.Symbol && filterTask.Chain == task.Chain {
				cnt1, _ := big.NewFloat(0).SetString(task.Balance)
				cnt2, _ := big.NewFloat(0).SetString(filterTask.Balance)

				res := big.NewFloat(0).Add(cnt1, cnt2)
				logrus.Info("merge chain: " + task.Chain + "symbol: " + task.Symbol + "addr: " + task.Address + "balance1:" + cnt1.String() + "+ balance2:" + cnt2.String() + "=" + res.String())
				filterTask.Balance = res.String()

				found = true
			}
		}
		if found == false {
			mergeTasks = append(mergeTasks, task)
		}
	}
	logrus.Info("为了节约gas，合并相同地址、代币、链的交易:")
	for _, mergeTask := range mergeTasks {
		logrus.Info(mergeTask.Address, mergeTask.Chain, mergeTask.Symbol)
	}

	//这里归并后，应该看相同地址的是否大于对应币种的门槛--只看本币
	for _, mergeTask := range mergeTasks {
		logrus.Info("开始调用GetTokenInfo")
		logrus.Info(mergeTask.Symbol + mergeTask.Symbol)
		str, err := c.GetTokenInfo(mergeTask.Symbol, mergeTask.Chain)

		if err != nil {
			logrus.Error(err)
			continue
		}
		collectThreshold := gjson.Get(str, "collect_threshold")
		hotWallet := gjson.Get(str, "hot_wallets")
		blacklist := gjson.Get(str, "blacklist")

		hotAddrs, err := c.GetHotWallet(hotWallet.String())
		if err != nil {
			logrus.Error(err)
			continue
		}
		logrus.Info("symbol: " + mergeTask.Symbol + " chain: " + mergeTask.Chain)
		logrus.Info("热钱包:")
		logrus.Info(hotAddrs)

		for _, hotAddr := range hotAddrs {
			logrus.Info(hotAddr)
		}

		logrus.Info("黑名单钱包:")
		logrus.Info(blacklist)

		var blackAddrs []string
		if blacklist.String() != "" {
			blackAddrs, err = c.GetHotWallet(blacklist.String())
			if err != nil {
				logrus.Error(err)
				continue
			}
		}

		//这里删除热钱包和黑名单相同地址的交易--todo：黑名单移到监控
		for _, blackAddr := range blackAddrs {
			if len(blackAddr) > 1 {
				blackAddr = blackAddr[1 : len(blackAddr)-1]
				logrus.Info("黑名单地址: " + blackAddr)
				if blackAddr == mergeTask.Address {
					logrus.Info("待归集源交易地址 :" + mergeTask.Address + "匹配到黑名单地址: " + blackAddr)
					c.db.DelCollectTask(mergeTask.Address, mergeTask.Symbol, mergeTask.Chain)
				}
			}
		}

		hotAddrs, err = c.SubBlack(hotAddrs, blackAddrs)
		if err != nil {
			logrus.Error(err)
			continue
		}

		if len(hotWallets[mergeTask.Chain]) == 0 {
			hotWallets[mergeTask.Chain] = map[string][]string{}
		}

		logrus.Info("排除黑名单后的热钱包:")
		for _, addr := range hotAddrs {
			logrus.Info(addr)
			hotWallets[mergeTask.Chain][mergeTask.Symbol] = append(hotWallets[mergeTask.Chain][mergeTask.Symbol], addr)
		}

		cnt1, _ := big.NewFloat(0).SetString(mergeTask.Balance)
		cnt2, _ := big.NewFloat(0).SetString(collectThreshold.String())

		logrus.Info("当前币种的门槛:" + collectThreshold.String())
		logrus.Info(cnt1.String(), cnt2.String())

		enough := cnt1.Cmp(cnt2)

		if enough >= 0 {
			threshold_tasks = append(threshold_tasks, mergeTask)
		} else {
			logrus.Info("排除钱包余额小于门槛详情,将要删除本条记录:")
			logrus.Info("addr: " + mergeTask.Address + " symbol:" + mergeTask.Symbol + " chain: " + mergeTask.Chain)
			// 这里直接将这条记录删除
			c.db.DelCollectTask(mergeTask.Address, mergeTask.Symbol, mergeTask.Chain)
		}
	}

	logrus.Info("得到大于门槛的待归集交易:")
	logrus.Info(threshold_tasks)

	for _, collectTask := range threshold_tasks {
		logrus.Info("Collect symbol:" + collectTask.Symbol + "chain:" + collectTask.Chain)
		tokenStr, err := c.GetTokenInfo(collectTask.Symbol, collectTask.Chain)
		if err != nil {
			logrus.Error(err)
		}

		//这里需要查询本币的资产
		str1, err := utils.GetAsset(collectTask.Chain, collectTask.Chain, collectTask.Address, c.config.Wallet.Url)
		if err != nil {
			logrus.Error(err)
			return err
		}
		balance1 := gjson.Get(str1, "balance")
		UserBalance, err := decimal.NewFromString(balance1.String())
		if err != nil {
			logrus.Error(err)
			return err
		}

		logrus.Info("得到余额:" + collectTask.Chain)
		logrus.Info(balance1.String())

		singleTxFee, _ := decimal.NewFromString("0")
		if collectTask.Chain == "hui" {
			singleTxFee, err = decimal.NewFromString(c.config.SingleFee.Fee)
			if err != nil {
				logrus.Error(err)
				return err
			}
		}
		if collectTask.Chain == "trx" {
			if collectTask.Chain == collectTask.Symbol {
				singleTxFee, err = decimal.NewFromString(c.config.TrxSingleFee.Fee)
				if err != nil {
					logrus.Error(err)
					return err
				}
			} else {
				singleTxFee, err = decimal.NewFromString(c.config.Trx20SingleFee.Fee)
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}
		logrus.Info("singleFee: " + singleTxFee.String())
		enough := UserBalance.Cmp(singleTxFee)

		if enough <= 0 { //反向打gas--fundFee 钱包模块
			logrus.Warn("fundFee:")
			//gas--getToken token模块
			fee_value := gjson.Get(tokenStr, "give_fee_value")

			fund := types.Fund{
				AppId:     "",
				OrderId:   utils.NewIDGenerator().Generate(),
				AccountId: collectTask.Uid,
				Chain:     collectTask.Chain,
				Symbol:    "hui",
				To:        collectTask.Address,
				Amount:    fee_value.String(),
			}
			msg, err := json.Marshal(fund)
			if err != nil {
				logrus.Error(err)
				continue
			}
			url := c.config.Wallet.Url + "/" + "fundFee"
			str, err := utils.Post(url, msg)
			if err != nil {
				logrus.Error(err)
				continue
			}
			logrus.Info("fundFee return " + str)

			//这里在循环查询用户的fundFee资产是否到账
			UserBalance2, err := decimal.NewFromString("0")
			logrus.Info("准备获取fundFee后的余额：")
			for {
				if UserBalance2.GreaterThan(UserBalance) {
					logrus.Info("获得新增后的余额: " + UserBalance2.String())
					collectTask.Balance = UserBalance2.String()
					break
				}
				time.Sleep(2 * time.Second)
				//这里需要查询本币的资产
				str2, err := utils.GetAsset(collectTask.Symbol, collectTask.Chain, collectTask.Address, c.config.Wallet.Url)
				if err != nil {
					logrus.Error(err)
					return err
				}
				balance2 := gjson.Get(str2, "balance")
				UserBalance2, err = decimal.NewFromString(balance2.String())
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}

		//直接归集个人地址--订单ID，插入DB中，目前仅仅是查看标志状态用
		err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
			//这里要按照一定策略选择热钱包目标地址--这里找到对应的热钱包地址然后选择
			to, err := utils.GetHotAddress(collectTask, hotWallets[collectTask.Chain][collectTask.Symbol], c.config.Wallet.Url)
			if err != nil {
				logrus.Error(err)
				return err
			}
			balance, err := decimal.NewFromString(collectTask.Balance)
			fmt.Println("balance:" + balance.String())

			shouldCollect, err := decimal.NewFromString(balance.String())

			if collectTask.Symbol == collectTask.Chain { //本币
				fmt.Println("开始计算本币应该归集的数量:")
				collectRemain := gjson.Get(tokenStr, "collect_remain")
				remain, _ := decimal.NewFromString(collectRemain.String())
				fmt.Println("remain:" + remain.String())

				shouldCollect = balance.Sub(remain)
				fmt.Println("shouldCollect:" + shouldCollect.String())
			}

			logrus.Info(shouldCollect)

			collectAmount := ""
			if collectTask.Symbol != collectTask.Chain { //remain 只对本币有效
				collectAmount = collectTask.Balance
			} else {
				collectAmount = shouldCollect.String()
			}
			logrus.Info("collectAmount" + collectAmount)
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
				Amount:    collectAmount,
			}

			msg, err := json.Marshal(fund)
			if err != nil {
				logrus.Error(err)
				return err
			}
			logrus.Info("调用归集接口")
			logrus.Info(fund)

			url := c.config.Wallet.Url + "/" + "collectToHotWallet"
			str, err := utils.Post(url, msg)
			if err != nil {
				logrus.Error(err)
				return err
			}
			logrus.Info("归集接口返回：" + str)
			if err := c.db.UpdateCollectTxState(collectTask.ID, int(types.TxCollectingState), collectTask.OrderId); err != nil {
				logrus.Errorf("update colelct transaction task error:%v tasks:[%v]", err, collectTask)
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("insert colelct sub transaction task error:%v", err)
		}

	}
	return
}

func (c CollectService) Name() string {
	return "Collect"
}
