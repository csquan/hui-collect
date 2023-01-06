package model

import (
	"encoding/hex"
	"fmt"
	"github.com/nyaruka/phonenumbers"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"user/pkg/util"
	"user/pkg/util/ecies"
)

const (
	BidLen = 12
)

type User struct {
	gorm.Model
	BID          string `gorm:"type:varchar(15);uniqueIndex;not null"` // business ID
	Nick         string `gorm:"type:varchar(80)"`
	FirmName     string `gorm:"type:varchar(300);not null"`
	FirmType     uint   `gorm:"not null"`
	FirmVerified int64  `gorm:"not null;default 0"`
	Country      string `gorm:"type:varchar(50)"`
	Email        string `gorm:"type:varchar(120)"`
	Mobile       string `gorm:"type:varchar(20)"`
	Addr         string `gorm:"type:varchar(42)"`
	Password     string `gorm:"type:varchar(100);not null"`
	Status       int8   `gorm:"not null;default 0"` // 0 正常用户 -1 封禁 etc 待定
	Ga           string `gorm:"varchar(100)"`
	LastEVTime   int64  `gorm:"not null"`
	LastMVTime   int64  `gorm:"not null"`
	LastGVTime   int64  `gorm:"not null"`
}

func (u *User) Auth2() bool {
	return u.BID != ""
}

type EmailInput struct {
	Email string `json:"email" example:"alice@example.com"`
}

// Validate the email format is correct.
func (ei *EmailInput) Validate() util.Err {
	_, err := mail.ParseAddress(ei.Email)
	if err != nil {
		return util.ErrEmailFormat(ei.Email)
	}
	return nil
}

type MobileInput struct {
	Mobile string `json:"mobile" binding:"required,max=20" example:"+8613901009988"`
}

// Number must is validated
func (mi *MobileInput) Number() string {
	if mi.Validate() != nil {
		return ""
	}
	num, _ := phonenumbers.Parse(mi.Mobile, "")
	regionNumber := phonenumbers.GetRegionCodeForNumber(num)
	countryCode := phonenumbers.GetCountryCodeForRegion(regionNumber)
	nationalNumber := phonenumbers.GetNationalSignificantNumber(num)
	return fmt.Sprintf("+%d %s", countryCode, nationalNumber)
}

// Validate the phone number is correct.
func (mi *MobileInput) Validate() util.Err {
	num, err := phonenumbers.Parse(mi.Mobile, "")
	if err != nil {
		return util.ErrMobileFormat(mi.Mobile)
	}
	if valid := phonenumbers.IsValidNumber(num); !valid {
		return util.ErrMobileFormat(mi.Mobile)
	}
	return nil
}

type VerifyEmailInput struct {
	Email string `json:"email" example:"alice@example.com"`
	Code  string `json:"code" example:"123456"`
}

// Validate the email format is correct.
func (ve *VerifyEmailInput) Validate() util.Err {
	ei := EmailInput{ve.Email}
	return ei.Validate()
}

type VerifyMobileInput struct {
	MobileInput
	Code string `json:"code" example:"123456"`
}

type ReVerifyInput struct {
	Code string `json:"code" example:"123456"`
}

type GenGaResponse struct {
	Image string `json:"image" example:"base64"`
	Text  string `json:"text" example:"please write down this code"`
}

// TokenUser 放到 jwt 中的用户结构体
type TokenUser struct {
	BID       string `json:"bid"`
	Nick      string `json:"nick"`
	Mobile    string `json:"mobile"`
	Email     string `json:"email"`
	Addr      string `json:"addr"`
	FirmName  string `json:"firm_name"`
	KycTime   int64  `json:"kyc_time"`
	MAuthTime int64  `json:"m_auth_time"`
	EAuthTime int64  `json:"e_auth_time"`
	GAuthTime int64  `json:"g_auth_time"`
	Auth2     bool   `json:"auth2"`
	SK        string `json:"sk"`
}

func (u *User) ToTU() *TokenUser {
	return &TokenUser{
		BID:       u.BID,
		Nick:      u.Nick,
		Mobile:    u.Mobile,
		Email:     u.Email,
		Addr:      u.Addr,
		FirmName:  u.FirmName,
		KycTime:   u.FirmVerified,
		MAuthTime: u.LastMVTime,
		EAuthTime: u.LastEVTime,
		GAuthTime: u.LastGVTime,
		Auth2:     u.Auth2(),
	}
}

