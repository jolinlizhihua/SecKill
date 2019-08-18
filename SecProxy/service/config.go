package service

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

const(
	ProductStatusNormal = 0
	ProductStatusSaleOut = 1
	ProductStatusForceSaleOut = 2
)

type RedisConf struct {
	RedisAddr string
	RedisMaxIdle int
	RedisMaxActive int
	RedisIdleTimeout int
}

type EtcdConf struct {
	EtcdAddr string
	Timeout int
	EtcdSecKeyPrefix string  //秒杀系统etcd的前缀key
	EtcdSecProductKey string
}

type SecSkillConf struct {
	RedisBlackConf RedisConf
	RedisProxy2LyerConf RedisConf

	EtcdConf EtcdConf
	LogPath string
	LogLevel string
	SecProductInfoMap map[int]*SecProductInfoConf
	RwSecProductLock sync.RWMutex
	CookieSecretKey string
	UserAccessLimit int
	ReferWiteList []string
	IpSecAccessLimit int
	IpBlackMap map[string]bool
	IdBlackMap map[int]bool

	BlackRedisPool *redis.Pool
	Proxy2LayerRedisPool *redis.Pool

	RWBlackLock sync.RWMutex

	WriteProxy2LayerGoroutineNum int
	ReadProxy2LayerGoroutineNum int

	SecReqChan chan *SecRequest
	SecReqChanSize int

	UserConnMap map[string]chan *SecResult
	UserConnMapLock sync.Mutex
}

//描述信息结构体
type SecProductInfoConf struct {
	ProductId int
	StartTime int64
	EndTime int64
	Status int
	Count int
	Left int
}

type SecResult struct {
	ProductId int
	UserId int
	Code int
	Token string
}

type SecRequest struct {
	ProductId int
	Source string
	AuthCode string
	SecTime string
	Nance string
	UserId int
	UserAuthSign string
	AccessTime time.Time
	ClientAddr string
	ClientRefence string
}
