package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

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

func (c *CollectService) GetBalances(chain string, addr string, contractAddr string) (string, error) {
	param := types.BalanceParam{
		Chain:    chain,
		Address:  addr,
		Contract: contractAddr,
	}
	logrus.Info(param)
	msg, err := json.Marshal(param)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	str, err := utils.Post(c.config.ChainNode.Url, msg)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	logrus.Info("balance return " + str)
	return str, nil
}

func (c *CollectService) GetMappedTokenInfo() ([]*string, error) {
	assetStrs := make([]*string, 0)
	url := c.config.Wallet.Url + "/" + "getSupportedMappedToken"
	res, err := utils.Get(url)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if res == "[]" {
		return nil, nil
	}
	logrus.Info(res)
	coinArray := strings.Split(res[1:len(res)-1], ",")
	logrus.Info(coinArray)
	url = c.config.Wallet.Url + "/" + "getMappedToken"
	for _, coin := range coinArray {
		logrus.Info("当前币种")
		logrus.Info(coin)
		param := types.Coin{
			MappedSymbol: coin[1 : len(coin)-1],
		}
		msg, err1 := json.Marshal(param)
		if err1 != nil {
			logrus.Error(err1)
			return nil, err1
		}
		res, err1 := utils.Post(url, msg)
		if err1 != nil {
			logrus.Error(err1)
			return nil, err1
		}
		logrus.Info("getMappedToken返回：")
		logrus.Info(res)
		assetStrs = append(assetStrs, &res)
	}
	return assetStrs, nil
}

type Stringer interface {
	String() string
}

