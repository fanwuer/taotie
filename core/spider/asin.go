package spider

import (
	"errors"
	"taotie/core/config"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

var AwsAsinTimerHasRun = false

func AwsAsinTimerStart() {
	if AwsAsinTimerHasRun {
		return
	} else {
		AwsAsinTimerHasRun = true
	}

	go AwsAsinStatistic()
	AwsAsinSentToPool()
}

func AwsAsinSentToPool() {
	awsAsinSentToPool()
	t := time.NewTimer(AwsAsinTickerSecond)
	for {
		if config.IsExpire {
			flog.Log.Errorf("software expire: %s", config.ExpireTime)
			return
		}
		select {
		case <-t.C:
			awsAsinSentToPool()
			t.Reset(AwsAsinTickerSecond)
		}
	}
}

func awsAsinSentToPool() {
	thisTime := time.Now().Add(-AwsPoolExpireTime).Unix()
	num, err := model.Rdb.Client.Where("open=?", 1).And("status=?", 0).And("last_catch_time<?", thisTime).Count(new(model.AwsAsinTask))
	if err != nil {
		flog.Log.Errorf("asin mysql count err:%s", err.Error())
		return
	}

	if num == 0 {
		return
	}

	page := num / AwsListAsinTaskLimit
	isAddOne := (num % AwsListAsinTaskLimit) > 0
	if isAddOne {
		page = page + 1
	}

	var i int64 = 0
	for ; i < page; i++ {
		tasks := make([]model.AwsAsinTask, 0)
		thisTime = time.Now().Add(-AwsPoolExpireTime).Unix()
		err = model.Rdb.Client.Cols("id", "asin").Where("open=?", 1).And("status=?", 0).And("last_catch_time<?", thisTime).Limit(int(AwsListAsinTaskLimit), int(i*AwsListAsinTaskLimit)).Find(&tasks)
		if err != nil {
			flog.Log.Errorf("asin mysql select err:%s", err.Error())
			return
		}

		if len(tasks) == 0 {
			return
		}

		for _, task := range tasks {
			//flog.Log.Debugf("asin mysql select:%v", task)
			if awsOneAsinSentToPool(task.Asin, 0, 0) != nil {
				return
			}
		}
	}
}

func awsOneAsinSentToPool(asin string, categoryId int64, taskType int64) error {
	now := time.Now().Unix()
	key := asin
	value := util.JoinSpilt([]interface{}{categoryId, taskType, asin, now}, ValueSpilt)

	exist, can, lastTime, err := awsPoolKeyCanLive(AwsAsinHashPoolDoneName, key, AwsAsinPoolLoopTime)
	if err != nil {
		flog.Log.Errorf("asin done hash find err:%s", err.Error())
		return err
	}
	if exist && can {
		flog.Log.Debugf("asin:%s loop is done,pass", asin)
		t := new(model.AwsAsinTask)
		t.LastCatchTime = lastTime
		_, err = model.Rdb.Client.Where("asin=?", key).Cols("last_catch_time").Update(t)
		return err
	}

	_, exist, err = GetHashPool(AwsAsinHashPoolDoingName, key)
	if err != nil {
		flog.Log.Errorf("asin doing hash find err:%s", err.Error())
		return err
	}
	if exist {
		flog.Log.Debugf("asin:%s loop is doing,pass", asin)
		return nil
	}

	_, exist, err = GetHashPool(AwsAsinHashPoolToDoName, key)
	if err != nil {
		flog.Log.Errorf("asin todo hash find err:%s", err.Error())
		return err
	}
	if exist {
		flog.Log.Debugf("asin:%s is todo,pass", asin)
		return nil
	}

	err = PutHashPool(AwsAsinHashPoolToDoName, key, now)
	if err != nil {
		flog.Log.Errorf("asin todo hash put err:%s", err.Error())
		return err
	}

	err = RPushListPool(AwsAsinListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("asin todo list put err:%s", err.Error())
		return err
	}

	flog.Log.Debugf("asin pool add:%s", asin)
	return nil
}

func AwsOneAsinSentToPoolRightNow(asin string) error {
	now := time.Now().Unix()
	key := asin
	value := util.JoinSpilt([]interface{}{0, 0, asin, now}, ValueSpilt)

	exist, can, lastTime, err := awsPoolKeyCanLive(AwsAsinHashPoolDoneName, key, AwsAsinSentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("asin done hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("asin:%s right now is done,pass", asin)
		t := new(model.AwsAsinTask)
		t.LastCatchTime = lastTime
		_, err = model.Rdb.Client.Where("asin=?", key).Cols("last_catch_time").Update(t)
		if err != nil {
			return err
		}

		return errors.New("busy now")
	}

	exist, can, _, err = awsPoolKeyCanLive(AwsAsinHashPoolDoingName, key, AwsAsinSentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("asin doing hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("asin:%s right now is doing,pass", asin)
		return errors.New("busy now")
	}

	exist, can, _, err = awsPoolKeyCanLive(AwsAsinHashPoolToDoName, key, AwsAsinSentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("asin todo hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("asin:%s right now is todo,pass", asin)
		return errors.New("busy now")
	}

	err = PutHashPool(AwsAsinHashPoolToDoName, key, now)
	if err != nil {
		flog.Log.Errorf("asin todo hash right now put err:%s", err.Error())
		return err
	}

	err = LRemListPool(AwsAsinListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("asin todo list right now rm err:%s", err.Error())
		return err
	}

	err = LPushListPool(AwsAsinListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("asin todo list right now put err:%s", err.Error())
		return err
	}

	flog.Log.Debugf("asin pool add:%s", asin)
	return nil
}
