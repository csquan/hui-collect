package api

import (
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Run(conf config.ServerConf) {
	r := gin.Default()

	r.POST("/tx/create", AddTask)

	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

// 接收注册过来的消息，存入db作为tx初始状态
func AddTask(c *gin.Context) {
	c.JSON(200, gin.H{
		"ok": "ok",
	})
}
