package config

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"taotie/core/util/kv"
	"taotie/core/util/mail"
	"taotie/core/util/oss"
	"taotie/core/util/rdb"
)

var (
	GlobalConfig *Config
)

type Config struct {
	DefaultConfig MyConfig       `yaml:"global"`
	DbConfig      rdb.MyDbConfig `yaml:"db"`
	OssConfig     oss.Key        `yaml:"oss"`
	KVConfig      kv.MyRedisConf `yaml:"redis"`
	MailConfig    mail.Sender    `yaml:"mail"`
	SpiderConfig  MySpiderConfig `yaml:"spider"`
}

type MyConfig struct {
	Host              string `yaml:"host"`
	Port              int64  `yaml:"port"`
	LogPath           string `yaml:"log_path"`
	TimeZone          int64  `yaml:"time_zone"`
	SingleLogin       bool   `yaml:"single_login"`
	SessionExpireTime int64  `yaml:"session_expire_time"`
	StoragePath       string `yaml:"storage_path"`
	IsOss             bool   `yaml:"is_oss"`
	CanMail           bool   `yaml:"can_mail"`
	Role              string `yaml:"role"`
}

type MySpiderConfig struct {
	Account               string `yaml:"account"`
	AwsCategoryTaskThread int64  `yaml:"aws_category_thread"`
	AwsAsinTaskThread     int64  `yaml:"aws_asin_thread"`
	//Redis         redis.Config `yaml:"redis"`
}

func JsonOutConfig(config Config) (string, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	back := string(raw)
	return back, nil
}

func InitConfig(configFilePath string) error {
	if GlobalConfig != nil {
		return nil
	}
	c, err := InitYamlConfig(configFilePath)
	if err != nil {
		return err
	}
	GlobalConfig = c
	return nil
}

func deal(c *Config) error {
	c.DbConfig.DriverName = DbDriverName
	c.DbConfig.Prefix = DbPrefix
	if c.DefaultConfig.LogPath == "" {
		c.DefaultConfig.LogPath = LogPath
	}

	if c.DefaultConfig.StoragePath == "" {
		c.DefaultConfig.StoragePath = StoragePath
	}

	if c.DefaultConfig.Host == "" {
		c.DefaultConfig.Host = Host
	}
	if c.DefaultConfig.Port == 0 {
		c.DefaultConfig.Port = Port
	}

	if c.DbConfig.Host == "" {
		c.DbConfig.Host = DbHost
	}
	if c.DbConfig.Port == "" {
		c.DbConfig.Port = DbPort
	}

	if c.DbConfig.MaxIdleConns == 0 {
		c.DbConfig.MaxIdleConns = MaxIdleCons
	}

	if c.DbConfig.MaxOpenConns == 0 {
		c.DbConfig.MaxOpenConns = MaxOpenCons
	}

	if c.DbConfig.Name == "" {
		c.DbConfig.Name = DbName
	}
	if c.DbConfig.Debug {
		if c.DbConfig.DebugToFile && c.DbConfig.DebugToFileName == "" {
			c.DbConfig.DebugToFileName = DbLogPath
		}
	}

	if c.KVConfig.RedisHost == "" {
		c.KVConfig.RedisHost = RedisHost
	}

	if c.KVConfig.RedisIdleTimeout == 0 {
		c.KVConfig.RedisIdleTimeout = RedisIdleTimeOut
	}

	if c.KVConfig.RedisMaxActive == 0 {
		c.KVConfig.RedisMaxActive = RedisMaxActive
	}

	if c.KVConfig.RedisMaxIdle == 0 {
		c.KVConfig.RedisMaxIdle = RedisMaxIdle
	}
	if c.DefaultConfig.TimeZone == 0 {
		c.DefaultConfig.TimeZone = TimeZone
	}

	if c.DefaultConfig.SessionExpireTime == 0 {
		c.DefaultConfig.SessionExpireTime = SessionExpireTime
	}

	if c.MailConfig.Port == 0 {
		c.MailConfig.Port = MailPort
	}

	if c.MailConfig.Subject == "" {
		c.MailConfig.Subject = MailSubject
	}

	if c.MailConfig.Body == "" {
		c.MailConfig.Body = MailBody
	}
	return nil
}

func InitJsonConfig(configFilePath string) (*Config, error) {
	c := new(Config)
	if configFilePath == "" {
		return nil, errors.New("config file empty")
	}

	raw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, c)
	if err != nil {
		return nil, err
	}

	err = deal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func InitYamlConfig(configFilePath string) (*Config, error) {
	c := new(Config)
	if configFilePath == "" {
		return nil, errors.New("config file empty")
	}

	raw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(raw, c)
	if err != nil {
		return nil, err
	}
	err = deal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