func ToString(any interface{}) string {
	if v, ok := any.(Stringer); ok {
		return v.String()
	}
	switch v := any.(type) {
	case int:
		return strconv.Itoa(v)
	}
	return "???"
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

	//获取所有支持的mappedToken名称
	tokensArr, err := c.GetMappedTokenInfo()
	if err != nil {
		logrus.Error(err)
	}
	tokens := map[string]types.TokenSymbol{}
	for _, token := range tokensArr {
		var infos []map[string]interface{}
		err = json.Unmarshal([]byte(*token), &infos)
		if err != nil {
			logrus.Error(err)
			continue
		}
		ContractAddress := common.HexToAddress(infos[0]["contract_address"].(string)).String()

		obj := types.TokenSymbol{
			Symbol:       infos[0]["symbol"].(string),
			MappedSymbol: infos[0]["mapped_symbol"].(string),
		}
		tokens[ContractAddress] = obj
	}
	logrus.Info(tokens)

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

		//这里根据合约地址找
		mappedSymbol := tokens[mergeTask.ContractAddress].MappedSymbol
		contractAddress := strings.TrimSpace(mergeTask.ContractAddress)
		if contractAddress == "" {
			mappedSymbol = mergeTask.Symbol
		}
		str, err := c.GetTokenInfo(mappedSymbol, mergeTask.Chain)

		if err != nil {
			logrus.Error(err)
			continue
		}
		collectThreshold := gjson.Get(str, "collect_threshold")
		hotWallet := gjson.Get(str, "hot_wallets")
		blacklist := gjson.Get(str, "blacklist")

		if len(hotWallet.String()) == 0 {
			logrus.Warn("*****热钱包为空，请检查配置******")
			continue
		}

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
		logrus.Info("-----Collect symbol:" + collectTask.Symbol + "chain:" + collectTask.Chain)
		mappedSymbol := tokens[collectTask.ContractAddress].MappedSymbol

		tokenStr, err := c.GetTokenInfo(mappedSymbol, collectTask.Chain)
		if err != nil {
			logrus.Error(err)
		}

		//这里获取本币余额
		str, err := c.GetBalances(collectTask.Chain, collectTask.Address, "")
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Info(str)
		code := gjson.Get(str, "code")
		if code.Int() != 0 {
			msg := gjson.Get(str, "msg")
			logrus.Error(msg.String())
			return fmt.Errorf(msg.String())
		}
		balance := gjson.Get(str, "balance")
		UserBalance, err := decimal.NewFromString(balance.String())
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Info("得到余额: " + UserBalance.String())

		singleTxFee, _ := decimal.NewFromString("0")
		if collectTask.Chain == "hui" || collectTask.Chain == "eth" {
			singleTxFee, err = decimal.NewFromString(c.config.SingleFee.Fee)
			if err != nil {
				logrus.Error(err)
				return err
			}
		}
		logrus.Info("开始取配置文件中的singleFee")
		if collectTask.Chain == "trx" {
			logrus.Info(collectTask.Chain + ":" + collectTask.Symbol)
			if collectTask.Chain == collectTask.Symbol {
				logrus.Info("collectTask.Chain == collectTask.Symbol")
				singleTxFee, err = decimal.NewFromString(c.config.TrxSingleFee.Fee)
				if err != nil {
					logrus.Error(err)
					return err
				}
			} else {
				logrus.Info("collectTask.Chain != collectTask.Symbol")
				singleTxFee, err = decimal.NewFromString(c.config.Trx20SingleFee.Fee)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
		logrus.Info("singleFee: " + singleTxFee.String())
		enough := UserBalance.Cmp(singleTxFee)

		if enough <= 0 {
			if collectTask.FundFeeOrderId == "" {
				logrus.Warn("fundFee:")
				//gas--getToken token模块
				fee_value := gjson.Get(tokenStr, "give_fee_value")
				logrus.Warn("fee_value:")
				logrus.Info(fee_value)
				fund := types.Fund{
					AppId:     "",
					OrderId:   utils.NewIDGenerator().Generate(),
					AccountId: collectTask.Uid,
					Chain:     collectTask.Chain,
					Symbol:    collectTask.Chain,
					To:        collectTask.Address,
					Amount:    fee_value.String(),
				}
				logrus.Info(fund)
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

				//这里保存FundFeeOrderId
				collectTask.FundFeeOrderId = fund.OrderId
				collectTask.BalanceBeforeFund = UserBalance.String()
				logrus.Info("更新fundFeeID为" + fund.OrderId + " Fund前的余额为：" + collectTask.BalanceBeforeFund)
				logrus.Info(collectTask.ID)
				c.db.UpdateCollectTxFundFeeInfo(collectTask.FundFeeOrderId, collectTask.BalanceBeforeFund, collectTask.ID)
			} else {
				logrus.Info("检测到上次已经有FundFee orderID" + collectTask.FundFeeOrderId)
			}

			CompareBalance := UserBalance
			if collectTask.BalanceBeforeFund != "" {
				BalanceBeforeFund, err := decimal.NewFromString(collectTask.BalanceBeforeFund)
				if err != nil {
					logrus.Error(err)
					continue
				}
				CompareBalance = BalanceBeforeFund
				logrus.Info("FundFee比较基准为" + collectTask.BalanceBeforeFund)
			}

			//这里在循环查询用户的fundFee资产是否到账
			UserBalance2, _ := decimal.NewFromString("0")
			logrus.Info("准备获取fundFee后的余额：")
			count := 0

			for {
				if UserBalance2.GreaterThan(CompareBalance) {
					logrus.Info("获得新增后的余额: " + UserBalance2.String())
					collectTask.Balance = UserBalance2.String()
					logrus.Info("已经赋值为最新的余额：" + collectTask.Balance)
					break
				}
				time.Sleep(2 * time.Second)
				if count >= 5 {
					logrus.Error("获得新增后的余额错误，超过5次，continue")
					return err
				}
				count = count + 1
				//这里需要查询本币的资产
				str2, err := c.GetBalances(collectTask.Chain, collectTask.Address, "")
				if err != nil {
					logrus.Error(err)
					continue
				}

				balance2 := gjson.Get(str2, "balance")
				UserBalance2, err = decimal.NewFromString(balance2.String())
				if err != nil {
					logrus.Error(err)
					continue
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
			//这里需要查询最新的资产-重新获取最新的
			str3, err := c.GetBalances(collectTask.Chain, collectTask.Address, collectTask.ContractAddress)
			if err != nil {
				logrus.Error(err)
				return err
			}
			balance2 := gjson.Get(str3, "balance")
			UserBalance3, err := decimal.NewFromString(balance2.String())
			if err != nil {
				logrus.Error(err)
				return err
			}

			balance, err := decimal.NewFromString(UserBalance3.String())
			fmt.Println("获取到最新balance:" + balance.String())

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
			//这里根据合约地址找
			mappedSymbol := tokens[collectTask.ContractAddress].MappedSymbol

			logrus.Info("collectAmount" + collectAmount)
			collectTask.OrderId = utils.NewIDGenerator().Generate()

			collectDecimal := decimal.New(int64(10), int32(collectTask.Decimal-1))
			logrus.Info("精度：" + collectDecimal.String())
			amount, _ := decimal.NewFromString(collectAmount)
			logrus.Info("转换前的金额：" + amount.String())

			if collectTask.Symbol == collectTask.Chain { //本币不带精度
				logrus.Info("本币归集，不需要转换")
			} else { //代币带有精度 就是很多0
				logrus.Info("代币归集，需要转换")
				amount = amount.Div(collectDecimal)
				logrus.Info("转换后的金额：" + amount.String())
			}

			//这里调用keep的归集交易接口  --collenttohotwallet
			fund := types.Fund{
				AppId:     "",
				OrderId:   collectTask.OrderId,
				AccountId: collectTask.Uid,
				Chain:     collectTask.Chain,
				Symbol:    mappedSymbol,
				From:      collectTask.Address,
				To:        to, //这里要按照一定策略选择热钱包
				Amount:    amount.String(),
			}

			msg, err := json.Marshal(fund)
			if err != nil {
				logrus.Error(err)
				return err
			}
			logrus.Info("****调用归集接口****")
			logrus.Info(fund)

			url := c.config.Wallet.Url + "/" + "collectToHotWallet"
			str, err := utils.Post(url, msg)
			if err != nil {
				logrus.Error(err)
				return err
			}
			logrus.Info("归集接口返回：" + str)
			errMsg := gjson.Get(str, "error")

			if errMsg.String() != "" {
				logrus.Errorf(errMsg.String())
				return err
			}
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
