package ctrl

import (
	"context"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"time"
	"user/pkg/conf"
	"user/pkg/db"
	"user/pkg/log"
	"user/pkg/model"
	"user/pkg/redis"
	"user/pkg/util"
	"user/pkg/util/ecies"
	"user/pkg/web"
)

type PubCtrl struct {
	*Ctrl
	Pk *ecies.PrivateKey
}

func NewPubCtrl() PubCtrl {
	rdb := redis.NewStore("go:user:controller:")
	logger := log.Log.With().Str("ctrl", "pub").Logger()
	pk, er := ecies.PrivateFromString(conf.Conf.KycPrivateKey)
	if er != nil {
		logger.Error().Msg(er.LStr())
		panic(er)
	}
	return PubCtrl{Ctrl: &Ctrl{DB: db.DB, RDB: rdb, Log: &logger}, Pk: pk}
}

// MyEmail godoc
//
//	@Summary		首次提交电子邮箱，要求发起认证真实性
//	@Description	后台生成 6 位随机数字邮件发送并保存在 Redis，5分钟过期
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			email	body		model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Success		200		{object}	web.HttpData
//	@Failure		502		{object}	web.HttpMsg
//	@Router			/pub/myEmail [post]
func (pc *PubCtrl) MyEmail(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Warn().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	// 首先保证数据库中不重复
	if !ensureEmailNotReg(pc.Ctrl, c, email) {
		return
	}
	sendEmailNStore(pc.Ctrl, c, email)
}

// MyMobile godoc
//
//	@Summary		首次提交手机号，要求发起认证真实性
//	@Description	后台生成 6 位随机数字短信发送给用户并保存在 Redis，5分钟过期
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			mobile	body		model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Success		200		{object}	web.HttpData
//	@Failure		502		{object}	web.HttpMsg
//	@Router			/pub/myMobile [post]
func (pc *PubCtrl) MyMobile(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Warn().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Number()

	// 首先保证不重复
	if !ensureMobileNotReg(pc.Ctrl, c, mobile) {
		return
	}

	sendMobileNStore(pc.Ctrl, c, mobile)
}

