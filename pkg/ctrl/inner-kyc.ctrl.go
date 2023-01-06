package ctrl

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"user/pkg/model"
	"user/pkg/util"
	"user/pkg/web"
)

// QueryUserByAddr godoc
//
//	@Summary		内部查询，归集模块查询企业用户信息
//	@Description	成功返回用户信息
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.AddrInput	true	"用户链上地址"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/i-q-user-by-addr [post]
func (pc *PubCtrl) QueryUserByAddr(c *gin.Context) {
	var ai model.AddrInput
	if err := c.ShouldBindJSON(&ai); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := ai.Validate(pc.Pk, pc.Log); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	if user, ok := dbUserByAddr(pc.Ctrl, c, ai.Addr); !ok {
		return
	} else {
		web.GoodResp(c, user.ToFr())
	}
}

// KYC godoc
//
//	@Summary		企业进行后台认证，验证成功后返回认证状态给用户模块
//	@Description	成功返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.FirmConfirmed	true	"用户ID和认证成功的企业信息+校验字段"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/kyc-ok [post]
func (pc *PubCtrl) KYC(c *gin.Context) {
	var fc model.FirmConfirmed
	if err := c.ShouldBindJSON(&fc); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := fc.Validate(pc.Pk, pc.Log); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, fc.BID)
	if !ok {
		return
	}
	user.FirmName = fc.FirmName
	user.FirmType = fc.FirmType
	user.Country = fc.Country
	user.FirmVerified = time.Now().Unix()
	if err := pc.DB.Model(&user).Where("b_id=?", fc.BID).
		Updates(map[string]interface{}{
			"firm_name":     user.FirmName,
			"firm_type":     user.FirmType,
			"country":       user.Country,
			"firm_verified": user.FirmVerified,
		}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}
	web.GoodResp(c, "Ok")
}

// KycQuery godoc
//
//	@Summary		后台管理员查询企业用户信息
//	@Description	成功返回用户信息
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.FirmQuery	true	"用户ID/邮箱/手机+校验字段"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/kyc-query [post]
func (pc *PubCtrl) KycQuery(c *gin.Context) {
	var fq model.FirmQuery
	if err := c.ShouldBindJSON(&fq); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := fq.Validate(pc.Pk, pc.Log); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	ok := false
	var user model.User
	if fq.Uid != "" {
		if user, ok = dbUserByID(pc.Ctrl, c, fq.Uid); !ok {
			return
		}
	}
	if fq.Email != "" {
		if user, ok = dbUserByEmail(pc.Ctrl, c, fq.Email); !ok {
			return
		}
	}
	if fq.Mobile != "" {
		if user, ok = dbUserByMobile(pc.Ctrl, c, fq.Number()); !ok {
			return
		}
	}
	web.GoodResp(c, user.ToFr())
}

// KycSetUserStatus godoc
//
//	@Summary		后台管理员设置企业用户状态
//	@Description	成功返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.KycUserStatusInput	true	"用户ID和设置的状态+校验字段"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/kyc-set-user-status [post]
func (pc *PubCtrl) KycSetUserStatus(c *gin.Context) {
	var usi model.KycUserStatusInput
	if err := c.ShouldBindJSON(&usi); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := usi.Validate(pc.Pk, pc.Log); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, usi.Uid)
	if !ok {
		return
	}
	if err := pc.DB.Model(&user).Update("status=?", usi.Status).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}
	if usi.Status < 0 {
		_, err := pc.RDB.Delete(context.Background(), "singleton:"+user.BID)
		if err != nil {
			web.BadRes(c, util.ErrDB)
			pc.Log.Error().Str("when deleting singleton:", user.BID).Msg(err.LStr())
			return
		}
	}
	web.GoodResp(c, user.ToFr())
}

// KycUserList godoc
//
//	@Summary		后台管理员查询企业用户列表
//	@Description	成功返回用户列表
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.KycUserListInput	true	"列表参数+校验字段"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/kyc-user-list [post]
func (pc *PubCtrl) KycUserList(c *gin.Context) {
	var uli model.KycUserListInput
	if err := c.ShouldBindJSON(&uli); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := uli.Validate(pc.Pk, pc.Log); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	var user model.User
	limit := uli.GetLimit()
	var rowCount int64
	if err := pc.DB.Model(&user).
		Where("created_at between ? and ?", time.Unix(uli.Start, 0), time.Unix(uli.End, 0)).
		Count(&rowCount).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}

	var users []model.User
	if err := pc.DB.Model(&user).
		Where("created_at between ? and ?", time.Unix(uli.Start, 0), time.Unix(uli.End, 0)).
		Offset(uli.GetFrom()).Limit(limit).Order("id asc").
		Find(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			web.BadRes(c, util.ErrNoRec)
		} else {
			web.BadRes(c, util.ErrDB)
		}
		pc.Log.Err(err).Send()
		return
	}

	rows := make([]model.FirmResp, len(users))
	for i, u := range users {
		rows[i] = *u.ToFr()
	}

	web.GoodResp(c, model.UserPage{
		Total: rowCount,
		Page:  uli.Page,
		Limit: limit,
		Rows:  rows,
	})
}
