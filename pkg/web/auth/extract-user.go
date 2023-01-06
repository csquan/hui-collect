package auth

import (
	"github.com/gin-gonic/gin"
	"strings"
	"user/pkg/conf"
	"user/pkg/log"
	"user/pkg/util"
	"user/pkg/web"
)

func MustExtractUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var accessToken string
		cookie, err := c.Cookie("access_token")

		authorizationHeader := c.Request.Header.Get("Authorization")
		fields := strings.Fields(authorizationHeader)

		if len(fields) != 0 && fields[0] == "Bearer" {
			accessToken = fields[1]
		} else if err == nil {
			accessToken = cookie
		}

		if accessToken == "" {
			web.BadRes(c, util.ErrTokenInvalid)
			log.Log.Error().Msg("no accessToken")
			return
		}

		sub, er := ValidateToken(accessToken, conf.Conf.AccessTokenPublicKey)
		if er != nil {
			web.BadRes(c, er)
			log.Log.Error().Msg(er.LStr())
			return
		}

		c.Set("currentUser", sub)
		c.Next()
	}
}

func SilentExtractUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var accessToken string
		cookie, err := c.Cookie("access_token")

		authorizationHeader := c.Request.Header.Get("Authorization")
		fields := strings.Fields(authorizationHeader)

		if len(fields) != 0 && fields[0] == "Bearer" {
			accessToken = fields[1]
		} else if err == nil {
			accessToken = cookie
		}

		if accessToken == "" {
			log.Log.Debug().Msg("no accessToken")
			return
		}

		sub, er := ValidateToken(accessToken, conf.Conf.AccessTokenPublicKey)
		if er != nil {
			log.Log.Warn().Msg(er.LStr())
			return
		}

		c.Set("currentUser", sub)
		c.Next()
	}
}
