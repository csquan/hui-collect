package api

import (
	"bytes"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
)

type ApiService struct {
	db     types.IDB
	config *config.Config
}

func NewApiService(db types.IDB, c *config.Config) *ApiService {
	return &ApiService{
		db:     db,
		config: c,
	}
}

func (s *ApiService) Run(conf config.ServerConf) {
	r := gin.Default()

	r.POST("/tx/create", s.AddTask)

	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

// 接收注册过来的消息，存入db作为tx初始状态
func (s *ApiService) AddTask(c *gin.Context) {
	from := c.PostForm("from")
	to := c.PostForm("to")
	data := c.PostForm("data")
	userID := c.PostForm("userid")

	//组装task
	task := types.TransactionTask{
		UserID:    userID,
		From:      from,
		To:        to,
		InputData: data,
	}
	task.State = int(types.TxInitState)

	err := utils.CommitWithSession(s.db, func(session *xorm.Session) (execErr error) {
		//create next state task
		if err := s.db.SaveTxTask(session, &task); err != nil {
			logrus.Errorf("save transaction task error:%v tasks:[%v]", err, task)
			return
		}
		s.tgalert(&task)
		return
	})
	if err != nil {
		logrus.Fatalf("SaveTxTask CommitWithSession err:%v", err)
	}

	c.JSON(200, gin.H{
		"ok": "ok",
	})
}

func (c *ApiService) tgalert(task *types.TransactionTask) {
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
	buffer.WriteString(fmt.Sprintf("告警:交易初始\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}