// VerifyEmail godoc
//
//	@Summary		校验邮件验证码，成功证明是本人邮箱
//	@Description	成功返回一个 token 用于保存当前用户的所有信息
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			input	body		model.VerifyEmailInput	true	"输入你的邮箱和收到的验证码"
//	@Success		200		{object}	web.HttpData
//	@Failure		502		{object}	web.HttpMsg
//	@Router			/pub/verifyEmail [post]
func (pc *PubCtrl) VerifyEmail(c *gin.Context) {
	var verifyEmailInput model.VerifyEmailInput
	if err := c.ShouldBindJSON(&verifyEmailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := verifyEmailInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	email := verifyEmailInput.Email
	code := verifyEmailInput.Code

	if !emailCodeExists(pc.Ctrl, c, email, code) {
		return
	}

	tu := model.TokenUser{
		Email:     email,
		EAuthTime: time.Now().Unix(),
		Auth2:     false,
	}
	tu.ReSk()
	AToken(pc.Ctrl, c, &tu)
}

// VerifyMobile godoc
//
//	@Summary		校验手机验证码，成功证明是本人手机
//	@Description	成功返回一个 token 可以用于保存当前用户的所有信息
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			input	body		model.VerifyMobileInput	true	"输入你的手机号和收到的验证码"
//	@Success		200		{object}	web.HttpData
//	@Failure		502		{object}	web.HttpMsg
//	@Router			/pub/verifyMobile [post]
func (pc *PubCtrl) VerifyMobile(c *gin.Context) {
	var verifyMobileInput model.VerifyMobileInput
	if err := c.ShouldBindJSON(&verifyMobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := verifyMobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Warn().Msg(err.LStr())
		return
	}
	mobile := verifyMobileInput.Number()
	code := verifyMobileInput.Code

	if ok := mobileCodeExists(pc.Ctrl, c, mobile, code); !ok {
		return
	}

	tu := model.TokenUser{
		Mobile:    mobile,
		MAuthTime: time.Now().Unix(),
		Auth2:     false,
	}
	tu.ReSk()
	AToken(pc.Ctrl, c, &tu)
}

// SignIn godoc
//
//	@Summary		登录系统
//	@Description	成功返回一个 token 可以用于保存当前用户的所有信息
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			user	body		model.SignInInput	true	"user information"
//	@Success		200		{object}	web.HttpData
//	@Failure		502		{object}	web.HttpMsg
//	@Router			/pub/signIn [post]
func (pc *PubCtrl) SignIn(c *gin.Context) {
	var signInInput model.SignInInput
	if err := c.ShouldBindJSON(&signInInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if data, err := base64.StdEncoding.DecodeString(signInInput.Password); err != nil {
		web.BadRes(c, util.ErrMsgDecode)
		pc.Log.Err(err).Send()
		return
	} else {
		signInInput.Password = string(data)
	}

	var user model.User
	ok := false
	user, ok = dbUser(pc.Ctrl, c, signInInput.User)
	if !ok {
		return
	}

	if user.Status < 0 {
		web.BadRes(c, util.ErrUserBan)
		pc.Log.Error().Int8("user.status", user.Status).Send()
		return
	}

	if err := util.VerifyPassword(user.Password, signInInput.Password); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	tu := user.ToTU()
	tu.Auth2 = false
	if !rCookie(pc.Ctrl, c, tu) {
		return
	}
	aToken(pc.Ctrl, c, tu)
}

// Refresh godoc
//
//	@Summary		刷新登录
//	@Description	因为token有效期并不是很长，失效后在可刷新周期内，刷新一下可以获得一个新token
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/refresh [get]
func (pc *PubCtrl) Refresh(c *gin.Context) {
	tu, ok := shouldGetRTu(pc.Ctrl, c)
	if !ok {
		return
	}
	if !tuDbOk(pc.Ctrl, c, tu) {
		return
	}
	if !rCookie(pc.Ctrl, c, tu) {
		return
	}
	AToken(pc.Ctrl, c, tu)
}

// ForgetE godoc
//
//	@Summary		忘记密码，从邮箱开始找回。
//	@Description	需要已经认证过的邮箱。返回可供认证的信息
//	@Tags			pub
//	@Accept			json
//	@Param			email	body	model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-e [post]
func (pc *PubCtrl) ForgetE(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	user, ok := dbUserByEmail(pc.Ctrl, c, email)
	if !ok {
		return
	}

	res := model.ForgetResponse{
		MaskedEmail:  user.MaskedEmail(),
		MaskedMobile: user.MaskedMobile(),
		E:            user.LastEVTime > 0,
		M:            user.LastMVTime > 0,
		G:            user.LastGVTime > 0,
	}
	web.GoodResp(c, res)
}

// ForgetM godoc
//
//	@Summary		忘记密码，从手机开始找回
//	@Description	需要手机已经认证过。返回可供验证的信息
//	@Tags			pub
//	@Accept			json
//	@Param			mobile	body	model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-m [post]
func (pc *PubCtrl) ForgetM(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Number()

	user, ok := dbUserByMobile(pc.Ctrl, c, mobile)
	if !ok {
		return
	}
	res := model.ForgetResponse{
		MaskedEmail:  user.MaskedEmail(),
		MaskedMobile: user.MaskedMobile(),
		E:            user.LastEVTime > 0,
		M:            user.LastMVTime > 0,
		G:            user.LastGVTime > 0,
	}
	web.GoodResp(c, res)
}

// ForgetSendEmailByEmail godoc
//
//	@Summary		知道邮箱，找回密码时，请求邮箱验证码
//	@Description	需要邮箱已经认证过。返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			email	body	model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-send-email-by-email [post]
func (pc *PubCtrl) ForgetSendEmailByEmail(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	if !ensureEmailExists(pc.Ctrl, c, email) {
		return
	}
	sendEmailNStore(pc.Ctrl, c, email)
}

// ForgetSendSMSByEmail godoc
//
//	@Summary		知道邮箱，找回密码时，请求手机验证码
//	@Description	需要邮箱和手机都已经认证过。返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			email	body	model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-send-sms-by-email [post]
func (pc *PubCtrl) ForgetSendSMSByEmail(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	// 查询用户是否存在
	var user model.User
	ok := false
	if user, ok = dbUserByEmail(pc.Ctrl, c, email); !ok {
		return
	}
	if !ensureUserHasMobile(pc.Ctrl, c, &user) {
		return
	}

	sendMobileNStore(pc.Ctrl, c, user.Mobile)
}

// ForgetSendEmailByMobile godoc
//
//	@Summary		知道手机，找回密码时，请求邮箱验证码
//	@Description	需要手机和邮箱都已认证过。返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			mobile	body	model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-send-email-by-mobile [post]
func (pc *PubCtrl) ForgetSendEmailByMobile(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Mobile

	// 查询用户是否存在
	var user model.User
	ok := false
	if user, ok = dbUserByMobile(pc.Ctrl, c, mobile); !ok {
		return
	}
	if !ensureUserHasEmail(pc.Ctrl, c, &user) {
		return
	}
	sendEmailNStore(pc.Ctrl, c, user.Email)
}

// ForgetSendSMSByMobile godoc
//
//	@Summary		知道手机号码，找回密码时，请求手机验证码
//	@Description	需要手机已经认证过。返回Ok
//	@Tags			pub
//	@Accept			json
//	@Param			mobile	body	model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/forget-send-sms-by-mobile [post]
func (pc *PubCtrl) ForgetSendSMSByMobile(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Mobile

	if !ensureMobileExists(pc.Ctrl, c, mobile) {
		return
	}

	sendMobileNStore(pc.Ctrl, c, mobile)
}

// PresetPassword godoc
//
//	@Summary		通过收到的各种验证吗，提交验证，颁发token
//	@Description	返回 Token
//	@Tags			pub
//	@Accept			json
//	@Param			input	body	model.ResetPasswordInput	true	"复位密码所需的信息"
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/preset-password [post]
func (pc *PubCtrl) PresetPassword(c *gin.Context) {
	var resetInput model.PresetPasswordInput
	if err := c.ShouldBindJSON(&resetInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := resetInput.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	var user model.User
	ok := false
	if user, ok = dbUser(pc.Ctrl, c, resetInput.User); !ok {
		return
	}

	redisEKey := ""
	redisMKey := ""
	c0 := context.Background()
	if user.LastEVTime > 0 {
		if redisEKey, ok = checkUserEmailCode(pc.Ctrl, c, &user, resetInput.ECode, true, false); !ok {
			return
		}
	}
	if user.LastMVTime > 0 {
		if redisMKey, ok = checkUserMobileCode(pc.Ctrl, c, &user, resetInput.MCode, true, false); !ok {
			return
		}
	}
	if user.LastGVTime > 0 {
		if !checkUserGCode(pc.Ctrl, c, &user, resetInput.GCode) {
			return
		}
	}
	err := pc.DB.Model(&user).Updates(map[string]interface{}{
		"last_ev_time": user.LastEVTime,
		"last_mv_time": user.LastMVTime,
		"last_gv_time": user.LastGVTime,
	}).Error
	if err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}
	if redisEKey != "" {
		_, _ = pc.RDB.Delete(c0, redisEKey)
	}
	if redisMKey != "" {
		_, _ = pc.RDB.Delete(c0, redisMKey)
	}
	tu := user.ToTU()
	tu2 := web.GetTokenUser(c)
	if tu2 != nil && tu2.BID == tu.BID {
		tu.Auth2 = tu2.Auth2
		tu.SK = tu2.SK
	} else {
		tu.Auth2 = false
		tu.ReSk()
	}

	AToken(pc.Ctrl, c, tu)
}

// ValidateEmail godoc
//
//	@Summary		提交SignUp之前，验证输入邮箱是否可用
//	@Description	只验证邮箱状态，不发送验证码
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			EmailInput	body		model.EmailInput	true	"邮箱"
//	@Success		200			{object}	web.HttpData
//	@Failure		502			{object}	web.HttpMsg
//	@Router			/pub/validate-email [post]
func (pc *PubCtrl) ValidateEmail(c *gin.Context) {
	var ei model.EmailInput
	if err := c.ShouldBindJSON(&ei); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := ei.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	if !ensureEmailNotReg(pc.Ctrl, c, ei.Email) {
		return
	}

	web.GoodResp(c, "Ok")
}

// ValidateMobile godoc
//
//	@Summary		提交SignUp之前，已验证邮箱的情况下，验证输入手机号是否可用
//	@Description	只验手机状态，不发送验证码
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			MobileInput	body		model.MobileInput	true	"手机"
//	@Success		200			{object}	web.HttpData
//	@Failure		502			{object}	web.HttpMsg
//	@Router			/pub/validate-mobile [post]
func (pc *PubCtrl) ValidateMobile(c *gin.Context) {
	var mi model.MobileInput
	if err := c.ShouldBindJSON(&mi); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	if err := mi.Validate(); err != nil {
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mi.Number()

	if !ensureMobileNotReg(pc.Ctrl, c, mobile) {
		return
	}

	web.GoodResp(c, "Ok")
}

// SignOut godoc
//
//	@Summary		退出系统
//	@Description	退出系统
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/signOut [post]
func (pc *PubCtrl) SignOut(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	web.GoodResp(c, "Ok")
}

// PubKey godoc
//
//	@Summary		获取 PublicKey
//	@Description	用于校验 token 的有效性
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/pubKey [get]
func (pc *PubCtrl) PubKey(c *gin.Context) {
	web.GoodResp(c, conf.Conf.AccessTokenPublicKey)
}

// ResEmail godoc
//
//	@Summary		已登录用户再次验证电子邮箱，发起认证是否本人操作
//	@Description	带token操作，无需主动输入参数。后台生成 6 位随机数字邮件发送并保存在 Redis，5分钟过期
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/resEmail [post]
func (pc *PubCtrl) ResEmail(c *gin.Context) {
	tu, ok := shouldGetTu(pc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Email == "" {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	sendEmailNStore(pc.Ctrl, c, user.Email)
}

// ResMobile godoc
//
//	@Summary		提交验证手机号，验证是否本人操作
//	@Description	已经登录情况下操作，无需其他参数。后台生成 6 位随机数字短信发送给用户并保存在 Redis，5分钟过期
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/resMobile [post]
func (pc *PubCtrl) ResMobile(c *gin.Context) {
	tu, ok := shouldGetTu(pc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Mobile == "" {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	sendMobileNStore(pc.Ctrl, c, user.Mobile)
}

// RevEmail godoc
//
//	@Summary		已登录状态下校验邮件验证码，证明是本人操作
//	@Description	成功返回一个新 token
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Param			input			body	model.ReVerifyInput	true	"输入你收到的验证码"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/revEmail [post]
func (pc *PubCtrl) RevEmail(c *gin.Context) {
	var reVerifyInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&reVerifyInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	code := reVerifyInput.Code

	tu, ok := shouldGetTu(pc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Email == "" {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	redisKey, ok := checkUserEmailCode(pc.Ctrl, c, &user, code, false, true)
	if !ok {
		return
	}

	if err := pc.DB.Model(&user).Update("last_ev_time", user.LastEVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}

	_, _ = pc.RDB.Delete(context.Background(), redisKey)

	tu2 := user.ToTU()
	if tu.SK != "" {
		tu2.SK = tu.SK
	} else {
		tu2.ReSk()
	}
	AToken(pc.Ctrl, c, tu2)
}

// RevMobile godoc
//
//	@Summary		校验手机验证码，成功证明是本人手机
//	@Description	成功返回一个 token 可以用于保存当前用户的所有信息
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Param			input			body	model.ReVerifyInput	true	"输入你收到的验证码"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/revMobile [post]
func (pc *PubCtrl) RevMobile(c *gin.Context) {
	var reVerifyInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&reVerifyInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}
	code := reVerifyInput.Code

	tu, ok := shouldGetTu(pc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(pc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Mobile == "" {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		pc.Log.Error().Msg(err.LStr())
		return
	}

	redisKey, ok := checkUserMobileCode(pc.Ctrl, c, &user, code, false, true)
	if !ok {
		return
	}

	if err := pc.DB.Model(&user).Update("last_mv_time", user.LastMVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}
	_, _ = pc.RDB.Delete(context.Background(), redisKey)

	tu2 := user.ToTU()
	if tu.SK != "" {
		tu2.SK = tu.SK
	} else {
		tu2.ReSk()
	}

	AToken(pc.Ctrl, c, tu2)
}

// VerifyGa godoc
//
//	@Summary		单独验证谷歌验证
//	@Description	成功返回一个 token 保存当前用户的所有信息
//	@Tags			pub
//	@Accept			json
//	@Product		json
//	@Param			code			body	model.ReVerifyInput	true	"ga code"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/pub/verifyGa [post]
func (pc *PubCtrl) VerifyGa(c *gin.Context) {
	var gaInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&gaInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		pc.Log.Err(err).Send()
		return
	}

	tu, ok := shouldGetTu(pc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(pc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	v, er := util.ValidateTOTP(gaInput.Code, user.Ga)
	if er != nil || !v {
		web.BadRes(c, util.ErrGaInvalid)
		pc.Log.Error().Msg(er.LStr())
		return
	}
	user.LastGVTime = time.Now().Unix()
	if err := pc.DB.Model(&user).Update("last_gv_time", user.LastGVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		pc.Log.Err(err).Send()
		return
	}
	tu2 := user.ToTU()
	if tu.SK != "" {
		tu2.SK = tu.SK
	} else {
		tu2.ReSk()
	}

	AToken(pc.Ctrl, c, tu2)
}
