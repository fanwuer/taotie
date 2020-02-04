package spider

import (
	"taotie/core/util"
	"taotie/core/util/redis"
	"time"
)

var (
	Pool *redis.MyRedis
)

func InitRedisPool(host, pass string, db, size int) error {
	if Pool != nil {
		return nil
	}
	pool, err := redis.NewRedisPool(redis.Config{
		Host:     host,
		Password: pass,
		DB:       db,
	}, size)
	Pool = pool
	return err
}

func awsPoolKeyCanLive(pool string, key string, maxLiveTime time.Duration) (exist, can bool, lastTime int64, err error) {
	now := time.Now().Unix()
	info, exist, err := GetHashPool(pool, key)
	if err != nil {
		return false, false, 0, err
	}
	if exist {
		i, _ := util.SInt64(info)

		if now-i > int64(maxLiveTime.Seconds()) {
			return true, false, i, nil
		}

		return true, true, i, nil
	}

	return false, false, 0, nil
}

func PutHashPool(pool, k string, v interface{}) error {
	return Pool.Set(pool+ValueSpilt+k, v, AwsPoolExpireTime)
}

func DeleteHashPool(pool, k string) error {
	return Pool.Del(pool + ValueSpilt + k)
}

func RPushListPool(pool, v string) error {
	_, err := Pool.RPush(pool, v)
	return err
}

func LPushListPool(pool, v string) error {
	_, err := Pool.LPush(pool, v)
	return err
}

func BLPopListPool(pool string) ([]string, error) {
	return Pool.BLPop(0, pool)
}

func LRemListPool(pool, k string) error {
	_, err := Pool.LRem(pool, 0, k)
	return err
}

func GetHashPool(pool, k string) (string, bool, error) {
	return Pool.Get(pool + ValueSpilt + k)
}
