package api

import (
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net/http"
)

type ApiService struct {
	config *config.Config
}

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
	//写合约
	//禁止账户交易-加入黑名单
	r.POST("/collectToColdWallet", a.collectToColdWallet)
	//允许账户交易-移出黑名单
	r.POST("/transferToHotWallet", a.transferToHotWallet)

	logrus.Info("HuiCollect api run at " + a.config.ServerConf.Port)

	err := r.Run(fmt.Sprintf(":%s", a.config.ServerConf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

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

	//accountId := gjson.Get(data1, "account_Id")
	//amount := gjson.Get(data1, "amount")
	//hotwalletId := gjson.Get(data1, "hotwalletId")

	res.Code = http.StatusOK
	res.Message = "success"
	//res.Data = string(d)

	c.SecureJSON(http.StatusOK, res)
}

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

	//contractAddr := gjson.Get(data1, "contractAddr")
	//operatorId := gjson.Get(data1, "operatorId")
	//targetId := gjson.Get(data1, "targetId")

	res.Code = http.StatusOK
	res.Message = "success"
	//res.Data = string(d)

	c.SecureJSON(http.StatusOK, res)
}
