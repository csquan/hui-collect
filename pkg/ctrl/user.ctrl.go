package ctrl

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"net/http"
	"strings"
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

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type UserCtrl struct {
	*Ctrl
	PubK *ecies.PublicKey
	Cli  *resty.Client
}

func NewUserCtrl() UserCtrl {
	rdb := redis.NewStore("go:user:controller:")
	logger := log.Log.With().Str("ctrl", "user").Logger()
	pubK, err := ecies.PublicFromString(conf.Conf.KeySvrPubKey)
	if err != nil {
		logger.Error().Msg(err.LStr())
		panic(err)
	}
	client := resty.New()
	client.JSONMarshal = json.Marshal
	client.JSONUnmarshal = json.Unmarshal
	client.SetBaseURL(conf.Conf.KeySvrUrl)
	return UserCtrl{Ctrl: &Ctrl{DB: db.DB, RDB: rdb, Log: &logger}, PubK: pubK, Cli: client}
}

// GenGa godoc
//
//	@Summary		在校验手机或者校验邮箱成功后，创建谷歌验证
//	@Description	生成绑定谷歌验证需要的二维码
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/genGa [post]
func (uc *UserCtrl) GenGa(c *gin.Context) {
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	if tu.BID == "" {
		web.BadRes(c, util.ErrUnAuthed)
		uc.Log.Debug().Interface("tu", tu).Send()
		return
	}

	if _, ok := dbUserByID(uc.Ctrl, c, tu.BID); !ok {
		return
	}
	if gr, ok := genGa(uc.Ctrl, c, tu); ok {
		web.GoodResp(c, gr)
	}
}

