package spider

import (
	"errors"
	"fmt"
	"taotie/core/config"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/util"
	"time"
)

var AwsCategoryTimerHasRun = false

func AwsCategoryTimerStart() {
	if AwsCategoryTimerHasRun {
		return
	} else {
		AwsCategoryTimerHasRun = true
	}

	go AwsCategoryStatistic()
	AwsCategorySentToPool()
}

func AwsCategorySentToPool() {
	awsCategorySentToPool()
	t := time.NewTimer(AwsCategoryTickerSecond)
	for {
		if config.IsExpire {
			flog.Log.Errorf("software expire: %s", config.ExpireTime)
			return
		}
		select {
		case <-t.C:
			awsCategorySentToPool()
			t.Reset(AwsCategoryTickerSecond)
		}
	}
}

func awsCategorySentToPool() {
	thisTime := time.Now().Add(-AwsPoolExpireTime).Unix()
	num, err := model.Rdb.Client.Where("open=?", 1).And("status=?", 0).And("last_catch_time<?", thisTime).Count(new(model.AwsCategoryTask))
	if err != nil {
		flog.Log.Errorf("ca mysql count err:%s", err.Error())
		return
	}

	if num == 0 {
		return
	}

	page := num / AwsListCategoryTaskLimit
	isAddOne := (num % AwsListCategoryTaskLimit) > 0
	if isAddOne {
		page = page + 1
	}

	var i int64 = 0
	for ; i < page; i++ {
		tasks := make([]model.AwsCategoryTask, 0)
		thisTime = time.Now().Add(-AwsPoolExpireTime).Unix()
		err = model.Rdb.Client.Cols("id", "page_num", "link", "link_hash_code", "type", "catch_detail").Where("open=?", 1).And("status=?", 0).And("last_catch_time<?", thisTime).Limit(int(AwsListCategoryTaskLimit), int(i*AwsListCategoryTaskLimit)).Find(&tasks)
		if err != nil {
			flog.Log.Errorf("ca mysql select err:%s", err.Error())
			return
		}

		if len(tasks) == 0 {
			return
		}

		for _, task := range tasks {
			//flog.Log.Debugf("ca mysql select:%v", task)
			if awsOneCategorySentToPool(task) != nil {
				return
			}
		}
	}
}

func awsOneCategorySentToPool(task model.AwsCategoryTask) error {
	now := time.Now().Unix()
	key := fmt.Sprintf("%d", task.Id)
	value := util.JoinSpilt([]interface{}{task.Link, task.Id, task.PageNum, task.Type, task.CatchDetail, now}, ValueSpilt)

	exist, can, lastTime, err := awsPoolKeyCanLive(AwsCategoryHashPoolDoneName, key, AwsCategoryPoolLoopTime)
	if err != nil {
		flog.Log.Errorf("ca done hash find err:%s", err.Error())
		return err
	}
	if exist && can {
		flog.Log.Debugf("ca hash:%s loop is done,pass", task.Link)
		t := new(model.AwsCategoryTask)
		t.LastCatchTime = lastTime
		_, err = model.Rdb.Client.ID(task.Id).Cols("last_catch_time").Update(t)
		return err
	}

	_, exist, err = GetHashPool(AwsCategoryHashPoolDoingName, key)
	if err != nil {
		flog.Log.Errorf("ca doing hash find err:%s", err.Error())
		return err
	}
	if exist {
		flog.Log.Debugf("ca hash:%s loop is doing,pass", task.Link)
		return nil
	}

	_, exist, err = GetHashPool(AwsCategoryHashPoolToDoName, key)
	if err != nil {
		flog.Log.Errorf("ca todo hash find err:%s", err.Error())
		return err
	}
	if exist {
		flog.Log.Debugf("ca hash:%s is todo,pass", task.Link)
		return nil
	}

	err = PutHashPool(AwsCategoryHashPoolToDoName, key, now)
	if err != nil {
		flog.Log.Errorf("ca todo hash put err:%s", err.Error())
		return err
	}

	err = RPushListPool(AwsCategoryListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("ca todo list put err:%s", err.Error())
		return err
	}

	flog.Log.Debugf("ca pool add:%s", task.Link)
	return nil
}

func AwsOneCategorySentToPoolRightNow(task model.AwsCategoryTask) error {
	now := time.Now().Unix()
	key := fmt.Sprintf("%d", task.Id)
	value := util.JoinSpilt([]interface{}{task.Link, task.Id, task.PageNum, task.Type, task.CatchDetail, now}, ValueSpilt)

	exist, can, lastTime, err := awsPoolKeyCanLive(AwsCategoryHashPoolDoneName, key, AwsCategorySentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("ca done hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("ca hash:%s right now is done,pass", task.Link)
		t := new(model.AwsCategoryTask)
		t.LastCatchTime = lastTime
		_, err = model.Rdb.Client.Where("link_hash_code=?", key).Cols("last_catch_time").Update(t)
		if err != nil {
			return err
		}

		return errors.New("busy now")
	}

	exist, can, _, err = awsPoolKeyCanLive(AwsCategoryHashPoolDoingName, key, AwsCategorySentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("ca doing hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("ca hash:%s right now is doing,pass", task.Link)
		return errors.New("busy now")
	}

	exist, can, _, err = awsPoolKeyCanLive(AwsCategoryHashPoolToDoName, key, AwsCategorySentToPoolRightNowMaxLiveTime)
	if err != nil {
		flog.Log.Errorf("ca todo hash right now find err:%s", err.Error())
		return err
	}

	if exist && can {
		flog.Log.Debugf("ca hash:%s right now is todo,pass", task.Link)
		return errors.New("busy now")
	}

	err = PutHashPool(AwsCategoryHashPoolToDoName, key, now)
	if err != nil {
		flog.Log.Errorf("ca todo hash right now put err:%s", err.Error())
		return err
	}

	err = LRemListPool(AwsCategoryListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("ca todo list right now rm err:%s", err.Error())
		return err
	}

	err = LPushListPool(AwsCategoryListPoolToDoName, value)
	if err != nil {
		flog.Log.Errorf("ca todo list right now put err:%s", err.Error())
		return err
	}

	flog.Log.Debugf("ca pool add:%s", task.Link)
	return nil
}
