package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"math"
	"strings"
	"taotie/core/config"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/session"
	"taotie/core/util"
	"taotie/core/util/mail"
	"time"
)

type RegisterUserRequest struct {
	Name       string `json:"name" validate:"required,alphanumunicode"`
	NickName   string `json:"nick_name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"alphanumunicode"`
	RePassword string `json:"repassword" validate:"eqfield=Password"`
	Gender     int    `json:"gender" validate:"oneof=0 1 2"`
	Describe   string `json:"describe"`
	ImagePath  string `json:"image_path"`
}

// User register, anyone can use email register
func RegisterUser(c *gin.Context) {
	resp := new(Resp)
	req := new(RegisterUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	// if close register direct return
	if !config.GlobalConfig.DefaultConfig.CanMail {
		resp.Error = Error(CloseRegisterError, "")
		return
	}

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("RegisterUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// name can not repeat and prefix with @
	u := new(model.User)
	if strings.Contains(req.Name, "@") {
		flog.Log.Errorf("RegisterUser err: %s", "@ can not be")
		resp.Error = Error(ParasError, "@ can not be")
		return
	}

	u.Name = req.Name
	repeat, err := u.IsNameRepeat()
	if err != nil {
		flog.Log.Errorf("RegisterUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if repeat {
		flog.Log.Errorf("RegisterUser err: %s", "name already use by other")
		resp.Error = Error(UserNameAlreadyBeUsed, "")
		return
	}

	// nickname also must unique
	u.NickName = req.NickName

	// email also
	u.Email = req.Email
	repeat, err = u.IsEmailRepeat()
	if err != nil {
		flog.Log.Errorf("RegisterUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if repeat {
		flog.Log.Errorf("RegisterUser err: %s", "email already use by other")
		resp.Error = Error(EmailAlreadyBeUsed, "")
		return
	}

	// activate code gen
	u.ActivateCode = util.GetGUID()
	u.ActivateCodeExpired = time.Now().Add(5 * time.Minute).Unix()
	u.Describe = req.Describe
	u.Password = req.Password
	u.Gender = req.Gender

	// send email
	mm := new(mail.Message)
	mm.Sender = config.GlobalConfig.MailConfig
	mm.To = u.Email
	mm.ToName = u.NickName
	mm.Body = fmt.Sprintf(mm.Body, "Register", u.ActivateCode)
	err = mm.Sent()
	if err != nil {
		flog.Log.Errorf("RegisterUser err:%s", err.Error())
		resp.Error = Error(EmailSendError, err.Error())
		return
	}

	err = u.InsertOne()
	if err != nil {
		flog.Log.Errorf("RegisterUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if debug will return some info
	if AuthDebug {
		resp.Data = u
	}

	resp.Flag = true
}

// Create user, admin url
func CreateUser(c *gin.Context) {
	resp := new(Resp)
	req := new(RegisterUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("CreateUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	u := new(model.User)
	if strings.Contains(req.Name, "@") {
		flog.Log.Errorf("CreateUser err: %s", "@ can not be")
		resp.Error = Error(ParasError, "@ can not be")
		return
	}

	u.Name = req.Name
	repeat, err := u.IsNameRepeat()
	if err != nil {
		flog.Log.Errorf("CreateUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if repeat {
		flog.Log.Errorf("CreateUser err: %s", "name already use by other")
		resp.Error = Error(UserNameAlreadyBeUsed, "")
		return
	}

	u.NickName = req.NickName

	// email check
	u.Email = req.Email
	repeat, err = u.IsEmailRepeat()
	if err != nil {
		flog.Log.Errorf("CreateUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if repeat {
		flog.Log.Errorf("CreateUser err: %s", "email already use by other")
		resp.Error = Error(EmailAlreadyBeUsed, "")
		return
	}

	// if image not empty
	if req.ImagePath != "" {
		u.HeadPhoto = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			flog.Log.Errorf("CreateUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			flog.Log.Errorf("CreateUser err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}
	}

	u.Describe = req.Describe
	u.NickName = req.NickName
	u.Password = req.Password
	u.Gender = req.Gender

	// default is activate
	u.Status = 1
	u.ActivateTime = time.Now().Unix()
	err = u.InsertOne()
	if err != nil {
		flog.Log.Errorf("CreateUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	resp.Data = u
}

type ActivateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

// Activate by oneself
func ActivateUser(c *gin.Context) {
	resp := new(Resp)
	req := new(ActivateUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ActivateUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// email and activate code must together
	u := new(model.User)
	u.ActivateCode = req.Code
	u.Email = req.Email

	// whether exist
	exist, err := u.IsActivateCodeExist()
	if err != nil {
		flog.Log.Errorf("ActivateUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("ActivateUser err:%s", "not exist code")
		resp.Error = Error(ActivateCodeWrong, "not exist code")
		return
	}

	// has been activate direct return
	if u.Status != 0 {
		resp.Flag = true
		return
	}

	// activate code expired, must gen again
	if u.ActivateCodeExpired < time.Now().Unix() {
		flog.Log.Errorf("ActivateUser err:%s", "code expired")
		resp.Error = Error(ActivateCodeExpired, "")
		return
	} else {
		u.Status = 1
		err = u.UpdateActivateStatus()
		if err != nil {
			flog.Log.Errorf("ActivateUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		// activate success will soon set session
		token, err := SetUserSession(u)
		if err != nil {
			flog.Log.Errorf("ActivateUser err:%s", err.Error())
			resp.Error = Error(SetUserSessionError, err.Error())
			return
		}

		// return token
		resp.Data = token
	}

	resp.Flag = true
}

type ResendActivateCodeToUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// Activate code expire must resent email ang get new one
func ResendActivateCodeToUser(c *gin.Context) {
	resp := new(Resp)
	req := new(ResendActivateCodeToUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ResendActivateCodeToUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// get user info by email
	u := new(model.User)
	u.Email = req.Email
	ok, err := u.GetUserByEmail()
	if err != nil {
		flog.Log.Errorf("ResendActivateCodeToUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !ok {
		flog.Log.Errorf("ResendActivateCodeToUser err:%s", "email not found")
		resp.Error = Error(EmailNotFound, "")
		return
	}

	if u.Status != 0 {
		resp.Flag = true
		return
	} else if u.ActivateCodeExpired > time.Now().Unix() {
		// can not gen a new code because expire time not reach
		flog.Log.Errorf("ResendUser err:%s", "code not expired")
		resp.Error = Error(ActivateCodeNotExpired, "")
		return
	}

	// update activate code, expire after 5 min
	err = u.UpdateActivateCode()
	if err != nil {
		flog.Log.Errorf("ResendUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// send email
	mm := new(mail.Message)
	mm.Sender = config.GlobalConfig.MailConfig
	mm.To = u.Email
	mm.ToName = u.NickName
	mm.Body = fmt.Sprintf(mm.Body, "Register", u.ActivateCode)
	err = mm.Sent()
	if err != nil {
		flog.Log.Errorf("ResendUser err:%s", err.Error())
		resp.Error = Error(EmailSendError, err.Error())
		return
	}
	resp.Flag = true
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// When user want to modify password or forget password, can gen a code to change password
func ForgetPasswordOfUser(c *gin.Context) {
	resp := new(Resp)
	req := new(ForgetPasswordRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("RegisterUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	u := new(model.User)
	u.Email = req.Email
	ok, err := u.GetUserByEmail()
	if err != nil {
		flog.Log.Errorf("ForgetPassword err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !ok {
		flog.Log.Errorf("ForgetPassword err:%s", "email not found")
		resp.Error = Error(EmailNotFound, "")
		return
	}

	// only code expired can gen a new one again
	if u.ResetCodeExpired < time.Now().Unix() {
		// code is valid in 5 min
		err = u.UpdateCode()
		if err != nil {
			flog.Log.Errorf("ForgetPassword comerr:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		// send email
		mm := new(mail.Message)
		mm.Sender = config.GlobalConfig.MailConfig
		mm.To = u.Email
		mm.ToName = u.NickName
		mm.Body = fmt.Sprintf(mm.Body, "Forget Password", u.ResetCode)
		err = mm.Sent()
		if err != nil {
			flog.Log.Errorf("ForgetPassword err:%s", err.Error())
			resp.Error = Error(EmailSendError, err.Error())
			return
		}

	} else {
		flog.Log.Errorf("ForgetPassword err:%s", "reset code expired time not reach")
		resp.Error = Error(ResetCodeExpiredTimeNotReach, "")
		return
	}

	resp.Flag = true
}

type ChangePasswordRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Code       string `json:"code" validate:"required"`
	Password   string `json:"password" validate:"alphanumunicode"`
	RePassword string `json:"repassword" validate:"eqfield=Password"`
}

// Change password by a forget password email code
func ChangePasswordOfUser(c *gin.Context) {
	resp := new(Resp)
	req := new(ChangePasswordRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ChangePassword err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	u := new(model.User)
	u.Email = req.Email
	ok, err := u.GetUserByEmail()
	if err != nil {
		flog.Log.Errorf("ChangePassword err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !ok {
		flog.Log.Errorf("ChangePassword err:%s", "email not found")
		resp.Error = Error(EmailNotFound, "")
		return
	}

	// rest code is the same can change
	if u.ResetCode == req.Code {
		u.Password = req.Password
		err = u.UpdatePassword()
		if err != nil {
			flog.Log.Errorf("ChangePassword err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	} else {
		flog.Log.Errorf("ChangePassword err:%s", "reset code wrong")
		resp.Error = Error(RestCodeWrong, "")
		return
	}

	// after change password, session will delete all
	DeleteUserAllSession(u.Id)
	resp.Flag = true
}

type UpdateUserRequest struct {
	NickName  string `json:"nick_name" validate:"omitempty"`
	Gender    int    `json:"gender" validate:"oneof=0 1 2"`
	Describe  string `json:"describe"`
	ImagePath string `json:"image_path"`
}

func UpdateUser(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// get oneself's info
	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateUser err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	uuu := new(model.User)
	uuu.Id = uu.Id
	ok, err := uuu.GetRaw()
	if err != nil {
		flog.Log.Errorf("UpdateUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("UpdateUser err: %s", "user not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	u := new(model.User)
	u.Id = uu.Id
	if req.ImagePath != "" && req.ImagePath != uuu.HeadPhoto {
		u.HeadPhoto = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			flog.Log.Errorf("UpdateUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			flog.Log.Errorf("UpdateUser err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}
	}

	// nickname can change 2 times one month
	if req.NickName != "" && req.NickName != uuu.NickName {
		u.NickName = req.NickName
	}

	u.Describe = req.Describe
	u.Gender = req.Gender
	err = u.UpdateInfo()
	if err != nil {
		flog.Log.Errorf("UpdateUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	err = session.FafaSessionMgr.RefreshUser([]int64{u.Id}, SessionExpireTime)
	if err != nil {
		flog.Log.Errorf("UpdateUser err:%s", err.Error())
		resp.Error = Error(RefreshUserCacheError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = u
}

type People struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	NickName     string `json:"nick_name"`
	Email        string `json:"email"`
	Gender       int    `json:"gender"`
	Describe     string `json:"describe"`
	HeadPhoto    string `json:"head_photo"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time,omitempty"`
	ActivateTime int64  `json:"activate_time,omitempty"`
	LoginTime    int64  `json:"login_time,omitempty"`
	LoginIp      string `json:"login_ip,omitempty"`
	IsInBlack    bool   `json:"is_in_black"`
	IsVip        bool   `json:"is_vip"`
}

// Take oneself's user info
func TakeUser(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	// just get from session
	u, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("TakeUser err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	user := new(model.User)
	user.Id = u.Id
	exist, err := model.Rdb.Client.Get(user)
	if err != nil {
		flog.Log.Errorf("TakeUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("TakeUser err:%s", "user  not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	v := user
	p := People{}
	p.Id = v.Id
	p.Describe = v.Describe
	p.CreateTime = v.CreateTime

	if v.Status == 2 {
		p.IsInBlack = true
	}

	p.UpdateTime = v.UpdateTime
	p.LoginTime = v.LoginTime
	p.LoginIp = v.LoginIp
	p.ActivateTime = v.ActivateTime

	p.Email = v.Email
	p.Name = v.Name
	p.NickName = v.NickName
	p.HeadPhoto = v.HeadPhoto
	p.Gender = v.Gender
	p.IsVip = v.Vip == 1
	resp.Flag = true
	resp.Data = p
}

type ListUserRequest struct {
	Id              int      `json:"id"`
	Name            string   `json:"name"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	Sort            []string `json:"sort"`
	Email           string   `json:"email" validate:"omitempty,email"`
	Gender          int      `json:"gender" validate:"oneof=-1 0 1 2"`
	Status          int      `json:"status" validate:"oneof=-1 0 1 2"`
	Vip             int      `json:"vip" validate:"oneof=-1 0 1"`
	PageHelp
}

type ListUserResponse struct {
	Users []model.User `json:"users"`
	PageHelp
}

// List user, admin url
func ListUser(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListUserResponse)
	req := new(ListUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ListUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.Rdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.User)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}
	if req.Name != "" {
		session.And("name=?", req.Name)
	}

	if req.Status != -1 {
		session.And("status=?", req.Status)
	}

	if req.Gender != -1 {
		session.And("gender=?", req.Gender)
	}

	if req.Vip != -1 {
		session.And("vip=?", req.Vip)
	}

	if req.Email != "" {
		session.And("email=?", req.Email)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	// count num
	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("ListUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if count>0 start list
	users := make([]model.User, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		// sql build
		p.build(session, req.Sort, model.UserSortName)
		// do query
		err = session.Find(&users)
		if err != nil {
			flog.Log.Errorf("ListUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	// result
	respResult.Users = users
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ListGroupUserRequest struct {
	GroupId int `json:"group_id" validate:"required"`
}

type ListGroupUserResponse struct {
	Users []model.User `json:"users"`
}

// List the users of group
func ListGroupUser(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListGroupUserResponse)
	req := new(ListGroupUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ListGroupUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.Rdb.Client.NewSession()
	defer session.Close()

	users := make([]model.User, 0)

	// group list where prepare
	err = session.Table(new(model.User)).Where("group_id=?", req.GroupId).Find(&users)
	if err != nil {
		flog.Log.Errorf("ListUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	respResult.Users = users
	resp.Data = respResult
	resp.Flag = true
}

type AssignGroupRequest struct {
	GroupId      int64   `json:"group_id"`
	GroupRelease int     `json:"group_release"`
	Users        []int64 `json:"users"`
}

// Assign user to a group, every user can only have less than one group
func AssignGroupToUser(c *gin.Context) {
	resp := new(Resp)
	req := new(AssignGroupRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if len(req.Users) == 0 {
		flog.Log.Errorf("AssignGroupToUser err:%s", "users empty")
		resp.Error = Error(ParasError, "users empty")
		return
	}

	// release the user of group, user will not belong to any group
	if req.GroupRelease == 1 {
		u := new(model.User)
		num, err := model.Rdb.Client.Cols("group_id").In("id", req.Users).Update(u)
		if err != nil {
			flog.Log.Errorf("AssignGroupToUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		err = session.FafaSessionMgr.RefreshUser(req.Users, SessionExpireTime)
		if err != nil {
			flog.Log.Errorf("AssignGroupToUser err:%s", err.Error())
			resp.Error = Error(RefreshUserCacheError, err.Error())
			return
		}
		resp.Data = num
	} else {
		if req.GroupId == 0 {
			flog.Log.Errorf("AssignGroupToUser err:%s", "group id empty")
			resp.Error = Error(ParasError, "group_id empty")
			return
		}

		g := new(model.Group)
		g.Id = req.GroupId
		exist, err := g.GetById()
		if err != nil {
			flog.Log.Errorf("AssignGroupToUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !exist {
			flog.Log.Errorf("AssignGroupToUser err:%s", "group not found")
			resp.Error = Error(GroupNotFound, "")
			return
		}

		u := new(model.User)
		u.GroupId = req.GroupId
		num, err := model.Rdb.Client.Cols("group_id").In("id", req.Users).Update(u)
		if err != nil {
			flog.Log.Errorf("AssignGroupToUser err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		err = session.FafaSessionMgr.RefreshUser(req.Users, SessionExpireTime)
		if err != nil {
			flog.Log.Errorf("AssignGroupToUser err:%s", err.Error())
			resp.Error = Error(RefreshUserCacheError, err.Error())
			return
		}
		resp.Data = num
	}

	resp.Flag = true
}

type UpdateUserAdminRequest struct {
	Id       int64  `json:"id" validate:"required"`
	NickName string `json:"nick_name" validate:"omitempty"`
	Password string `json:"password,omitempty"`
	Status   int    `json:"status" validate:"oneof=0 1 2"` // 0 nothing 1 activate 2 ban
	Vip      int    `json:"vip" validate:"oneof=0 1 2"`    // 1 become vip, 2 no vip
}

// Update user info, admin url. Can change user password, black one user, change nickname etc.
func UpdateUserAdmin(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateUserAdminRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateUserAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu := new(model.User)
	uu.Id = req.Id
	ok, err := uu.GetRaw()
	if err != nil {
		flog.Log.Errorf("UpdateUserAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("UpdateUserAdmin err: %s", "user not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	u := new(model.User)

	// admin can change nickname no more limit
	if req.NickName != "" && req.NickName != uu.NickName {
		u.NickName = req.NickName
	}
	u.Id = req.Id
	u.Password = req.Password

	// change user status, 1->2, 2->1
	u.Status = req.Status

	// vip change
	u.Vip = uu.Vip
	if req.Vip == 1 {
		u.Vip = 1
	} else if req.Vip == 2 {
		u.Vip = 0
	}
	err = u.UpdateInfoMustVip()
	if err != nil {
		flog.Log.Errorf("UpdateUserAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	err = session.FafaSessionMgr.RefreshUser([]int64{u.Id}, SessionExpireTime)
	if err != nil {
		flog.Log.Errorf("UpdateUserAdmin err:%s", err.Error())
		resp.Error = Error(RefreshUserCacheError, err.Error())
		return
	}
	resp.Data = u
	resp.Flag = true
}
