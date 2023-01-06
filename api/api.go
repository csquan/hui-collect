package api

import (
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net/http"
)

const ADDRLEN = 42

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

	r.POST("/monitor/add", s.Add)
	r.POST("/monitor/remove", s.Remove)

	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

func checkAddr(addr string) error {
	if addr[:2] != "0x" {
		return errors.New("addr must start with 0x")
	}
	if len(addr) != ADDRLEN {
		return errors.New("addr len wrong ,must 40")
	}
	return nil
}

func (s *ApiService) Add(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	data := string(buf[0:n])

	isValid := gjson.Valid(data)
	if isValid == false {
		fmt.Println("Not valid json")
	}

	accountAddr := gjson.Get(data, "account_address")

	err := checkAddr(accountAddr.String())

	res := types.HttpRes{}

	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	monitor := types.Monitor{
		Addr:   accountAddr.String(),
		Height: 0,
	}

	err = utils.CommitWithSession(s.db, func(session *xorm.Session) (execErr error) {
		if err := s.db.SaveMonitorTask(session, &monitor); err != nil {
			logrus.Errorf("save monitor task error:%v tasks:[%v]", err, task)
			return
		}
		return
	})
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		c.SecureJSON(http.StatusInternalServerError, res)
	}

	res.Code = 0
	res.Message = "OK"
	c.SecureJSON(http.StatusOK, res)
}

func (s *ApiService) Remove(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	data := string(buf[0:n])

	isValid := gjson.Valid(data)
	if isValid == false {
		fmt.Println("Not valid json")
	}

	accountAddr := gjson.Get(data, "account_address")

	res := types.HttpRes{}

	err := s.db.RemoveMonitorTask(accountAddr.String())
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	res.Code = 0
	res.Message = "OK"
	c.SecureJSON(http.StatusOK, res)
}
