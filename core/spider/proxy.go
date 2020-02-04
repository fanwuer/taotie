package spider

import (
	"encoding/json"
	"fmt"
	"github.com/hunterhug/marmot/miner"
	"taotie/core/config"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

type mvpResult struct {
	Code          interface{}       `json:"code"`
	CodeMsg       string            `json:"code_msg"`
	TodayFetchNum interface{}       `json:"today_fetch_num"`
	TodayTotalNum interface{}       `json:"today_total_num"`
	CostTime      string            `json:"cost_time"`
	ResultCount   int64             `json:"result_count"`
	ResponseTime  string            `json:"dtime"`
	Result        []mvpResultDetail `json:"result"`
}

type mvpResultDetail struct {
	Ip           string `json:"ip:port"`   //"ip:port": "67.207.95.138:8080",
	Type         string `json:"http_type"` //"http_type": "HTTPS",
	An           string `json:"anonymous"` //"anonymous": "高匿",
	Isp          string `json:"isp"`       //"isp": "null",
	Country      string `json:"country"`   //"country": "美国"
	TransferTime int64  `json:"transfer_time"`
	PingTime     int64  `json:"ping_time"`
}

var ProxyHasRun = false

func ProxyPoolTickerStart(account string) {
	if ProxyHasRun {
		return
	} else {
		ProxyHasRun = true
	}
	url := "http://proxy.mimvp.com/api/fetch.php?orderid=%s&num=%d&result_format=json&anonymous=5&result_fields=1,2,3,4,5&http_type=1,2,5&ping_time=5&transfer_time=5"
	worker := miner.NewAPI()
	worker.Url = fmt.Sprintf(url, account, IPFetchNum)

	ProxyPoolAction(worker)

	for {
		if config.IsExpire {
			flog.Log.Errorf("software expire: %s", config.ExpireTime)
			return
		}
		select {
		case <-time.After(IPTickerTime):
			ProxyPoolAction(worker)
		}
	}
}

func ProxyPoolAction(worker *miner.Worker) {
	num, err := GetProxyPoolNum()
	if err != nil {
		flog.Log.Errorf("proxy count from pool err:%s", err.Error())
		return
	}
	flog.Log.Debugf("proxy count pool num:%d", num)
	if num > IPDangerNum {
		return
	}
	data, err := worker.Get()
	if err != nil {
		flog.Log.Errorf("proxy get err:%s", err.Error())
		return
	}
	r := new(mvpResult)
	err = json.Unmarshal(data, r)
	if err != nil {
		flog.Log.Errorf("proxy parse err:%s", err.Error())
		return
	}
	if fmt.Sprintf("%v", r.Code) != "0" {
		flog.Log.Errorf("proxy api wrong:%s", r.CodeMsg)
		return
	}

	flog.Log.Debugf("proxy get ip %s, cal:%v-%v-%v, cost:%s", r.CodeMsg, r.ResultCount, r.TodayFetchNum, r.TodayTotalNum, r.CostTime)

	IPs := make([]interface{}, 0)
	for _, v := range r.Result {
		flog.Log.Debugf("proxy ip:%v", v)
		if v.Type == "Socks5" {
			v.Ip = "socks5://" + v.Ip
		} else if v.Type == "HTTPS" {
			v.Ip = "https://" + v.Ip
		} else {
			v.Ip = "http://" + v.Ip
		}

		v.Ip = util.JoinSpilt([]interface{}{v.Ip, time.Now().Unix(), v.Country}, ValueSpilt)
		IPs = append(IPs, v.Ip)
	}

	num, err = PutProxyPool(IPs)
	if err != nil {
		flog.Log.Errorf("proxy put pool err:%s", err.Error())
		return
	}

	flog.Log.Debugf("proxy put pool num:%d", num)

	count, today, err := IncrPoolCountToday(IPPoolStatisticsName, num)
	if err != nil {
		flog.Log.Errorf("proxy incr pool err:%s", err.Error())
		return
	}

	if count > 0 {
		proxyPoolStatistics := new(model.AwsStatistics)
		proxyPoolStatistics.Today = today
		proxyPoolStatistics.Type = IPPoolStatisticsType
		ok, err := proxyPoolStatistics.GetRaw()
		if err != nil {
			flog.Log.Errorf("proxy get count today from mysql err:%s", err.Error())
			return
		}
		if ok {
			if proxyPoolStatistics.Times == count {
				return
			}
			proxyPoolStatistics.Times = count
			_, err = proxyPoolStatistics.UpdateCount()
			if err != nil {
				flog.Log.Errorf("proxy update count today from mysql err:%s", err.Error())
			}
		} else {
			proxyPoolStatistics.Times = count
			proxyPoolStatistics.Name = IPPoolStatisticsName
			_, err := proxyPoolStatistics.InsertOne()
			if err != nil {
				flog.Log.Errorf("proxy insert count today from mysql err:%s", err.Error())
			}
		}
	}

}

func GetProxyPoolNum() (num int64, err error) {
	return Pool.LLen(IPPoolName)
}

func PutProxyPool(IPs []interface{}) (num int64, err error) {
	return Pool.RPush(IPPoolName, IPs...)
}

func GetIPFromPool() ([]string, error) {
	return Pool.BLPop(0, IPPoolName)
}
