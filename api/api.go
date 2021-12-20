package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type Ret struct {
	Message string
	Data    interface{}
}

type FullRebalanceHandler struct {
	db types.IDB
}

func (h *FullRebalanceHandler) AddTask(c *gin.Context) {
	tasks, err := h.db.GetOpenedFullReBalanceTasks()
	if err != nil {
		logrus.Errorf("get opened full task err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	if len(tasks) != 0 {
		c.JSON(http.StatusConflict, "full rebalance exist")
		return
	}

	partTasks, err := h.db.GetOpenedPartReBalanceTasks()
	if err != nil {
		logrus.Errorf("get opened part task err:%v", err)
		c.JSON(http.StatusInternalServerError, "get partRebalanceTasks err")
		return
	}
	if len(partTasks) != 0 {
		c.JSON(http.StatusConflict, "part rebalance exist")
		return
	}
	task := &types.FullReBalanceTask{
		BaseTask: &types.BaseTask{},
	}
	err = h.db.SaveFullRebalanceTask(h.db.GetEngine(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"msg": "suc",
		"data": struct {
			TaskID uint64 `json:"task_id"`
		}{
			TaskID: task.ID,
		},
	})
}

func Run(conf config.APIConf, db types.IDB) {
	h := &FullRebalanceHandler{
		db: db,
	}
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"user0": "123",
	}))
	authorized.POST("fullRebalance/create", h.AddTask)
	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}
