package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/types"
)

type FullRebalanceTask struct {
	ID      uint64 `json:"id"`
	State   int    `json:"state"`
	Message string `json:"message"`
	Created string `json:"created"`
	Updated string `json:"updated"`
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
		c.JSON(http.StatusInternalServerError, "server err")
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
		logrus.Errorf("save task err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"task_id": task.ID,
	})
}

func (h *FullRebalanceHandler) GetTask(c *gin.Context) {
	taskIdStr := c.Request.FormValue("task_id")
	if taskIdStr == "" {
		c.JSON(http.StatusBadRequest, "task_id not set")
		return
	}
	taskId, err := strconv.ParseUint(taskIdStr, 10, 64)
	if err != nil {
		logrus.Errorf("task_id not int %s", taskIdStr)
		c.JSON(http.StatusBadRequest, "task_id type err")
		return
	}
	task, err := h.db.GetFullRelalanceTask(taskId)
	if err != nil {
		logrus.Errorf("get task err:%v,task_id:%d", err, taskId)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	taskView := &FullRebalanceTask{
		ID:      task.ID,
		Message: task.Message,
		State:   task.State,
		Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
		Updated: task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	c.JSON(http.StatusOK, taskView)
}

func Run(conf config.APIConf, db types.IDB) {
	h := &FullRebalanceHandler{
		db: db,
	}
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts(conf.Users)))
	authorized.POST("fullRebalance/create", h.AddTask)
	authorized.GET("fullRebalance/get", h.GetTask)
	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}
