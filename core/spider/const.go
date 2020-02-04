package spider

import (
	"taotie/core/util"
	"time"
)

const (
	IPPoolName           = "ip_p"
	IPPoolStatisticsName = "ip_sta"

	AwsCategoryHashPoolToDoName  = "aws_ca_h_todo"
	AwsCategoryHashPoolDoingName = "aws_ca_h_doi"
	AwsCategoryHashPoolDoneName  = "aws_ca_h_done"

	AwsAsinHashPoolToDoName  = "aws_a_h_todo"
	AwsAsinHashPoolDoingName = "aws_a_h_doi"
	AwsAsinHashPoolDoneName  = "aws_a_h_done"

	AwsCategoryListPoolToDoName = "aws_ca_l_todo"
	AwsAsinListPoolToDoName     = "aws_a_l_todo"

	AwsCategoryPoolStatisticsName = "aws_c_sta"
	AwsAsinPoolStatisticsName     = "aws_a_sta"

	IPPoolStatisticsType          = 1
	AwsCategoryPoolStatisticsType = 2
	AwsAsinPoolStatisticsType     = 3
)

var (
	TimeZone            int64 = 8
	ValueSpilt                = "+"
	IPTickerTime              = 20 * time.Second
	IPFetchNum          int64 = 60
	IPDangerNum         int64 = 30
	IPExpireTime        int64 = 20 * 60
	MinerDefaultTimeOut       = 10
	DataPath                  = "data/asin"

	AwsProxyMaxErrTimes                            = 2
	AwsPoolExpireTime                              = 13 * time.Hour
	AwsCategoryPoolLoopTime                        = 12 * time.Hour
	AwsAsinPoolLoopTime                            = 12 * time.Hour
	AwsCategorySentToPoolRightNowMaxLiveTime       = 8 * time.Minute
	AwsAsinSentToPoolRightNowMaxLiveTime           = 8 * time.Minute
	AwsCategoryTickerSecond                        = 50 * time.Second
	AwsAsinTickerSecond                            = 50 * time.Second
	AwsCategoryStatisticTickerSecond               = 50 * time.Second
	AwsAsinStatisticTickerSecond                   = 50 * time.Second
	AwsListCategoryTaskLimit                 int64 = 100
	AwsListAsinTaskLimit                     int64 = 100
)

func init() {
	util.MakeDir(DataPath)
}