func (tu *TokenUser) Name() string {
	if tu.Email != "" {
		return tu.Email
	}
	if tu.Mobile != "" {
		return tu.Mobile
	}
	return tu.BID
}

func (tu *TokenUser) ReSk() {
	tu.SK, _ = util.CryptoRandomString(8)
}

type SignUpInput struct {
	Email    string `json:"email" example:"alice@gmail.com"`
	Mobile   string `json:"mobile" binding:"max=20" example:"+86 13901009988"`
	Password string `json:"password" binding:"required,min=8"`
	FirmName string `json:"firm-name" binding:"required"`
	FirmType uint   `json:"firm-type" binding:"required"`
	Country  string `json:"country" binding:"required"`
}

func passwordRule(password string) util.Err {
	e, n, u, _ := util.VerifyRule(password)
	if e && n && u {
		return nil
	} else {
		return util.ErrBrokenPasswordRule
	}
}

func (su *SignUpInput) Validate() util.Err {
	err := passwordRule(su.Password)
	if err != nil {
		return err
	}
	if su.Mobile != "" {
		mo := MobileInput{su.Mobile}
		return mo.Validate()
	}
	return nil
}

type SignInInput struct {
	User     string `json:"user"  binding:"required" example:"alice@gmail.com/+8613901009988"`
	Password string `json:"password"  binding:"required" example:"********"`
}

func (u *User) MaskedEmail() string {
	if u.Email == "" {
		return ""
	}
	two := strings.Split(u.Email, "@")
	if len(two) != 2 {
		return ""
	}
	if len(two[0]) <= 3 {
		return two[0] + "****@" + two[1]
	} else {
		return two[0][:3] + "****@" + two[1]
	}
}

func (u *User) MaskedMobile() string {
	if u.Mobile == "" {
		return ""
	}
	two := strings.SplitN(u.Mobile, " ", 2)
	if len(two) == 1 {
		return two[0][:5] + "****" + two[0][len(two[0])-2:]
	}
	return two[0] + " " + two[1][:2] + "****" + two[1][len(two[1])-2:]
}

// ForgetResponse user validation status 1/0
type ForgetResponse struct {
	MaskedEmail  string
	MaskedMobile string
	E            bool
	M            bool
	G            bool
}

type PresetPasswordInput struct {
	BindGAInput
	User string `json:"user" binding:"required" example:"alice@gmail.com/+8613901009988"`
}

type ResetPasswordInput struct {
	Password string `json:"password" binding:"required,min=8"`
}

func (ri *ResetPasswordInput) Validate() util.Err {
	return passwordRule(ri.Password)
}

type UnbindGAInput struct {
	ECode string `json:"e-code,omitempty"`
	MCode string `json:"m-code,omitempty"`
}

func (ub *UnbindGAInput) Validate() util.Err {
	if ub.MCode == "" && ub.ECode == "" {
		return util.ErrInvalidArgument
	}
	return nil
}

type BindGAInput struct {
	UnbindGAInput
	GCode string `json:"g-code,omitempty"`
}

type ChangeInput struct {
	OldCode string `json:"old-code" binding:"required"`
	NewCode string `json:"new-code" binding:"required"`
}

type ChangeEmailInput struct {
	ChangeInput
	Email string `json:"email" binding:"required"`
}

func (ce *ChangeEmailInput) Validate() util.Err {
	ei := EmailInput{ce.Email}
	return ei.Validate()
}

type ChangeMobileInput struct {
	ChangeInput
	MobileInput
}

type BindEmailInput struct {
	BindGAInput
	Email string `json:"email" binding:"required" example:"alice@gmail.com"`
}

func (be *BindEmailInput) Validate() util.Err {
	ei := EmailInput{Email: be.Email}
	return ei.Validate()
}

type BindMobileInput struct {
	BindGAInput
	Mobile string `json:"mobile" binding:"required" example:"+86 13901009988"`
}

func (bm *BindMobileInput) Validate() util.Err {
	mi := MobileInput{Mobile: bm.Mobile}
	return mi.Validate()
}

type FirmConfirmed struct {
	Verify
	BID      string `json:"bid" binding:"required" example:"1233210123"`
	FirmName string `json:"firm-name" binding:"required" example:"一地鸡毛蒜皮小公司"`
	FirmType uint   `json:"firm-type" binding:"required" example:"2"`
	Country  string `json:"country" binding:"required" example:"+86"`
}

