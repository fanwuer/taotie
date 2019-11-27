package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"io/ioutil"
	"runtime"
	"strings"
	. "taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

// Parse the json into request struct
func ParseJSON(c *gin.Context, req interface{}) *ErrorResp {
	pc, _, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	requestBody, _ := ioutil.ReadAll(c.Request.Body)

	ip := c.ClientIP()

	//Log.Debugf("%s ParseJSON [%v,line:%v]:%s", ip, f.Name(), line, string(requestBody))
	if err := json.Unmarshal(requestBody, req); err != nil {
		Log.Debugf("%s ParseJSONErr [%v,line:%v]:%s", ip, f.Name(), line, err.Error())
		// if parse wrong will not record log
		c.Set("skipLog", true)
		return Error(ParseJsonError, err.Error())
	}
	return nil
}

// Log the json output
func JSONL(c *gin.Context, code int, req interface{}, obj *Resp) {
	if c.GetBool("skipLog") {
		c.Render(code, render.JSON{Data: obj})
		return
	}

	// log will record
	record := new(model.Log)
	record.Ip = c.ClientIP()
	record.Url = c.Request.URL.Path
	record.LogTime = time.Now().Unix()
	record.Ua = c.Request.UserAgent()
	record.UserId = c.GetInt("uid")
	flag := obj.Flag
	if !flag && obj.Error != nil {
		errStr := obj.Error.Error()
		errStrSplit := strings.Split(errStr, "|")
		if len(errStrSplit) >= 2 {
			record.ErrorId = errStrSplit[0]
			record.ErrorMessage = strings.Join(errStrSplit[1:], "|")
		}
	}
	record.Flag = flag

	if req != nil {
		in, _ := json.Marshal(req)
		if len(in) > 0 {
			record.In = string(in)
		}
	}

	if obj != nil {
		out, _ := json.Marshal(obj)
		if len(out) > 0 {
			record.Out = string(out)
		}
	}
	cid := util.GetGUID()
	record.Cid = cid

	Log.Debugf("FaFa Monitor:%#v", record)

	// log table not read fot it will slow the service
	//_, err := model.FaFaRdb.InsertOne(record)
	//if err != nil {
	//	Log.Errorf("insert log record:%s", err.Error())
	//}

	obj.Cid = cid
	c.Render(code, render.JSON{Data: obj})
}

// Just render the json
func JSON(c *gin.Context, code int, obj *Resp) {
	c.Render(code, render.JSON{Data: obj})
}
