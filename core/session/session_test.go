package session

import (
	"fmt"
	"taotie/core/model"
	"taotie/core/util/kv"
	"testing"
)

func TestRedisSession_CheckToken(t *testing.T) {
	pool, err := kv.NewRedis(&kv.MyRedisConf{
		RedisHost:        "192.168.91.129:6379",
		RedisPass:        "123456789",
		RedisDB:          0,
		RedisIdleTimeout: 15,
		RedisMaxActive:   0,
		RedisMaxIdle:     0,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	s := &RedisSession{Pool: pool}
	token, err := s.SetToken(&model.User{
		Id:   3,
		Name: "SSSSS",
	}, 2000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(token)
}
