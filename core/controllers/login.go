package controllers

import (
	"github.com/gin-gonic/gin"
	"strings"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/session"
	"time"
)

type LoginRequest struct {
	UserName string `json:"user_name"`
	PassWd   string `json:"pass_wd"`
}

func Login(c *gin.Context) {
	resp := new(Resp)
	req := new(LoginRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	// check session
	//userInfo, _ := GetUserSession(c)
	//if userInfo != nil {
	//	//c.Set("skipLog", true)
	//	c.Set("uid", userInfo.Id)
	//	resp.Flag = true
	//	return
	//}

	// paras not empty
	if req.UserName == "" || req.PassWd == "" {
		flog.Log.Errorf("login err:%s", "paras wrong")
		resp.Error = Error(ParasError, "field username or pass_wd")
		return
	}

	// common people login
	uu := new(model.User)
	if strings.Contains(req.UserName, "@") {
		uu.Email = req.UserName
	} else {
		uu.Name = req.UserName
	}
	uu.Password = req.PassWd
	ok, err := uu.GetRaw()
	if err != nil {
		flog.Log.Errorf("login err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("login err:%s", "user or password wrong")
		resp.Error = Error(LoginWrong, "user or password wrong")
		return
	}

	c.Set("uid", uu.Id)

	u := new(model.User)
	u.Id = uu.Id
	u.LoginIp = c.ClientIP()
	u.LoginTime = time.Now().Unix()

	// Update the login ip into db
	u.UpdateLoginInfo()

	// Refresh the user info in session(redis)
	session.FafaSessionMgr.RefreshUser([]int64{u.Id}, SessionExpireTime)

	// Activate or black user can login, but those auth api can not use
	token, err := SetUserSession(uu)
	if err != nil {
		flog.Log.Errorf("login err:%s", err.Error())
		resp.Error = Error(SetUserSessionError, err.Error())
		return
	}

	resp.Data = token
	resp.Flag = true
}

func Logout(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSON(c, 200, resp)
	}()
	user, err := GetUserSession(c)

	if err != nil {
		flog.Log.Errorf("logout err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	if user != nil {
		err = DeleteUserSession(c)
		if err != nil {
			flog.Log.Errorf("logout err:%s", err.Error())
			resp.Error = Error(DeleteUserSessionError, err.Error())
			return
		}
	}
	resp.Flag = true
}

func Refresh(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSON(c, 200, resp)
	}()
	user, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("refresh err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	if user != nil {
		err = RefreshUserSession(c)
		if err != nil {
			flog.Log.Errorf("refresh err:%s", err.Error())
			resp.Error = Error(RefreshUserCacheError, err.Error())
			return
		}
	}
	resp.Flag = true
}
