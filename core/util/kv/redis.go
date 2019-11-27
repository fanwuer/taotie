package kv

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

type MyRedisConf struct {
	RedisHost        string `yaml:"host"`
	RedisMaxIdle     int    `yaml:"max_idle"`
	RedisMaxActive   int    `yaml:"max_active"`
	RedisIdleTimeout int    `yaml:"idle_timeout"`
	RedisDB          int    `yaml:"database"`
	RedisPass        string `yaml:"pass"`
}

func NewRedis(redisConf *MyRedisConf) (pool *redis.Pool, err error) {
	pool = &redis.Pool{
		MaxIdle:     redisConf.RedisMaxIdle,
		MaxActive:   redisConf.RedisMaxActive,
		IdleTimeout: time.Duration(redisConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisConf.RedisHost, redis.DialPassword(redisConf.RedisPass), redis.DialDatabase(redisConf.RedisDB))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	conn := pool.Get()
	defer conn.Close()
	_, err = conn.Do("ping")
	return
}