// BindGa godoc
//
//	@Summary		绑定谷歌验证
//	@Description	成功返回一个 token 保存当前用户的所有信息
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			code			body	model.BindGAInput	true	"ga code & (email-code | sms-code)"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/bindGa [post]
func (uc *UserCtrl) BindGa(c *gin.Context) {
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	var bindGAInput model.BindGAInput
	if err := c.ShouldBindJSON(&bindGAInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}

	var user model.User
	user, ok = dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	redisEKey := ""
	if user.LastEVTime > 0 {
		redisEKey, ok = checkUserEmailCode(uc.Ctrl, c, &user, bindGAInput.ECode, true, true)
		if !ok {
			return
		}
	}
	redisMKey := ""
	if user.LastMVTime > 0 {
		redisMKey, ok = checkUserMobileCode(uc.Ctrl, c, &user, bindGAInput.MCode, true, true)
		if !ok {
			return
		}
	}
	redisGKey, ok := bindGa(uc.Ctrl, c, &user, bindGAInput.GCode)
	if !ok {
		return
	}
	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"ga":           user.Ga,
		"last_gv_time": user.LastGVTime,
		"last_ev_time": user.LastEVTime,
		"last_mv_time": user.LastMVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	c0 := context.Background()
	if redisEKey != "" {
		_, _ = uc.RDB.Delete(c0, redisEKey)
	}
	if redisMKey != "" {
		_, _ = uc.RDB.Delete(c0, redisMKey)
	}
	_, _ = uc.RDB.Delete(c0, redisGKey)
	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// SignUp godoc
//
//	@Summary		提交用户信息表单
//	@Description	此时用户已经验证完Email或者Mobile中的至少一项，已经具备token才能提交信息
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Param			user			body	model.SignUpInput	true	"user information"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/signUp [post]
func (uc *UserCtrl) SignUp(c *gin.Context) {
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	var signUpInput model.SignUpInput
	if err := c.ShouldBindJSON(&signUpInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if data, err := base64.StdEncoding.DecodeString(signUpInput.Password); err != nil {
		web.BadRes(c, util.ErrMsgDecode)
		uc.Log.Err(err).Send()
		return
	} else {
		signUpInput.Password = string(data)
	}

	if err := signUpInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if tu.Email != "" && signUpInput.Email != tu.Email {
		err := util.ErrEmailNotEq(tu.Email, signUpInput.Email)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if tu.Mobile != "" && signUpInput.Mobile != tu.Mobile {
		err := util.ErrMobileNotEq(tu.Mobile, signUpInput.Mobile)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	var user model.User
	email := ""
	if tu.Email != "" {
		email = tu.Email
	} else if signUpInput.Email != "" {
		email = signUpInput.Email
	}
	if email != "" {
		if !ensureEmailNotReg(uc.Ctrl, c, email) {
			return
		}
	}
	mobile := ""
	if tu.Mobile != "" {
		mobile = tu.Mobile
	} else if signUpInput.Mobile != "" {
		mi := model.MobileInput{Mobile: signUpInput.Mobile}
		mobile = mi.Number()
	}
	if mobile != "" {
		if !ensureMobileNotReg(uc.Ctrl, c, mobile) {
			return
		}
	}

	bid := uc.mustBid()
	hashedPass, err := util.HashPassword(signUpInput.Password)
	if err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	bd := model.BindData{UID: bid}
	bdMsg, _ := json.Marshal(bd)
	enc, err := ecies.Encrypt(rand.Reader, uc.PubK, bdMsg, nil, nil)
	if err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	var result string
	resp, er := uc.Cli.R().
		SetFormData(map[string]string{"data": hex.EncodeToString(enc)}).
		SetResult(&result).
		Post("/bind")
	if er != nil || resp.StatusCode() != http.StatusOK {
		err := util.ErrWalletSvr
		web.BadRes(c, err)
		uc.Log.Err(er).Send()
		return
	}

	nick := uc.mustNick()
	user = model.User{
		BID:          bid,
		Nick:         nick,
		Email:        signUpInput.Email,
		Mobile:       mobile,
		Addr:         result,
		Password:     hashedPass,
		FirmName:     signUpInput.FirmName,
		FirmType:     signUpInput.FirmType,
		Country:      signUpInput.Country,
		FirmVerified: 0,
		Status:       0,
		LastMVTime:   tu.MAuthTime,
		LastEVTime:   tu.EAuthTime,
		LastGVTime:   tu.GAuthTime,
	}
	er = uc.DB.Model(&user).Create(&user).Error
	if er != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(er).Interface("user", user).Send()
		return
	}
	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

func (uc *UserCtrl) mustBid() string {
	bid, _ := util.CryptoRandomNumerical(model.BidLen)
	var user model.User
	err := uc.DB.Model(&user).First(&user, "b_id=?", bid).Error
	if err == nil {
		return uc.mustBid()
	} else {
		if strings.Contains(err.Error(), "not found") {
			return bid
		} else {
			return uc.mustBid()
		}
	}
}

func (uc *UserCtrl) mustNick() string {
	nick, _ := util.CryptoRandomString(6)
	nick = "Anonymous-User-" + nick
	var user model.User
	err := uc.DB.Model(&user).First(&user, "nick=?", nick).Error
	if err == nil {
		return uc.mustNick()
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nick
		} else {
			return uc.mustNick()
		}
	}
}

// ResEmail godoc
//
//	@Summary		已登录用户再次验证电子邮箱，发起认证是否本人操作
//	@Description	带token操作，无需主动输入参数。后台生成 6 位随机数字邮件发送并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/resEmail [post]
func (uc *UserCtrl) ResEmail(c *gin.Context) {
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Email == "" {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	sendEmailNStore(uc.Ctrl, c, user.Email)
}

// ResMobile godoc
//
//	@Summary		提交验证手机号，验证是否本人操作
//	@Description	已经登录情况下操作，无需其他参数。后台生成 6 位随机数字短信发送给用户并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string	true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/resMobile [post]
func (uc *UserCtrl) ResMobile(c *gin.Context) {
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Mobile == "" {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	sendMobileNStore(uc.Ctrl, c, user.Mobile)
}

// RevEmail godoc
//
//	@Summary		已登录状态下校验邮件验证码，证明是本人操作
//	@Description	成功返回一个新 token
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Param			input			body	model.ReVerifyInput	true	"输入你收到的验证码"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/revEmail [post]
func (uc *UserCtrl) RevEmail(c *gin.Context) {
	var reVerifyInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&reVerifyInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	code := reVerifyInput.Code

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Email == "" {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	redisKey, ok := checkUserEmailCode(uc.Ctrl, c, &user, code, false, true)
	if !ok {
		return
	}

	if err := uc.DB.Model(&user).Update("last_ev_time", user.LastEVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	_, _ = uc.RDB.Delete(context.Background(), redisKey)

	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// RevMobile godoc
//
//	@Summary		校验手机验证码，成功证明是本人手机
//	@Description	成功返回一个 token 可以用于保存当前用户的所有信息
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Param			input			body	model.ReVerifyInput	true	"输入你收到的验证码"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/revMobile [post]
func (uc *UserCtrl) RevMobile(c *gin.Context) {
	var reVerifyInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&reVerifyInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	code := reVerifyInput.Code

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}

	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.Mobile == "" {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	redisKey, ok := checkUserMobileCode(uc.Ctrl, c, &user, code, false, true)
	if !ok {
		return
	}

	if err := uc.DB.Model(&user).Update("last_mv_time", user.LastMVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	_, _ = uc.RDB.Delete(context.Background(), redisKey)

	tu2 := user.ToTU()
	tu2.SK = tu.SK

	AToken(uc.Ctrl, c, tu2)
}

// MyEmail godoc
//
//	@Summary		已生成用户并登录手机的情况下提交电子邮箱，要求发起认证真实性
//	@Description	后台生成 6 位随机数字邮件发送并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			email			body	model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/myEmail [post]
func (uc *UserCtrl) MyEmail(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	if !ensureEmailNotReg(uc.Ctrl, c, email) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if tu.Mobile == "" {
		web.BadRes(c, util.ErrUnAuthed)
		uc.Log.Debug().Interface("tu=", tu).Send()
		return
	}
	if user.LastEVTime > 0 && user.Email != "" {
		err := util.ErrEmailAlready(user.Email)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	user.Email = email
	if err := uc.DB.Model(&user).Update("email", user.Email).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	sendEmailNStore(uc.Ctrl, c, email)
}

// MyMobile godoc
//
//	@Summary		已注册并登录邮箱的情况下提交手机号，要求发起认证真实性
//	@Description	后台生成 6 位随机数字短信发送给用户并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			mobile			body	model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/myMobile [post]
func (uc *UserCtrl) MyMobile(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Mobile

	if !ensureMobileNotReg(uc.Ctrl, c, mobile) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}
	if user.LastMVTime > 0 && user.Mobile != "" {
		err := util.ErrMobileAlready(user.Mobile)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	user.Mobile = mobile
	if err := uc.DB.Model(&user).Update("mobile", user.Mobile).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	sendMobileNStore(uc.Ctrl, c, mobile)
}

// VerifyGa godoc
//
//	@Summary		单独验证谷歌验证
//	@Description	成功返回一个 token 保存当前用户的所有信息
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			code			body	model.ReVerifyInput	true	"ga code"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/verifyGa [post]
func (uc *UserCtrl) VerifyGa(c *gin.Context) {
	var gaInput model.ReVerifyInput
	if err := c.ShouldBindJSON(&gaInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	v, er := util.ValidateTOTP(gaInput.Code, user.Ga)
	if er != nil || !v {
		web.BadRes(c, util.ErrGaInvalid)
		uc.Log.Error().Msg(er.LStr())
		return
	}
	user.LastGVTime = time.Now().Unix()
	if err := uc.DB.Model(&user).Update("last_gv_time", user.LastGVTime).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	tu2 := user.ToTU()
	tu2.SK = tu.SK

	AToken(uc.Ctrl, c, tu2)
}

// MyNewEmail godoc
//
//	@Summary		已经认证过电子邮箱的情况下，要求换电子邮箱。如果从未验证过邮箱，需要调用 myEmail 接口
//	@Description	提交新电子邮箱，要求发起认证真实性。后台生成 6 位随机数字邮件发送并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			email			body	model.EmailInput	true	"在 email 字段填入你的邮箱"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/myNewEmail [post]
func (uc *UserCtrl) MyNewEmail(c *gin.Context) {
	var emailInput model.EmailInput
	if err := c.ShouldBindJSON(&emailInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := emailInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	email := emailInput.Email

	if !ensureEmailNotReg(uc.Ctrl, c, email) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}
	if user.LastEVTime == 0 {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	sendEmailNStore(uc.Ctrl, c, email)
}

// MyNewMobile godoc
//
//	@Summary		已经认证过手机的情况下要求更换手机号，要求对新手机号发起真实性认证。
//	@Description	后台生成 6 位随机数字短信发送给用户并保存在 Redis，5分钟过期
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			mobile			body	model.MobileInput	true	"在 mobile 字段填入你的手机号，带国家字冠"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/myNewMobile [post]
func (uc *UserCtrl) MyNewMobile(c *gin.Context) {
	var mobileInput model.MobileInput
	if err := c.ShouldBindJSON(&mobileInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := mobileInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := mobileInput.Number()

	if !ensureMobileNotReg(uc.Ctrl, c, mobile) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastMVTime == 0 {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	sendMobileNStore(uc.Ctrl, c, mobile)
}

// ChangeEmail godoc
//
//	@Summary		只认证了电子邮箱的情况下，要求换电子邮箱。提交新邮箱和验证码。
//	@Description	认证成功更换邮箱。要求刚验证过旧邮箱
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			email			body	model.VerifyEmailInput	true	"新邮箱和验证码"
//	@Param			Authorization	header	string					true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/changeEmail [post]
func (uc *UserCtrl) ChangeEmail(c *gin.Context) {
	var ceInput model.VerifyEmailInput
	if err := c.ShouldBindJSON(&ceInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := ceInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	email := ceInput.Email

	if !ensureEmailNotReg(uc.Ctrl, c, email) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	if !recentToken(uc.Ctrl, c, tu) {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastEVTime == 0 {
		err := util.ErrEmailNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastMVTime > 0 {
		err := util.ErrEmailByMobile(user.Mobile)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastGVTime > 0 {
		err := util.ErrEmailByGa
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	if !emailCodeExists(uc.Ctrl, c, email, ceInput.Code) {
		return
	}

	user.Email = email
	user.LastEVTime = time.Now().Unix()

	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"email":        user.Email,
		"last_ev_time": user.LastEVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// ChangeMobile godoc
//
//	@Summary		只认证了手机的情况下，要求换手机号码。提交新手机和验证码。
//	@Description	认证成功更换手机号。要求刚认证过旧手机
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			mobile			body	model.VerifyMobileInput	true	"新手机和验证码"
//	@Param			Authorization	header	string					true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/changeMobile [post]
func (uc *UserCtrl) ChangeMobile(c *gin.Context) {
	var cmInput model.VerifyMobileInput
	if err := c.ShouldBindJSON(&cmInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := cmInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := cmInput.Number()

	if !ensureMobileNotReg(uc.Ctrl, c, mobile) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	if !recentToken(uc.Ctrl, c, tu) {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastMVTime == 0 {
		err := util.ErrMobileNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastEVTime > 0 {
		err := util.ErrMobileByEmail(user.Email)
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastGVTime > 0 {
		err := util.ErrMobileByGa
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	if !mobileCodeExists(uc.Ctrl, c, mobile, cmInput.Code) {
		return
	}

	user.Mobile = mobile
	user.LastMVTime = time.Now().Unix()
	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"mobile":       user.Mobile,
		"last_mv_time": user.LastMVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// BindEmailBy godoc
//
//	@Summary		已有手机和/或谷歌认证，要求换电子邮箱。
//	@Description	提交新邮箱和新邮箱验证码，认证成功更换邮箱。前提是刚验证过手机和/或谷歌验证
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			email			body	model.VerifyEmailInput	true	"新邮箱和验证码"
//	@Param			Authorization	header	string					true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/bindEmailBy [post]
func (uc *UserCtrl) BindEmailBy(c *gin.Context) {
	var beInput model.VerifyEmailInput
	if err := c.ShouldBindJSON(&beInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := beInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	email := beInput.Email

	if !ensureEmailNotReg(uc.Ctrl, c, email) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastMVTime == 0 && user.LastGVTime == 0 {
		err := util.ErrMobileGaNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastMVTime > 0 {
		if time.Now().Unix()-user.LastMVTime > 300 {
			err := util.ErrMobileFirst
			web.BadRes(c, err)
			uc.Log.Error().Msg(err.LStr())
			return
		}
	}
	if user.LastGVTime > 0 {
		if time.Now().Unix()-user.LastGVTime > 300 {
			err := util.ErrGaFirst
			web.BadRes(c, err)
			uc.Log.Error().Msg(err.LStr())
			return
		}
	}

	if !emailCodeExists(uc.Ctrl, c, email, beInput.Code) {
		return
	}

	user.Email = email
	user.LastEVTime = time.Now().Unix()
	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"email":        user.Email,
		"last_ev_time": user.LastEVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// BindMobileBy godoc
//
//	@Summary		认证过邮箱和/或谷歌验证的情况下，要求换手机号码。提交新手机和验证码。
//	@Description	认证成功更换手机号。前提是刚验证过邮箱和/或谷歌验证
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			mobile			body	model.VerifyMobileInput	true	"新手机和验证码"
//	@Param			Authorization	header	string					true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/bindMobileBy [post]
func (uc *UserCtrl) BindMobileBy(c *gin.Context) {
	var bmInput model.VerifyMobileInput
	if err := c.ShouldBindJSON(&bmInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := bmInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	mobile := bmInput.Number()

	if !ensureMobileNotReg(uc.Ctrl, c, mobile) {
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastEVTime == 0 && user.LastGVTime == 0 {
		err := util.ErrEmailGaNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastEVTime > 0 {
		if time.Now().Unix()-user.LastEVTime > 300 {
			err := util.ErrEmailFirst
			web.BadRes(c, err)
			uc.Log.Error().Msg(err.LStr())
			return
		}
	}
	if user.LastGVTime > 0 {
		if time.Now().Unix()-user.LastGVTime > 300 {
			err := util.ErrGaFirst
			web.BadRes(c, err)
			uc.Log.Error().Msg(err.LStr())
			return
		}
	}

	if !mobileCodeExists(uc.Ctrl, c, mobile, bmInput.Code) {
		return
	}

	user.Mobile = mobile
	user.LastMVTime = time.Now().Unix()

	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"mobile":       mobile,
		"last_mv_time": user.LastMVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	tu2 := user.ToTU()
	tu2.SK = tu.SK
	AToken(uc.Ctrl, c, tu2)
}

// UnbindGa godoc
//
//	@Summary		解绑谷歌验证
//	@Description	成功返回一个 token 包含当前用户更新后的信息
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			code			body	model.UnbindGAInput	true	"(email-code & | sms-code)"
//	@Param			Authorization	header	string				true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/unbindGa [post]
func (uc *UserCtrl) UnbindGa(c *gin.Context) {
	var ubInput model.UnbindGAInput
	if err := c.ShouldBindJSON(&ubInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if err := ubInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	redisEKey := ""
	if user.LastEVTime > 0 {
		if redisEKey, ok = checkUserEmailCode(uc.Ctrl, c, &user, ubInput.ECode, true, true); !ok {
			return
		}
	}
	redisMKey := ""
	if user.LastMVTime > 0 {
		if redisMKey, ok = checkUserMobileCode(uc.Ctrl, c, &user, ubInput.MCode, true, true); !ok {
			return
		}
	}

	user.Ga = ""
	user.LastGVTime = 0
	if err := uc.DB.Model(&user).Updates(map[string]interface{}{
		"ga":           user.Ga,
		"last_gv_time": user.LastGVTime,
		"last_ev_time": user.LastEVTime,
		"last_mv_time": user.LastMVTime,
	}).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	c0 := context.Background()
	if redisEKey != "" {
		_, _ = uc.RDB.Delete(c0, redisEKey)
	}
	if redisMKey != "" {
		_, _ = uc.RDB.Delete(c0, redisMKey)
	}

	tu2 := user.ToTU()
	tu2.SK = tu.SK

	AToken(uc.Ctrl, c, tu2)
}

// DoResetPassword godoc
//
//	@Summary		认证完2FA的所有认证，直接提交新密码
//	@Description	刚认证完所有2FA认证，只需要提交新密码即可复位密码
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			newPassword		body	model.ResetPasswordInput	true	"新密码，需要符合密码规则"
//	@Param			Authorization	header	string						true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/do-reset-password [post]
func (uc *UserCtrl) DoResetPassword(c *gin.Context) {
	var rpInput model.ResetPasswordInput
	if err := c.ShouldBindJSON(&rpInput); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if data, err := base64.StdEncoding.DecodeString(rpInput.Password); err != nil {
		web.BadRes(c, util.ErrMsgDecode)
		uc.Log.Err(err).Send()
		return
	} else {
		rpInput.Password = string(data)
	}

	if err := rpInput.Validate(); err != nil {
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	// 根据 token 拿到用户
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if user.LastMVTime == 0 && user.LastEVTime == 0 {
		err := util.ErrEmailMobileNo
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}
	if user.LastEVTime > 0 && time.Now().Unix()-user.LastEVTime > 300 {
		err := util.ErrEmailFirst
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return

	}
	if user.LastMVTime > 0 && time.Now().Unix()-user.LastMVTime > 300 {
		err := util.ErrMobileFirst
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return

	}
	if user.LastGVTime > 0 && time.Now().Unix()-user.LastGVTime > 300 {
		err := util.ErrGaFirst
		web.BadRes(c, err)
		uc.Log.Error().Msg(err.LStr())
		return
	}

	hashedPass, er := util.HashPassword(rpInput.Password)
	if er != nil {
		web.BadRes(c, er)
		uc.Log.Error().Msg(er.LStr())
		return
	}
	user.Password = hashedPass
	if err := uc.DB.Model(&user).Update("password", hashedPass).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}
	web.GoodResp(c, "Ok")
}

// ChangeNick godoc
//
//	@Summary		修改自己的昵称
//	@Description	返回带有新昵称的token
//	@Tags			user
//	@Accept			json
//	@Product		json
//	@Param			Nick			body	model.NickInput	true	"新昵称"
//	@Param			Authorization	header	string			true	"Authentication header"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	web.HttpData
//	@Failure		502	{object}	web.HttpMsg
//	@Router			/user/change-nick [post]
func (uc *UserCtrl) ChangeNick(c *gin.Context) {
	var ni model.NickInput
	if err := c.ShouldBindJSON(&ni); err != nil {
		web.BadRes(c, util.ErrInvalidArgument)
		uc.Log.Err(err).Send()
		return
	}
	if !ensureNickNotReg(uc.Ctrl, c, ni.Nick) {
		return
	}

	// 根据 token 拿到用户
	tu, ok := mustGetTu(uc.Ctrl, c)
	if !ok {
		return
	}
	user, ok := dbUserByID(uc.Ctrl, c, tu.BID)
	if !ok {
		return
	}

	if err := uc.DB.Model(&user).Update("nick", ni.Nick).Error; err != nil {
		web.BadRes(c, util.ErrDB)
		uc.Log.Err(err).Send()
		return
	}

	user.Nick = ni.Nick
	tu2 := user.ToTU()
	tu2.SK = tu.SK
	tu2.Auth2 = tu.Auth2
	AToken(uc.Ctrl, c, tu2)
}
