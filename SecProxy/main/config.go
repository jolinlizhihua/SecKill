package main

import (
	"SecKill/SecProxy/service"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
)

var (
	secKillConf = &service.SecSkillConf{
		SecProductInfoMap: make(map[int]*service.SecProductInfoConf, 1024),
	}
)

func initConf() (err error) {
	redisAddr := beego.AppConfig.String("redis_black_addr")
	redisMaxIdle, err := beego.AppConfig.Int("redis_black_idle")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_black_idle[%s] config is error", err)
		return
	}
	redisMaxActive, err := beego.AppConfig.Int("redis_black_active")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_black_active[%s] config is error", err)
		return
	}
	redisIdleTimeout, err := beego.AppConfig.Int("redis_black_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_black_idle_timeout[%s] config is error", err)
		return
	}
	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed,etcd_timeout[%s] config is error", err)
		return
	}
	etcdAddr := beego.AppConfig.String("etcd_addr")

	logs.Debug("redisAddr=%v,etcdAddr=%v", redisAddr, etcdAddr)

	if len(redisAddr) == 0 || len(etcdAddr) == 0 {
		err = fmt.Errorf("init config failed,redis[%s] or etcd[%s] config is null", redisAddr, etcdAddr)
		return
	}

	secKillConf.RedisBlackConf.RedisAddr = redisAddr
	secKillConf.RedisBlackConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisBlackConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisBlackConf.RedisIdleTimeout = redisIdleTimeout
	secKillConf.EtcdConf.EtcdAddr = etcdAddr
	secKillConf.EtcdConf.Timeout = etcdTimeout
	secKillConf.EtcdConf.EtcdSecKeyPrefix = beego.AppConfig.String("etcd_sec_key_prefix")
	if len(secKillConf.EtcdConf.EtcdSecKeyPrefix) == 0 {
		err = fmt.Errorf("init config failed,etcd_sec_key_prefix err:%v", err)
		return
	}
	productKey := beego.AppConfig.String("etcd_sec_key_product")
	if len(productKey) == 0 {
		err = fmt.Errorf("init config failed,etcd_sec_key_product err:%v", err)
		return
	}
	if !strings.HasSuffix(secKillConf.EtcdConf.EtcdSecKeyPrefix, "/") {
		secKillConf.EtcdConf.EtcdSecKeyPrefix = secKillConf.EtcdConf.EtcdSecKeyPrefix + "/"
	}
	secKillConf.EtcdConf.EtcdSecProductKey = fmt.Sprintf("%s%s", secKillConf.EtcdConf.EtcdSecKeyPrefix, productKey)
	secKillConf.LogLevel = beego.AppConfig.String("log_level")
	secKillConf.LogPath = beego.AppConfig.String("log_path")

	secKillConf.CookieSecretKey = beego.AppConfig.String("cookie_secretKey")
	secLimit, err := beego.AppConfig.Int("user_sec_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed,read user_sec_access_limit err:%v", err)
		return
	}
	secKillConf.UserAccessLimit = secLimit

	referList := beego.AppConfig.String("refer_witeList")
	if len(referList) > 0 {
		secKillConf.ReferWiteList = strings.Split(referList, ",")
	}

	ipList, err := beego.AppConfig.Int("ip_sec_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed,read ip_sec_access_limit error:%v", err)
		return
	}
	secKillConf.IpSecAccessLimit = ipList

	redisProxy2LayerAddr := beego.AppConfig.String("redis_proxy2layer_addr")
	redisMaxIdle, err = beego.AppConfig.Int("redis_proxy2layer_idle")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_proxy2layer_idle[%s] config is error", err)
		return
	}
	redisMaxActive, err = beego.AppConfig.Int("redis_proxy2layer_active")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_proxy2layer_active[%s] config is error", err)
		return
	}
	redisIdleTimeout, err = beego.AppConfig.Int("redis_proxy2layer_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_black_idle_timeout[%s] config is error", err)
		return
	}

	secKillConf.RedisProxy2LyerConf.RedisAddr = redisProxy2LayerAddr
	secKillConf.RedisProxy2LyerConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisProxy2LyerConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisProxy2LyerConf.RedisIdleTimeout = redisIdleTimeout

	writeGoNums, err := beego.AppConfig.Int("write_proxy2layer_goroutine_num")
	if err != nil {
		err = fmt.Errorf("init config failed,write_proxy2layer_goroutine_num[%s] config is error", err)
		return
	}

	secKillConf.WriteProxy2LayerGoroutineNum = writeGoNums

	readGoNums, err := beego.AppConfig.Int("red_layer2proxy_goroutine_num")
	if err != nil {
		err = fmt.Errorf("init config failed,red_layer2proxy_goroutine_num[%s] config is error", err)
		return
	}

	secKillConf.ReadProxy2LayerGoroutineNum = readGoNums

	//读取业务逻辑层到proxy的redis配置
	redisLayer2ProxyAddr := beego.AppConfig.String("redis_layer2proxy_addr")

	if len(redisLayer2ProxyAddr) == 0 {
		err = fmt.Errorf("init config failed, redis[%s]  config is null", redisProxy2LayerAddr)
		return
	}

	redisMaxIdle, err = beego.AppConfig.Int("redis_layer2proxy_idle")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_layer2proxy_idle error:%v", err)
		return
	}

	redisMaxActive, err = beego.AppConfig.Int("redis_layer2proxy_active")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_layer2proxy_active error:%v", err)
		return
	}

	redisIdleTimeout, err = beego.AppConfig.Int("redis_layer2proxy_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_layer2proxy_idle_timeout error:%v", err)
		return
	}

	secKillConf.RedisLayer2ProxyConf.RedisAddr = redisLayer2ProxyAddr
	secKillConf.RedisLayer2ProxyConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisLayer2ProxyConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisLayer2ProxyConf.RedisIdleTimeout = redisIdleTimeout

	minIdLimit, err := beego.AppConfig.Int("user_min_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed, read user_min_access_limit error:%v", err)
		return
	}

	secKillConf.AccessLimitConf.UserMinAccessLimit = minIdLimit
	minIpLimit, err := beego.AppConfig.Int("ip_min_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed, read ip_min_access_limit error:%v", err)
		return
	}

	secKillConf.AccessLimitConf.IPMinAccessLimit = minIpLimit

	return
}
