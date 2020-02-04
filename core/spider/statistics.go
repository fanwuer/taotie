package spider

import (
	"taotie/core/config"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

func IncrPoolCountToday(pool string, num1 int64) (num2 int64, today string, err error) {
	today = util.TodayStringByZone(3, TimeZone)
	num2, err = Pool.Client.IncrBy(pool+ValueSpilt+today, num1).Result()
	return
}

func GetPoolCountToday(pool string) (int64, string, error) {
	today := util.TodayStringByZone(3, TimeZone)
	i, _, err := Pool.Get(pool + ValueSpilt + today)
	if err != nil {
		return 0, today, err
	}

	count, err := util.SInt64(i)
	return count, today, err
}

func AwsCategoryStatistic() {
	t := time.NewTimer(AwsCategoryStatisticTickerSecond)
	for {
		if config.IsExpire {
			flog.Log.Errorf("software expire: %s", config.ExpireTime)
			return
		}
		select {
		case <-t.C:
			awsCategoryStatistic()
			t.Reset(AwsCategoryStatisticTickerSecond)
		}
	}

}

func awsCategoryStatistic() {
	count, today, err := GetPoolCountToday(AwsCategoryPoolStatisticsName)
	if err != nil {
		flog.Log.Errorf("ca get count today from redis err:%s", err.Error())
		return
	}
	if count > 0 {
		proxyPoolStatistics := new(model.AwsStatistics)
		proxyPoolStatistics.Today = today
		proxyPoolStatistics.Type = AwsCategoryPoolStatisticsType
		ok, err := proxyPoolStatistics.GetRaw()
		if err != nil {
			flog.Log.Errorf("ca get count today from mysql err:%s", err.Error())
			return
		}
		if ok {
			if proxyPoolStatistics.Times == count {
				return
			}
			proxyPoolStatistics.Times = count
			_, err = proxyPoolStatistics.UpdateCount()
			if err != nil {
				flog.Log.Errorf("ca update count today from mysql err:%s", err.Error())
			}
		} else {
			proxyPoolStatistics.Times = count
			proxyPoolStatistics.Name = AwsCategoryPoolStatisticsName
			_, err := proxyPoolStatistics.InsertOne()
			if err != nil {
				flog.Log.Errorf("ca insert count today from mysql err:%s", err.Error())
			}
		}
	}
}

func AwsAsinStatistic() {
	t := time.NewTimer(AwsAsinStatisticTickerSecond)
	for {
		if config.IsExpire {
			flog.Log.Errorf("software expire: %s", config.ExpireTime)
			return
		}
		select {
		case <-t.C:
			awsAsinStatistic()
			t.Reset(AwsAsinStatisticTickerSecond)
		}
	}

}

func awsAsinStatistic() {
	count, today, err := GetPoolCountToday(AwsAsinPoolStatisticsName)
	if err != nil {
		flog.Log.Errorf("asin get count today from redis err:%s", err.Error())
		return
	}
	if count > 0 {
		proxyPoolStatistics := new(model.AwsStatistics)
		proxyPoolStatistics.Today = today
		proxyPoolStatistics.Type = AwsAsinPoolStatisticsType
		ok, err := proxyPoolStatistics.GetRaw()
		if err != nil {
			flog.Log.Errorf("asin get count today from mysql err:%s", err.Error())
			return
		}
		if ok {
			if proxyPoolStatistics.Times == count {
				return
			}
			proxyPoolStatistics.Times = count
			_, err = proxyPoolStatistics.UpdateCount()
			if err != nil {
				flog.Log.Errorf("asin update count today from mysql err:%s", err.Error())
			}
		} else {
			proxyPoolStatistics.Times = count
			proxyPoolStatistics.Name = AwsAsinPoolStatisticsName
			_, err := proxyPoolStatistics.InsertOne()
			if err != nil {
				flog.Log.Errorf("asin insert count today from mysql err:%s", err.Error())
			}
		}
	}
}