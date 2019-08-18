package main

import (
	"SecKill/SecProxy/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	etcd_client "github.com/coreos/etcd/clientv3"
	"time"
)

var(
	redisPool *redis.Pool
	etcdClient *etcd_client.Client

)

//初始化redis
func initRedis()(err error){
	redisPool = &redis.Pool{
		MaxIdle:      secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:    secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout:  time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout)*time.Second,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp",secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	//检测一下redis链接是否可以用
	conn := redisPool.Get()
	defer conn.Close()
	_,err = conn.Do("Set","name","lizhihua")
	if err != nil {
		logs.Error("ping redis err:",err)
		return
	}
	return
}

//初始化etcd
func initEtcd()(err error){
	cli,err := etcd_client.New(etcd_client.Config{
		Endpoints: []string{secKillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secKillConf.EtcdConf.Timeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err:",err)
		return
	}
	etcdClient = cli
	return
}

func converLogLevel(level string) int{
	switch(level){
		case "debug":
			return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

//初始化日志
func initLogs() (err error){
	config := make(map[string]interface{})
	config["filename"] = secKillConf.LogPath
	config["level"] = converLogLevel(secKillConf.LogLevel)
	configStr,err := json.Marshal(config)
	if err != nil {
		fmt.Println("maeshal failed,err：",err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func loadSecConf()(err error){
	resp,err := etcdClient.Get(context.Background(),secKillConf.EtcdConf.EtcdSecProductKey)
	if err != nil {
		logs.Error("get [%s] from etcd failed,err:%v",err)
		return
	}
	var secProductInfo []service.SecProductInfoConf
	for k,v := range resp.Kvs {
		logs.Debug("key[%s] valued[%s]",k,v)
		err = json.Unmarshal(v.Value,&secProductInfo)
		if err != nil {
			logs.Error("unmarshal failed,err:",err)
			return
		}
		logs.Debug("sec info conf is [%v]",secProductInfo)
	}
	updateSecProductInfo(secProductInfo)
	return
}

func initSecProductWatcher() (err error){
	go watchSecProductKey(secKillConf.EtcdConf.EtcdSecProductKey)
	return
}

func watchSecProductKey(key string) {

	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return
	}

	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []service.SecProductInfoConf
		var getConfSucc = true

		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] 's config deleted", key)
					continue
				}

				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key [%s], Unmarshal[%s], err:%v ", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd, %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}

			if getConfSucc {
				logs.Debug("get config from etcd succ, %v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}

	}
}

func updateSecProductInfo(secKillProductInfo []service.SecProductInfoConf) {
	var tmp map[int]*service.SecProductInfoConf = make(map[int]*service.SecProductInfoConf,1024)
	for _,v := range secKillProductInfo{
		productInfo := v
		tmp[v.ProductId] = &productInfo
	}
	secKillConf.RwSecProductLock.Lock()
	secKillConf.SecProductInfoMap = tmp
	secKillConf.RwSecProductLock.Unlock()
}

func initSec()(err error){

	err = initLogs()
	if err != nil {
		logs.Error("init log failed,err:",err)
		return
	}

	err = initRedis()
	if err != nil {
		logs.Error("init redis failed,err:%v",err)
		return
	}

	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed,err:%v",err)
		return
	}

	//加载
	err = loadSecConf()
	if err != nil {
		logs.Error("init sec conf failed,err:",err)
		return
	}

	err = initSecProductWatcher()
	if err != nil {
		logs.Error("init sec conf watcher failed,err:",err)
		return
	}

	service.InitService(secKillConf)

	logs.Info("init sec success")
	return

}