func (fc *FirmConfirmed) Validate(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	return fc.verify(pk, log)
}

type BindData struct {
	UID string `json:"uid"`
}

type Verify struct {
	Verified string `json:"verified" binding:"required" example:"BASE198964"`
}

func (v *Verify) verify(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	byt, err := hex.DecodeString(v.Verified)
	if err != nil {
		return util.ErrMsgDecode
	}
	dt, er := pk.Decrypt(byt, nil, nil)
	if er != nil {
		return er
	}
	t1, err := time.Parse(http.TimeFormat, string(dt))
	if err != nil {
		return util.ErrMsgDecrypt
	}
	span := time.Now().Sub(t1)
	if span > 3*time.Second || span < -3*time.Second {
		log.Debug().Dur("span", span).Send()
		return util.ErrMsgDecrypt
	}
	return nil
}

type FirmQuery struct {
	Verify
	Uid    string `json:"uid,omitempty"`
	Email  string `json:"email,omitempty"`
	Mobile string `json:"mobile,omitempty"`
}

func (fq *FirmQuery) Number() string {
	mi := MobileInput{Mobile: fq.Mobile}
	if mi.Validate() != nil {
		return ""
	}
	return mi.Number()
}

func (fq *FirmQuery) Validate(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	if fq.Uid == "" && fq.Email == "" && fq.Mobile == "" {
		return util.ErrTokenInvalid
	}
	return fq.verify(pk, log)
}

type FirmResp struct {
	Uid          string
	Nick         string
	Email        string
	Mobile       string
	Addr         string
	FirmName     string
	Country      string
	FirmVerified int64
	Status       int8
	Created      int64
}

func (u *User) ToFr() *FirmResp {
	return &FirmResp{
		Uid:          u.BID,
		Nick:         u.Nick,
		Email:        u.Email,
		Mobile:       u.Mobile,
		Addr:         u.Addr,
		FirmName:     u.FirmName,
		Country:      u.Country,
		FirmVerified: u.FirmVerified,
		Status:       u.Status,
		Created:      u.CreatedAt.Unix(),
	}
}

type AddrInput struct {
	Verify
	Addr string `json:"addr" binding:"required" example:"0x8cbf3d676bab7e93e94a9a2de153aff1e2f3124c"`
}

const HexChars = "0123456789ABCDEFabcdef"

func (ai *AddrInput) Validate(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	if len(ai.Addr) != 42 {
		return util.ErrAddr
	}
	if ai.Addr[:2] != "0x" {
		return util.ErrAddr
	}
	for i := 0; i < 40; i++ {
		if strings.IndexByte(HexChars, ai.Addr[i+2]) < 0 {
			return util.ErrAddr
		}
	}
	return ai.verify(pk, log)
}

type KycUserStatusInput struct {
	Verify
	Uid    string `json:"uid" binding:"required" example:"323567890"`
	Status int8   `json:"status" binding:"required" example:"-1=disable, 0=enable"`
}

func (usi *KycUserStatusInput) Validate(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	if len(usi.Uid) != BidLen {
		return util.ErrInvalidArgument
	}
	return usi.verify(pk, log)
}

type KycUserListInput struct {
	Verify
	Start int64 `json:"start" binding:"required" example:"1672728135"`
	End   int64 `json:"end" binding:"required" example:"1672728135"`
	Page  int   `json:"page" binding:"required" example:"1"`
	Limit int   `json:"limit" binding:"required" example:"20"`
}

func (uli *KycUserListInput) Validate(pk *ecies.PrivateKey, log *zerolog.Logger) util.Err {
	return uli.verify(pk, log)
}

func (uli *KycUserListInput) GetLimit() int {
	if uli.Limit < 10 || uli.Limit > 150 {
		return 20
	} else {
		return uli.Limit
	}
}

func (uli *KycUserListInput) GetFrom() int {
	if uli.Page <= 0 {
		uli.Page = 1
	}
	return (uli.Page - 1) * uli.GetLimit()
}

type UserPage struct {
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Rows  []FirmResp `json:"rows"`
}

type NickInput struct {
	Nick string `json:"nick" binding:"required" example:"MadDog"`
}
