package api

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net/http"
)

type ApiService struct {
	config *config.Config
}

const StatusOk = 0

func NewApiService(cfg *config.Config) *ApiService {
	apiService := &ApiService{
		config: cfg,
	}
	return apiService
}

func (a *ApiService) Run() {
	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowOrigins = []string{"*"}
	r.Use(func(ctx *gin.Context) {
		method := ctx.Request.Method
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "*")
		// ctx.Header("Access-Control-Allow-Headers", "Content-Type,addr,GoogleAuth,AccessToken,X-CSRF-Token,Authorization,Token,token,auth,x-token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	})

	r.POST("/collectFromHotToColdWallet", a.collectToColdWallet)

	r.POST("/collectFromHotToHotWallet", a.transferToHotWallet)

	r.POST("/collectFromUserToHotWallet", a.collectToHotWallet)

	logrus.Info("HuiCollect api run at " + a.config.ServerConf.Port)

	err := r.Run(fmt.Sprintf(":%s", a.config.ServerConf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

// 将热钱包中的钱归集到冷钱包
func (a *ApiService) collectToColdWallet(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	data1 := string(buf[0:n])

	res := types.HttpRes{}

	isValid := gjson.Valid(data1)
	if isValid == false {
		logrus.Error("Not valid json")
		res.Code = http.StatusBadRequest
		res.Message = "Not valid json"
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	accountId := gjson.Get(data1, "accountId")
	chain := gjson.Get(data1, "chain")
	symbol := gjson.Get(data1, "symbol")
	to := gjson.Get(data1, "to")
	amount := gjson.Get(data1, "amount")

	url := a.config.Account.EndPoint + "/" + "query"
	fromAddr, err := utils.GetAccountId(url, accountId.String())
	if err != nil {
		logrus.Error(err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	fund := types.Fund{
		AppId:     "",
		OrderId:   utils.NewIDGenerator().Generate(),
		AccountId: accountId.String(),
		Chain:     chain.String(),
		Symbol:    symbol.String(),
		From:      fromAddr,
		To:        to.String(),
		Amount:    amount.String(),
	}

	msg, err := json.Marshal(fund)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("调用collectToColdWallet接口:将热钱包中的钱归集到冷钱包")
	logrus.Info(fund)

	url = a.config.Wallet.Url + "/" + "collectToColdWallet"
	str, err := utils.Post(url, msg)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("collectToColdWallet接口返回：" + str)

	res.Code = StatusOk
	res.Message = str

	c.SecureJSON(http.StatusOK, res)
}

// 从热钱包到热钱包
func (a *ApiService) transferToHotWallet(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	data1 := string(buf[0:n])

	res := types.HttpRes{}

	isValid := gjson.Valid(data1)
	if isValid == false {
		logrus.Error("Not valid json")
		res.Code = http.StatusBadRequest
		res.Message = "Not valid json"
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	accountId := gjson.Get(data1, "accountId")
	chain := gjson.Get(data1, "chain")
	symbol := gjson.Get(data1, "symbol")
	toId := gjson.Get(data1, "toId")
	amount := gjson.Get(data1, "amount")

	url := a.config.Account.EndPoint + "/" + "query"
	fromAddr, err := utils.GetAccountId(url, accountId.String())
	if err != nil {
		logrus.Error(err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}
	toAddr, err := utils.GetAccountId(url, toId.String())
	if err != nil {
		logrus.Error(err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	fund := types.Fund{
		AppId:     "",
		OrderId:   utils.NewIDGenerator().Generate(),
		AccountId: accountId.String(),
		Chain:     chain.String(),
		Symbol:    symbol.String(),
		From:      fromAddr,
		To:        toAddr,
		Amount:    amount.String(),
	}

	msg, err := json.Marshal(fund)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("调用transferToHotWallet接口：将热钱包中的钱转移到另一个热钱包")
	logrus.Info(fund)

	url = a.config.Wallet.Url + "/" + "transferToHotWallet"
	str, err := utils.Post(url, msg)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("transferToHotWallet返回：" + str)

	res.Code = StatusOk
	res.Message = str

	c.SecureJSON(StatusOk, res)
}

// 从用户地址转移到热钱包
func (a *ApiService) collectToHotWallet(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	data1 := string(buf[0:n])

	res := types.HttpRes{}

	isValid := gjson.Valid(data1)
	if isValid == false {
		logrus.Error("Not valid json")
		res.Code = http.StatusBadRequest
		res.Message = "Not valid json"
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	accountId := gjson.Get(data1, "accountId")
	chain := gjson.Get(data1, "chain")
	symbol := gjson.Get(data1, "symbol")
	toId := gjson.Get(data1, "toId")
	amount := gjson.Get(data1, "amount")

	url := a.config.Account.EndPoint + "/" + "query"
	fromAddr, err := utils.GetAccountId(url, accountId.String())
	if err != nil {
		logrus.Error(err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}
	toAddr, err := utils.GetAccountId(url, toId.String())
	if err != nil {
		logrus.Error(err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
		return
	}

	fund := types.Fund{
		AppId:     "",
		OrderId:   utils.NewIDGenerator().Generate(),
		AccountId: accountId.String(),
		Chain:     chain.String(),
		Symbol:    symbol.String(),
		From:      fromAddr,
		To:        toAddr,
		Amount:    amount.String(),
	}

	msg, err := json.Marshal(fund)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("调用collectToHotWallet接口:将用户账户中的钱归集到热钱包")
	logrus.Info(fund)

	url = a.config.Wallet.Url + "/" + "collectToHotWallet"
	str, err := utils.Post(url, msg)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		c.SecureJSON(http.StatusBadRequest, res)
	}
	logrus.Info("collectToHotWallet返回：" + str)

	res.Code = StatusOk
	res.Message = str

	c.SecureJSON(StatusOk, res)
}
