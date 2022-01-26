package api

import (
	"bytes"
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
	isRun, err := h.db.GetTaskSwitch()
	if err != nil {
		logrus.Errorf("get task switch err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	if !isRun {
		c.JSON(http.StatusConflict, "task switch is closed")
		return
	}
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

	if task == nil {
		c.JSON(http.StatusNotFound, "task not found")
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

func (h *FullRebalanceHandler) Open(c *gin.Context) {
	logrus.Info("open task switch")
	err := h.db.UpdateTaskSwitch(true)
	if err != nil {
		logrus.Errorf("open task switch err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	c.JSON(http.StatusOK, "success")
}

func (h *FullRebalanceHandler) Close(c *gin.Context) {
	logrus.Info("close task switch")
	err := h.db.UpdateTaskSwitch(false)
	if err != nil {
		logrus.Errorf("close task switch err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	c.JSON(http.StatusOK, "success")
}

func (h *FullRebalanceHandler) GetTaskSwitch(c *gin.Context) {
	isRun, err := h.db.GetTaskSwitch()
	if err != nil {
		logrus.Errorf("get task switch err:%v", err)
		c.JSON(http.StatusInternalServerError, "server err")
		return
	}
	c.JSON(http.StatusOK, isRun)
}

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func AccessLogHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		logrus.Infof("url:%s, status:%d, res:%s", c.Request.URL, c.Writer.Status(), blw.body.String())
	}
}

func Run(conf config.ServerConf, db types.IDB) {
	h := &FullRebalanceHandler{
		db: db,
	}
	r := gin.Default()
	r.Use(AccessLogHandler())
	authorized := r.Group("/", gin.BasicAuth(conf.Users))

	authorized.POST("fullRebalance/create", h.AddTask)
	authorized.GET("fullRebalance/get", h.GetTask)
	authorized.GET("taskSwitch/get", h.GetTaskSwitch)
	authorized.POST("taskSwitch/open", h.Open)
	authorized.POST("taskSwitch/close", h.Close)

	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}
