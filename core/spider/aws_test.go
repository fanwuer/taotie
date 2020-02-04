package spider

import (
	"testing"
)

func TestStart(t *testing.T) {
	initTest()

	account := "fsdfef@qq.com"
	go ProxyPoolTickerStart(account)
}
