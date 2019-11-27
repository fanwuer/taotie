package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"taotie/core/util"
	"time"
)

// redis tool
type Config struct {
	Host     string
	Password string
	DB       int
}

type MyRedis struct {
	Config Config
	Client *redis.Client
}

func NewRedisPool(config Config, size int) (*MyRedis, error) {
	r := &MyRedis{Config: config}
	client := redis.NewClient(&redis.Options{
		Addr:        config.Host,
		Password:    config.Password, // no password set
		DB:          config.DB,       // use default DB
		MaxRetries:  5,               // fail command retry 2
		PoolSize:    size,            // redis pool size
		DialTimeout: util.Second(20),
		// another options is default
	})

	pong, err := client.Ping().Result()
	if err == nil && pong == "PONG" {
		r.Client = client
	}
	return r, err
}

// set key
func (db *MyRedis) Set(key string, value interface{}, expire time.Duration) error {
	return db.Client.Set(key, value, expire).Err()
}

// get key
func (db *MyRedis) Get(key string) (string, bool, error) {
	result, err := db.Client.Get(key).Result()
	if err == redis.Nil {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	} else {
		return result, true, err
	}
}

func (db *MyRedis) Del(key string) error {
	return db.Client.Del(key).Err()
}

func (db *MyRedis) LPush(key string, values ...interface{}) (int64, error) {
	return db.Client.LPush(key, values...).Result()
}

func (db *MyRedis) LPushX(key string, values interface{}) (int64, error) {
	num, err := db.Client.LPushX(key, values).Result()
	if err != nil {
		return 0, err
	}
	if num == 0 {
		return 0, errors.New("list not exist")
	} else {
		return num, err
	}
}

func (db *MyRedis) RPush(key string, values ...interface{}) (int64, error) {
	return db.Client.RPush(key, values...).Result()
}

func (db *MyRedis) RPushX(key string, values interface{}) (int64, error) {
	num, err := db.Client.RPushX(key, values).Result()
	if err != nil {
		return 0, err
	}
	if num == 0 {
		return 0, errors.New("list not exist")
	} else {
		return num, err
	}
}

func (db *MyRedis) LLen(key string) (int64, error) {
	return db.Client.LLen(key).Result()
}

func (db *MyRedis) HLen(key string) (int64, error) {
	return db.Client.HLen(key).Result()
}

func (db *MyRedis) RPop(key string) (string, error) {
	return db.Client.RPop(key).Result()
}

func (db *MyRedis) LPop(key string) (string, error) {
	return db.Client.LPop(key).Result()
}

func (db *MyRedis) BRPop(timeout int, keys ...string) ([]string, error) {
	timeouts := time.Duration(timeout) * time.Second
	return db.Client.BRPop(timeouts, keys...).Result()
}

// if timeout is zero will be block until...
// and if  keys has many will return one such as []string{"pool","b"},pool is list,b is value
func (db *MyRedis) BLPop(timeout int, keys ...string) ([]string, error) {
	timeouts := time.Duration(timeout) * time.Second
	return db.Client.BLPop(timeouts, keys...).Result()
}

func (db *MyRedis) BRPopLPush(source, destination string, timeout int) (string, error) {
	timeouts := time.Duration(timeout) * time.Second
	return db.Client.BRPopLPush(source, destination, timeouts).Result()
}

func (db *MyRedis) RPopLPush(source, destination string) (string, error) {
	return db.Client.RPopLPush(source, destination).Result()
}

func (db *MyRedis) HExists(key, field string) (bool, error) {
	return db.Client.HExists(key, field).Result()
}

func (db *MyRedis) HGet(key, field string) (string, error) {
	return db.Client.HGet(key, field).Result()
}

func (db *MyRedis) HSet(key, field, value string) (bool, error) {
	return db.Client.HSet(key, field, value).Result()
}

// return item rem number if count==0 all rem if count>0 from the list head to rem
func (db *MyRedis) LRem(key string, count int64, value interface{}) (int64, error) {
	return db.Client.LRem(key, count, value).Result()
}
