package spider

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestAwsCategoryTaskStart(t *testing.T) {
	initTest()
	account := "fsdfef@qq.com"
	go ProxyPoolTickerStart(account)
	go AwsCategoryTimerStart()
	go AwsAsinTimerStart()
	AwsCategoryTaskStart(1)
}

func TestDownloadCategory(t *testing.T) {
	initTest()
	account := "fsdfef@qq.com"
	go ProxyPoolTickerStart(account)
	taskName, categoryIdStr := "", "1"
	ip, link := "", "https://www.amazon.com/Best-Sellers-Appliances-Clothes-Washing-Machines/zgbs/appliances/13397491"
	var pageNum int64 = 5
	var taskType int64 = 1
	var catchDetail int64 = 1
	DownloadCategory(taskName, categoryIdStr, ip, link, pageNum, taskType, catchDetail)
}

func TestDownloadCategory1(t *testing.T) {
	initTest()
	account := "fsdfef@qq.com"
	go ProxyPoolTickerStart(account)
	taskName, categoryIdStr := "", "2"
	ip, link := "", "https://www.amazon.com/s?me=A3K0O9OTHV8P5O"
	var pageNum int64 = 5
	var taskType int64 = 2
	var catchDetail int64 = 1
	DownloadCategory(taskName, categoryIdStr, ip, link, pageNum, taskType, catchDetail)
}

func TestDownload(t *testing.T) {
	initTest()
	link := "https://www.amazon.com/s?me=A3K0O9OTHV8P5O&page=11"
	raw, err := Download(getIp(), link)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ioutil.WriteFile("./11.html", raw, 0777)
}

func TestAwsAsinTaskStart(t *testing.T) {
	initTest()
	account := "fsdfef@qq.com"
	go ProxyPoolTickerStart(account)
	//B00FMWWN6U
	//B07GCGSZG8
	//B075JKRBL2
	// B07VY4KZRZ
	link := "https://www.amazon.com/dp/B07VY4KZRZ"

	for {
		raw, err := Download(getIp(), link)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ioutil.WriteFile("./B07VY4KZRZ.html", raw, 0777)
		return
	}
}
