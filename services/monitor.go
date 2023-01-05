package services

import (
	"bytes"
	"fmt"
	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"time"
)

type MonitorService struct {
	db       types.IDB
	block_db types.IDB
	config   *config.Config
}

func NewMonitorService(db types.IDB, block_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		db:       db,
		block_db: block_db,
		config:   c,
	}
}

func (c *MonitorService) tgAlert(task *types.TransactionTask) {
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

func createMonitorMsg(task *types.TransactionTask) (string, error) {
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

func (c *MonitorService) InsertCollectTx(parentID uint64, from string, to string, userID string, requestID string, chainId string, inputdata string, value string, tx_type int) error {
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

func (c *MonitorService) Run() (err error) {
	erc20_txs, err := c.block_db.GetMonitorCollectTask("0x206beddf4f9fc55a116890bb74c6b79999b14eb1")
	if err != nil {
		return
	}
	if len(erc20_txs) == 0 {
		logrus.Infof("no tx of target addr.")
		return
	}

	if len(erc20_txs) != 1 { // 先测试一次取出一条
		logrus.Infof("test:only one tx one time!")
		return
	}

	for _, erc20_tx := range erc20_txs {
		collectTask := types.CollectTxDB{}
		collectTask.Copy(erc20_tx)

		collectTask.CollectState = int(types.TxReadyCollectState)

		err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
			if err := c.db.InsertCollectTx(s, &collectTask); err != nil {
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

func (c MonitorService) Name() string {
	return "Monitor"
}
