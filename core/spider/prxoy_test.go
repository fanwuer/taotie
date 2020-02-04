package spider

import (
	"fmt"
	"os"
	"taotie/core/config"
	"taotie/core/model"
	"testing"
)

func initTest() {
	err := config.InitConfig("/Users/zhujiang/Documents/jinhan/taotie/config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	model.InitRdb(config.GlobalConfig.DbConfig)
	InitRedisPool(config.GlobalConfig.KVConfig.RedisHost, config.GlobalConfig.KVConfig.RedisPass, config.GlobalConfig.KVConfig.RedisDB+1, config.GlobalConfig.KVConfig.RedisMaxIdle)
}

func TestProxyPoolStart(t *testing.T) {
	initTest()
	account := "fsdfef@qq.com"
	ProxyPoolTickerStart(account)
}

func TestProxyPoolNum(t *testing.T) {
	initTest()
	num, err := GetProxyPoolNum()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(num)
}

func TestGetIPFromPool(t *testing.T) {
	initTest()
	s, err := GetIPFromPool()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(s)
}

func TestAddProxyPoolNumToday(t *testing.T) {
	initTest()
	s, today, err := IncrPoolCountToday(IPPoolStatisticsName, 250)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(s, today)

	c, today, err := GetPoolCountToday(IPPoolStatisticsName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(c, today)
}
