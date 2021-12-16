package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
)

type Ret struct {
	Message string
	Data    interface{}
}

type FullRebalanceReq struct {
	Message string `json:"message"`
	Params  string `json:"params"`
}
type FullRebalanceHandler struct {
	db types.IDB
}

func (h *FullRebalanceHandler) AddTask(c *gin.Context) {
	var (
		req FullRebalanceReq
	)
	err := c.BindJSON(&req)
	if err != nil && err != io.EOF {
		logrus.Errorf("read req err:%v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	tasks, err := h.db.GetOpenedFullReBalanceTasks()
	if err != nil {
		logrus.Errorf("get opened full task err:%v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(tasks) != 0 {
		c.JSON(http.StatusForbidden, "full rebalance exist")
		return
	}

	partTasks, err := h.db.GetOpenedPartReBalanceTasks()
	if err != nil {
		logrus.Errorf("get opened part task err:%v", err)
		c.JSON(http.StatusInternalServerError, "get partRebalanceTasks err")
		return
	}
	if len(partTasks) != 0 {
		c.JSON(http.StatusForbidden, "part rebalance exist")
		return
	}
	task := &types.FullReBalanceTask{
		BaseTask: &types.BaseTask{
			Message: req.Message,
		},
		Params: req.Params,
	}
	err = h.db.SaveFullRebalanceTask(h.db.GetEngine(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "suc")
}

func Run(port int, db types.IDB) {
	h := &FullRebalanceHandler{
		db: db,
	}
	r := gin.Default()
	r.POST("/fullRebalance", h.AddTask)
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}
