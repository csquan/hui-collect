package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"user/pkg/model"
	"user/pkg/util"
)

// UnAuthed error response
func UnAuthed(c *gin.Context) {
	er := HttpMsg{
		Code:    http.StatusUnauthorized,
		Message: util.ErrUnAuthed.Msg(),
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, er)
}

func UnAuth2(c *gin.Context, err util.Err) {
	er := HttpMsg{
		Code:    err.Code(),
		Message: err.Msg(),
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, er)
}

// Forbidden error response
func Forbidden(c *gin.Context, err error) {
	er := HttpMsg{
		Code:    http.StatusForbidden,
		Message: err.Error(),
	}
	c.AbortWithStatusJSON(http.StatusForbidden, er)
}

func Forbid2(c *gin.Context, err util.Err) {
	er := HttpMsg{
		Code:    err.Code(),
		Message: err.Msg(),
	}
	c.AbortWithStatusJSON(http.StatusForbidden, er)
}

// BadResp error response
func BadResp(c *gin.Context, err error) {
	er := HttpMsg{
		Code:    http.StatusBadGateway,
		Message: err.Error(),
	}
	c.AbortWithStatusJSON(http.StatusOK, er)
}

func BadRes(c *gin.Context, err util.Err) {
	er := HttpMsg{
		Code:    err.Code(),
		Message: err.Msg(),
	}
	c.AbortWithStatusJSON(http.StatusOK, er)
}

// HttpMsg error msg
type HttpMsg struct {
	Code    int    `json:"code" example:"502"`
	Message string `json:"message" example:"可耻滴失败鸟"`
}

// HttpData success data
type HttpData struct {
	Code int         `json:"code" example:"0"`
	Data interface{} `json:"data"`
}

func Header(c *gin.Context, k, v string) {
	c.Header("Access-Control-Expose-Headers", k)
	c.Header(k, v)
}

// GoodResp success response
func GoodResp(c *gin.Context, data interface{}) {
	ret := HttpData{
		Code: 0,
		Data: data,
	}
	c.JSON(http.StatusOK, ret)
}

func GetTokenUser(c *gin.Context) *model.TokenUser {
	if value, exists := c.Get("currentUser"); !exists {
		return nil
	} else {
		if tu, ok := value.(model.TokenUser); ok {
			return &tu
		} else {
			return nil
		}
	}
}
