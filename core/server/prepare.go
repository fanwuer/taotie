package server

import (
	"taotie/core/config"
	"time"
)

func CheckExpire() {
	t, _ := time.Parse("20060102", config.ExpireTime)
	s := t.Unix()
	if time.Now().Unix() > s {
		config.IsExpire = true
		return
	}
	for {
		select {
		case <-time.After(20 * time.Second):
			if time.Now().Unix() > s {
				config.IsExpire = true
				return
			}
		}
	}
}
