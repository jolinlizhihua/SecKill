package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

//初始化service
func InitService(serviceConf *SecSkillConf)(err error){
	secKillConf = serviceConf
	err = loadBlackList()
	if err != nil {
		logs.Debug("init server success,config:%v",secKillConf)
		return
	}

	err = initProxy2LayerRedis()
	if err != nil {
		logs.Error("load proxy2Layer redis pool failed,err :%v",err)
		return
	}

	secKillConf.SecReqChan  = make(chan *SecRequest,secKillConf.SecReqChanSize)

	initRedisProcessFunc()

	return
}

func initRedisProcessFunc(){
	for i := 0; i < secKillConf.WriteProxy2LayerGoroutineNum; i++ {
		go WriteHandle()
	}
	for i :=0; i < secKillConf.ReadProxy2LayerGoroutineNum; i++ {
		go ReadHandle()
	}
}

func initProxy2LayerRedis()(err error){
	secKillConf.Proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:      secKillConf.RedisProxy2LyerConf.RedisMaxIdle,
		MaxActive:    secKillConf.RedisProxy2LyerConf.RedisMaxActive,
		IdleTimeout:  time.Duration(secKillConf.RedisProxy2LyerConf.RedisIdleTimeout)*time.Second,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp",secKillConf.RedisProxy2LyerConf.RedisAddr)
		},
	}
	//检测一下redis链接是否可以用
	conn := secKillConf.BlackRedisPool .Get()
	defer conn.Close()
	_,err = conn.Do("Set","name","lizhihua")
	if err != nil {
		logs.Error("ping redis err:",err)
		return
	}
	return
}

func initBlackRedis()(err error){
	secKillConf.BlackRedisPool = &redis.Pool{
		MaxIdle:      secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:    secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout:  time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout)*time.Second,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp",secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	//检测一下redis链接是否可以用
	conn := secKillConf.BlackRedisPool .Get()
	defer conn.Close()
	_,err = conn.Do("Set","name","lizhihua")
	if err != nil {
		logs.Error("ping redis err:",err)
		return
	}
	return
}

func loadBlackList()(err error){
	err = initBlackRedis()
	if err != nil {
		logs.Error("init black redis failed,err:%v",err)
		return
	}

	conn := secKillConf.BlackRedisPool.Get()
	defer conn.Close()

	reply,err := conn.Do("hgetall","idblacklist")
	idList ,err := redis.Strings(reply,err)
	if err != nil {
		logs.Error("hget all failed,err:%v",err)
		return
	}
	for _,v := range idList {
		id,err := strconv.Atoi(v)
		if err != nil {
			logs.Warn("invalid user id [%v]",id)
			continue
		}
		secKillConf.IdBlackMap[id] = true
	}

	reply,err = conn.Do("hgetall","ipblacklist")
	ipList ,err := redis.Strings(reply,err)
	if err != nil {
		logs.Error("hget all failed,err:%v",err)
		return
	}
	for _,v := range ipList {
		id,err := strconv.Atoi(v)
		if err != nil {
			logs.Warn("invalid ip id [%v]",id)
			continue
		}
		secKillConf.IdBlackMap[id] = true
	}
	go SyncIpBlackList()
	go SyncIdBlackList()
	return
}

//同步Ip黑名单
func SyncIpBlackList(){
	var ipList []string
	lastTime := time.Now().Unix()
	for{
		conn := secKillConf.BlackRedisPool.Get()
		defer conn.Close()
		reply,err := conn.Do("LPOP","blackiplist",time.Second)
		ip,err := redis.String(reply,err)
		if err != nil {
			continue
		}
		curtime := time.Now().Unix()
		ipList = append(ipList,ip)
		if len(ipList) > 100 || curtime-lastTime >5 {
			secKillConf.RWBlackLock.Lock()
			for _,v := range ipList {
				secKillConf.IpBlackMap[v] = true
			}
			secKillConf.RWBlackLock.Unlock()

			lastTime = curtime
			logs.Info("sync ip list from redis succ,ip[%v]",ipList)
		}
	}
}

//同步Id黑名单
func SyncIdBlackList(){
	var idList []int
	lastTime := time.Now().Unix()
	for{
		conn := secKillConf.BlackRedisPool.Get()
		defer conn.Close()
		reply,err := conn.Do("LPOP","blackiplist",time.Second)
		id,err := redis.Int(reply,err)
		if err != nil {
			continue
		}
		curtime := time.Now().Unix()
		idList = append(idList,id)
		if len(idList) > 100 || curtime-lastTime > 5 {
			secKillConf.RWBlackLock.Lock()
			for _,v := range idList {
				secKillConf.IdBlackMap[v] = true
			}
			secKillConf.RWBlackLock.Unlock()

			lastTime = curtime
			logs.Info("sync ip list from redis succ,id[%v]",idList)
		}

	}
}
