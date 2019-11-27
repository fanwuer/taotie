package controllers

import (
	"github.com/gin-gonic/gin"
	"taotie/core/config"
	"time"
)

var TimeZone int64 = 0

func GetSecond2DateTimes(second int64) string {
	second = second + 3600*TimeZone
	tm := time.Unix(second, 0)
	return tm.UTC().Format("2006-01-02 15:04:05")

}

func Home(c *gin.Context) {
	resp := new(Resp)
	resp.Flag = true
	resp.Data = "TaoTie Version:" + config.Version
	defer func() {
		c.JSON(200, resp)
	}()
}
