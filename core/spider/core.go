package spider

import (
	"fmt"
	"io/ioutil"
	"os"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

var (
	AwsAsinTaskHasRun     = false
	AwsCategoryTaskHasRun = false
	Debug                 = false
)

func AwsAsinTaskStart(thread int64) {
	if AwsAsinTaskHasRun {
		return
	}

	AwsAsinTaskHasRun = true
	end := make(chan string, 0)

	for i := 0; i < int(thread); i++ {
		go awsAsinTaskStart("thread_" + util.IS(i))
	}

	<-end
}

func AwsCategoryTaskStart(thread int64) {
	if AwsCategoryTaskHasRun {
		return
	}

	AwsCategoryTaskHasRun = true

	end := make(chan string, 0)

	for i := 0; i < int(thread); i++ {
		go awsCategoryTaskStart("thread_" + util.IS(i))
	}

	<-end
}

func awsCategoryTaskStart(taskName string) {
	ip := getIp()

	for {
		v, err := BLPopListPool(AwsCategoryListPoolToDoName)
		if err != nil {
			flog.Log.Errorf("ca %s pop todo err:%s", taskName, err.Error())
			os.Exit(1)
		}
		now := time.Now().Unix()
		vv := util.SpiltJoin(v[1], ValueSpilt)
		link := vv[0]
		categoryIdStr := vv[1]

		pageNum, _ := util.SInt64(vv[2])

		if pageNum > 5 || pageNum < 0 {
			pageNum = 5
		}

		taskType, _ := util.SInt64(vv[3])
		catchDetail, _ := util.SInt64(vv[4])
		genTime, _ := util.SInt64(vv[5])

		if now-genTime > int64(AwsCategoryPoolLoopTime.Seconds()) {
			flog.Log.Debugf("ca %s deal %s expire", taskName, link)
			continue
		} else {
			flog.Log.Debugf("ca %s deal %s,page:%d,type:%d.detail:%d", taskName, link, pageNum, taskType, catchDetail)
		}

		err = DeleteHashPool(AwsCategoryHashPoolToDoName, categoryIdStr)
		if err != nil {
			flog.Log.Errorf("ca %s rm todo redis err:%s", taskName, err.Error())
		}
		err = PutHashPool(AwsCategoryHashPoolDoingName, categoryIdStr, time.Now().Unix())
		if err != nil {
			flog.Log.Errorf("ca %s put doing redis err:%s", taskName, err.Error())
		}

		DownloadCategory(taskName, categoryIdStr, ip, link, pageNum, taskType, catchDetail)

		err = DeleteHashPool(AwsCategoryHashPoolDoingName, categoryIdStr)
		if err != nil {
			flog.Log.Errorf("ca %s rm doing redis err:%s", taskName, err.Error())
		}
		err = PutHashPool(AwsCategoryHashPoolDoneName, categoryIdStr, time.Now().Unix())
		if err != nil {
			flog.Log.Errorf("ca %s put done redis err:%s", taskName, err.Error())
		}
	}
}

func DownloadCategory(taskName, categoryIdStr, ip, link string, pageNum, taskType, catchDetail int64) {
	if ip == "" {
		ip = getIp()
	}

	if taskType != 1 && taskType != 2 {
		return
	}

	if taskType == 2 {
		pageNum = 50
	}

	var i int64 = 1
	for ; i <= pageNum; i++ {
		url := fmt.Sprintf("%s?_encoding=UTF8&pg=%d&ajax=1", link, i)
		if taskType == 2 {
			url = fmt.Sprintf("%s&page=%d", link, i)
		}

		for {
			content, err := Download(ip, url)
			ipSpider, ok := Spiders.Get(ip)
			if err != nil {
				flog.Log.Errorf("ca %s:%s err:%s", taskName, url, err.Error())
				if ok && ipSpider.Errortimes > AwsProxyMaxErrTimes {
					Spiders.Delete(ip)
					ip = getIp()
				}
				continue
			} else {
				if Is404(content) {
					return
				}

				if s := IsRobot(content); s != "" {
					flog.Log.Errorf("ca %s:%s err:%s", taskName, url, s)
					Spiders.Delete(ip)
					ip = getIp()
					continue
				}

				err := TooSortSizes(content, 3)
				if err != nil {
					return
				}

				returnMap, err := ParseList(content, taskType == 2)
				if err != nil {
					return
				}

				asins := make([]*model.AwsAsin, 0)
				for _, m := range returnMap {
					asin := new(model.AwsAsin)
					asin.Asin = m["asin"]
					asin.Img = m["img"]
					if m["is_prime"] == "true" {
						asin.IsPrime = 1
					}

					asin.Price, _ = util.SFloat64(m["price"])
					asin.Reviews, _ = util.SInt64(m["reviews"])
					asin.SmallRank, _ = util.SInt64(m["small_rank"])
					asin.Score, _ = util.SFloat64(m["score"])
					asin.Name = m["title"]
					asin.CategoryTaskId, _ = util.SInt64(categoryIdStr)
					asin.CategoryTaskType = taskType
					asin.CreateTime = time.Now().Unix()
					if catchDetail == 1 {
						awsOneAsinSentToPool(asin.Asin, asin.CategoryTaskId, taskType)
					}

					asins = append(asins, asin)
				}

				IncrPoolCountToday(AwsCategoryPoolStatisticsName, 1)
				_, err = model.Rdb.Client.Insert(asins)
				if err != nil {
					flog.Log.Errorf("ca %s:%s err:%s", taskName, url, err.Error())
				}
				break
			}
		}
	}
}

func awsAsinTaskStart(taskName string) {
	ip := getIp()

	for {
		v, err := BLPopListPool(AwsAsinListPoolToDoName)
		if err != nil {
			flog.Log.Errorf("asin %s pop todo err:%s", taskName, err.Error())
			os.Exit(1)
		}
		now := time.Now().Unix()
		vv := util.SpiltJoin(v[1], ValueSpilt)
		categoryIdStr := vv[0]
		categoryType := vv[1]
		asin := vv[2]
		genTime, _ := util.SInt64(vv[3])

		if now-genTime > int64(AwsAsinPoolLoopTime.Seconds()) {
			flog.Log.Debugf("asin %s deal %s expire", taskName, asin)
			continue
		} else {
			flog.Log.Debugf("asin %s deal %s,cid:%v-%v", taskName, asin, categoryIdStr, categoryType)
		}

		err = DeleteHashPool(AwsAsinHashPoolToDoName, asin)
		if err != nil {
			flog.Log.Errorf("asin %s rm todo redis err:%s", taskName, err.Error())
		}
		err = PutHashPool(AwsAsinHashPoolDoingName, asin, time.Now().Unix())
		if err != nil {
			flog.Log.Errorf("asin %s put doing redis err:%s", taskName, err.Error())
		}

		DownloadAsin(taskName, categoryIdStr, ip, asin, categoryType)

		err = DeleteHashPool(AwsAsinHashPoolDoingName, asin)
		if err != nil {
			flog.Log.Errorf("asin %s rm doing redis err:%s", taskName, err.Error())
		}
		err = PutHashPool(AwsAsinHashPoolDoneName, asin, time.Now().Unix())
		if err != nil {
			flog.Log.Errorf("asin %s put done redis err:%s", taskName, err.Error())
		}
	}
}

func DownloadAsin(taskName, categoryIdStr, ip, asin, taskType string) {
	if ip == "" {
		ip = getIp()
	}

	if asin == "" {
		return
	}

	url := fmt.Sprintf("https://www.amazon.com/dp/%s", asin)

	for {
		content, err := Download(ip, url)
		ipSpider, ok := Spiders.Get(ip)
		if err != nil {
			flog.Log.Errorf("asin %s:%s err:%s", taskName, url, err.Error())
			if ok && ipSpider.Errortimes > AwsProxyMaxErrTimes {
				Spiders.Delete(ip)
				ip = getIp()
			}
			continue
		} else {
			if Is404(content) {
				return
			}

			if s := IsRobot(content); s != "" {
				flog.Log.Errorf("asin %s:%s err:%s", taskName, url, s)
				Spiders.Delete(ip)
				ip = getIp()
				continue
			}

			if Debug {
				ioutil.WriteFile(fmt.Sprintf("%s/%s.html", DataPath, asin), content, 0777)
			}

			m, err := ParseDetail(content)
			if err != nil {
				flog.Log.Errorf("asin %s:%s err:%s", taskName, url, err.Error())
				return
			} else {
			}

			lib := new(model.AwsAsinLib)
			lib.Asin = asin
			ok, err := lib.GetRaw()
			if err != nil {
				flog.Log.Errorf("asin %s:%s err:%s", taskName, url, err.Error())
			} else {
				if ok {
					lib.Incr()
				} else {
					lib.Times = 1
					if m.BigName != "" {
						lib.Tag = m.BigName
					}
					lib.InsertOne()
				}
			}

			asinDetail := new(model.AwsAsin)
			asinDetail.Asin = asin
			asinDetail.Img = m.Img
			if m.IsPrime {
				asinDetail.IsPrime = 1
			}

			asinDetail.Price = m.Price
			asinDetail.Reviews = m.Reviews
			asinDetail.Score = m.Score
			asinDetail.Name = m.Title
			asinDetail.CategoryTaskId, _ = util.SInt64(categoryIdStr)
			asinDetail.CategoryTaskType, _ = util.SInt64(taskType)
			asinDetail.CreateTime = time.Now().Unix()
			asinDetail.Describe = m.Describe
			asinDetail.RankDetail = m.RankDetail
			if m.IsAwsSold {
				asinDetail.IsAwsSold = 1
			}
			if m.IsFba {
				asinDetail.IsFba = 1
			}

			if !m.IsStock {
				asinDetail.Remark = "not in stock"
			}
			asinDetail.BigRank = m.BigRank
			asinDetail.IsDetail = 1
			asinDetail.Tag = m.BigName
			asinDetail.SoldBy = m.SoldBy
			asinDetail.SoldById = m.SoldById
			IncrPoolCountToday(AwsAsinPoolStatisticsName, 1)

			_, err = model.Rdb.Client.Insert(asinDetail)
			if err != nil {
				flog.Log.Errorf("asin %s:%s err:%s", taskName, url, err.Error())
			}
			break

		}
	}
}

func getIp() string {
	ipStr, err := GetIPFromPool()
	if err != nil {
		flog.Log.Errorf("get proxy:%v", err.Error())
		os.Exit(1)
	}

	v := util.SpiltJoin(ipStr[1], ValueSpilt)
	ipTime, _ := util.SInt64(v[1])
	if time.Now().Unix()-ipTime > IPExpireTime {
		flog.Log.Debugf("get proxy:%v diu", v)
		return getIp()
	} else {
		flog.Log.Debugf("get proxy:%v", v)
	}

	ip := v[0]
	return ip
}
