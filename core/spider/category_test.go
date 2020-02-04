package spider

import "testing"

func TestAwsCategoryTimerStart(t *testing.T) {
	initTest()
	go AwsCategoryTimerStart()
	AwsAsinTimerStart()
}
