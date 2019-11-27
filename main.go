/*
	版权所有，侵权必究
	署名-非商业性使用-禁止演绎 4.0 国际
	警告： 以下的代码版权归属hunterhug，请不要传播或修改代码
	你可以在教育用途下使用该代码，但是禁止公司或个人用于商业用途(在未授权情况下不得用于盈利)
	商业授权请联系邮箱：gdccmcm14@live.com QQ:459527502

	All right reserved
	Attribution-NonCommercial-NoDerivatives 4.0 International
	Notice: The following code's copyright by hunterhug, Please do not spread and modify.
	You can use it for education only but can't make profits for any companies and individuals!
	For more information on commercial licensing please contact hunterhug.
	Ask for commercial licensing please contact Mail:gdccmcm14@live.com Or QQ:459527502

	2019.11 by hunterhug
*/
package main // import "taotie"

import (
	"flag"
	"fmt"
	"strings"
	"taotie/core/config"
	"taotie/core/controllers"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/router"
	"taotie/core/server"
	"taotie/core/session"
	"taotie/core/spider"
	"taotie/core/util"
	"taotie/core/util/mail"
	"time"
)

var (
	configFile  string
	createTable bool
	canSkipAuth bool
	role        string
	debug       bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&configFile, "f", "config.yaml", "Config file")
	flag.BoolVar(&createTable, "init_db", false, "Init create db table")
	flag.BoolVar(&canSkipAuth, "auth_skip_debug", false, "Auth skip debug")
	flag.StringVar(&role, "role", "", "Role")
	flag.Parse()
}

func initResource() (adminUrl map[string]int64) {
	adminUrl = make(map[string]int64)
	for url, handler := range router.V1Router {
		if !handler.Admin {
			continue
		}
		r := new(model.Resource)
		url1 := fmt.Sprintf("/v1%s", url)
		r.UrlHash, _ = util.Sha256([]byte(url1))
		r.Admin = true
		exist, err := r.GetRaw()
		if err != nil {
			panic(err)
		}

		if exist {
			adminUrl[url1] = r.Id
			continue
		} else {
			r := new(model.Resource)
			r.Url = url1
			r.UrlHash, _ = util.Sha256([]byte(url1))
			r.Name = handler.Name
			r.Describe = handler.Name
			r.Admin = handler.Admin
			r.CreateTime = time.Now().Unix()
			err := r.InsertOne()
			if err != nil {
				panic(err)
			}
			adminUrl[url1] = r.Id
		}
	}
	return adminUrl
}

func main() {
	var err error
	err = config.InitConfig(configFile)
	if err != nil {
		panic(err)
	}

	if role == "" {
		role = config.GlobalConfig.DefaultConfig.Role
	}

	switch role {
	case config.RoleAll, config.RoleWeb, config.RoleProxy, config.RoleAwsAsinTimer, config.RoleAwsCategoryTimer, config.RoleAwsAsinTask, config.RoleAwsCategoryTask:
	default:
		role = config.RoleWeb
	}

	logPath := ""
	temp := strings.Split(config.GlobalConfig.DefaultConfig.LogPath, ".")
	if len(temp) >= 2 {
		num := len(temp)
		logPath = fmt.Sprintf("%s_%s.%s", strings.Join(temp[:num-1], "."), role, temp[num-1])
	} else {
		logPath = fmt.Sprintf("%s_%s.%s", temp[0], role, "log")
	}

	if debug {
		flog.SetLogLevel("DEBUG")
	} else {
		flog.InitLog(logPath)
	}

	flog.Log.Debugf("Hi! Config is %#v", config.GlobalConfig)
	flog.Log.Noticef("\n%s-%s-v%s\n", config.Title, config.Version, role)

	defer func() {
		if err := recover(); err != nil {
			flog.Log.Errorf("Service internal err:%v", err)
		}
	}()

	err = model.InitRdb(config.GlobalConfig.DbConfig)
	if err != nil {
		panic(err)
	}

	go server.CheckExpire()

	spider.TimeZone = config.GlobalConfig.DefaultConfig.TimeZone
	if createTable {
		model.CreateTable([]interface{}{
			model.AwsStatistics{},
			model.AwsAsin{},
			model.AwsAsinLib{},
			model.AwsAsinTask{},
			model.AwsCategoryTask{},
		})
	}
	err = spider.InitRedisPool(config.GlobalConfig.KVConfig.RedisHost, config.GlobalConfig.KVConfig.RedisPass, config.GlobalConfig.KVConfig.RedisDB+1, config.GlobalConfig.KVConfig.RedisMaxIdle)
	if err != nil {
		panic(err)
	}
	spider.Debug = debug
	flog.Log.Debugf("Role:%s", role)
	switch role {
	case config.RoleProxy:
		spider.ProxyPoolTickerStart(config.GlobalConfig.SpiderConfig.Account)
	case config.RoleAwsAsinTimer:
		spider.AwsAsinTimerStart()
	case config.RoleAwsCategoryTimer:
		spider.AwsCategoryTimerStart()
	case config.RoleAwsAsinTask:
		spider.AwsAsinTaskStart(config.GlobalConfig.SpiderConfig.AwsAsinTaskThread)
	case config.RoleAwsCategoryTask:
		spider.AwsCategoryTaskStart(config.GlobalConfig.SpiderConfig.AwsCategoryTaskThread)
	case config.RoleWeb, config.RoleAll:
		controllers.AuthDebug = canSkipAuth
		controllers.TimeZone = config.GlobalConfig.DefaultConfig.TimeZone
		controllers.SingleLogin = config.GlobalConfig.DefaultConfig.SingleLogin
		controllers.SessionExpireTime = config.GlobalConfig.DefaultConfig.SessionExpireTime
		mail.Debug = !config.GlobalConfig.DefaultConfig.CanMail

		if createTable {
			model.CreateTable([]interface{}{
				model.User{},
				model.Group{},
				model.Resource{},
				model.GroupResource{},
				model.File{},
			})
		}

		err = session.InitSession(config.GlobalConfig.KVConfig)
		if err != nil {
			panic(err)
		}

		controllers.AdminUrl = initResource()
		engine := server.Server()
		engine.Static("/storage", config.GlobalConfig.DefaultConfig.StoragePath)
		engine.Static("/storage_x", config.GlobalConfig.DefaultConfig.StoragePath+"_x")

		router.SetRouter(engine)
		v1 := engine.Group("/v1")
		v1.Use(controllers.AuthFilter)
		router.SetAPIRouter(v1, router.V1Router)
		serverHost := fmt.Sprintf("%s:%d", config.GlobalConfig.DefaultConfig.Host, config.GlobalConfig.DefaultConfig.Port)

		if role == config.RoleAll {
			flog.Log.Debugf("spider start")
			go func() {
				go spider.ProxyPoolTickerStart(config.GlobalConfig.SpiderConfig.Account)
				go spider.AwsAsinTimerStart()
				go spider.AwsCategoryTimerStart()
				go spider.AwsAsinTaskStart(config.GlobalConfig.SpiderConfig.AwsAsinTaskThread)
				go spider.AwsCategoryTaskStart(config.GlobalConfig.SpiderConfig.AwsCategoryTaskThread)
			}()
		}
		err = engine.Run(serverHost)
		flog.Log.Noticef("Server run in %s", serverHost)
		if err != nil {
			flog.Log.Errorf("Server run err: %s", err.Error())
			return
		}
	}
}
