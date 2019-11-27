package config

const (
	Title   = "TaoTie"
	Version = "1.0.0"

	RoleWeb               = "Web"
	RoleAll               = "All"
	RoleProxy             = "Proxy"
	RoleAwsCategoryTimer  = "AwsCategoryTimer"
	RoleAwsAsinTimer      = "AwsAsinTimer"
	RoleAwsCategoryTask   = "AwsCategoryTask"
	RoleAwsAsinTask       = "AwsAsinTask"

	Host        = "0.0.0.0"
	Port        = 8080
	LogPath     = "data/log/svr.log"
	StoragePath = "data/storage"

	DbHost       = "127.0.0.1"
	DbDriverName = "mysql"
	DbName       = "taotie"
	DbPort       = "3306"
	DbPrefix     = "taotie_"
	DbLogPath    = "data/log/db.log"
	MaxIdleCons  = 10
	MaxOpenCons  = 20

	TimeZone          = 8
	SessionExpireTime = 24 * 3600 * 7

	RedisHost        = "127.0.0.1:6379"
	RedisMaxIdle     = 64
	RedisMaxActive   = 0
	RedisIdleTimeOut = 20

	MailPort    = 587
	MailSubject = "TaoTie Code"
	MailBody    = "%s Code is <br/> <p style='text-align:center'>%s</p> <br/>Valid in 5 minute."
)

var (
	ExpireTime = "20200201"
	//ExpireTime = "20190101"
	IsExpire = false
)
