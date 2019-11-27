package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"taotie/core/model"
	"taotie/core/util"
	"taotie/core/util/kv"
)

var (
	// redis key
	redisToken = "ff_tokens"
	redisUser  = "ff_users"
)

// diy user redis
type TokenManage interface {
	CheckAndSetToken(token string, validTimes int64) (user *model.User, err error) // Check the token, when redis exist direct return user info, others hit the mysql db and save in redis then return
	SetToken(user *model.User, validTimes int64) (token string, err error)         // Set token, expire 7 days
	RefreshToken(token string, validTime int64) error                              // Refresh token，token expire time will be again 7 days
	DeleteToken(token string) error                                                // Delete token when logout
	RefreshUser(id []int64, validTime int64) error                                 // Refresh redis cache of user info
	DeleteUserToken(id int64) error                                                // Delete all token of those user
	DeleteUser(id int64) error                                                     // Delete user info in redis cache
	AddUser(id int64, validTime int64) (user *model.User, err error)               // Add the user info to session redis，expire days:7
}

type RedisSession struct {
	Pool *redis.Pool
}

func (s *RedisSession) Set(key string, value []byte, expireSecond int64) (err error) {
	conn := s.Pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer conn.Close()
	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("SET", key, value)
	if err != nil {
		return err
	}

	if expireSecond <= 0 {
		expireSecond = 7 * 24 * 3600
	}
	err = conn.Send("EXPIRE", key, expireSecond)
	if err != nil {
		return err
	}
	_, err = conn.Do("EXEC")
	return
}

func (s *RedisSession) Delete(key string) (err error) {
	conn := s.Pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return err
}

func (s *RedisSession) EXPIRE(key string, expireSecond int) (err error) {
	conn := s.Pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer conn.Close()
	if expireSecond <= 0 {
		expireSecond = 7 * 24 * 3600
	}
	_, err = conn.Do("EXPIRE", key, expireSecond)
	return err
}

func (s *RedisSession) Keys(pattern string) (result []string, exist bool, err error) {
	conn := s.Pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer conn.Close()
	keys, err := redis.ByteSlices(conn.Do("KEYS", pattern))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	result = make([]string, len(keys))
	for k, v := range keys {
		result[k] = string(v)
	}
	return result, true, nil
}

func (s *RedisSession) Get(key string) (value []byte, exist bool, err error) {
	conn := s.Pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	value, err = redis.Bytes(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func HashTokenKey(token string) string {
	return fmt.Sprintf("%s_%s", redisToken, token)
}

func GenToken(id int64) string {
	return fmt.Sprintf("%d_%s", id, util.GetGUID())
}

func HashUserKey(id int64, name string) string {
	return fmt.Sprintf("%s_%d_%s", redisUser, id, name)
}

func UserKeys(id int64) string {
	return fmt.Sprintf("%s_%d_*", redisUser, id)
}

func UserTokenKeys(id int64) string {
	return fmt.Sprintf("%s_%d_*", redisToken, id)
}
func (s *RedisSession) CheckAndSetToken(token string, validTimes int64) (user *model.User, err error) {
	if token == "" {
		err = errors.New("token nil")
		return
	}

	value, exist, err := s.Get(HashTokenKey(token))
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.New("token not found")
	}

	userKey := string(value)
	value, exist, err = s.Get(userKey)
	if err != nil {
		return nil, err
	}

	if exist {
		user = new(model.User)
		json.Unmarshal(value, user)
		return
	}

	temp := strings.Split(userKey, "_")
	if len(temp) != 3 || temp[0] != redisUser {
		return nil, errors.New("token invalid")
	}

	id, err := strconv.Atoi(temp[1])
	if err != nil {
		return nil, errors.New("token invalid")
	}
	user, err = s.AddUser(int64(id), validTimes)
	return
}

func (s *RedisSession) RefreshToken(token string, validTimes int64) (err error) {
	return s.EXPIRE(HashTokenKey(token), int(validTimes))
}

func (s *RedisSession) DeleteToken(token string) (err error) {
	return s.Delete(HashTokenKey(token))
}

func (s *RedisSession) DeleteUserToken(id int64) (err error) {
	result, exist, err := s.Keys(UserTokenKeys(id))
	if err == nil && exist {
		for _, v := range result {
			s.Delete(v)
		}
	}
	return
}

func (s *RedisSession) DeleteUser(id int64) (err error) {
	result, exist, err := s.Keys(UserKeys(id))
	if err == nil && exist {
		for _, v := range result {
			return s.Delete(v)
		}
	}
	return
}

func (s *RedisSession) AddUser(id int64, validTimes int64) (user *model.User, err error) {
	user = new(model.User)
	user.Id = id
	exist, err := user.GetRaw()
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.New("user not exist in db")
	}

	user.Password = ""
	user.ActivateCode = ""
	user.ActivateCodeExpired = 0
	user.ResetCode = ""
	user.ResetCodeExpired = 0
	userKey := HashUserKey(user.Id, user.Name)
	raw, _ := json.Marshal(user)
	err = s.Set(userKey, raw, validTimes)
	if err != nil {
		return nil, err
	}

	return
}

func (s *RedisSession) RefreshUser(ids []int64, validTime int64) (err error) {
	for _, id := range ids {
		s.AddUser(id, validTime)
	}
	return
}

func (s *RedisSession) SetToken(user *model.User, validTimes int64) (token string, err error) {
	if user == nil || user.Id == 0 {
		err = errors.New("user nil")
		return
	}

	user.Password = ""
	user.ActivateCode = ""
	user.ActivateCodeExpired = 0
	user.ResetCode = ""
	user.ResetCodeExpired = 0

	token = GenToken(user.Id)
	userKey := HashUserKey(user.Id, user.Name)
	err = s.Set(HashTokenKey(token), []byte(userKey), validTimes)
	if err != nil {
		return
	}

	raw, _ := json.Marshal(user)
	s.Set(userKey, raw, validTimes)
	return
}

var FafaSessionMgr TokenManage

func InitSession(redisConf kv.MyRedisConf) error {
	pool, err := kv.NewRedis(&redisConf)
	if err != nil {
		return err
	}
	FafaSessionMgr = &RedisSession{Pool: pool}
	return nil
}
