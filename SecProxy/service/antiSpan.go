package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
)

type SecLimitMgr struct {
	UserLimitMap map[int]*Limit
	IpLimitMap   map[string]*Limit
	lock         sync.Mutex
}

func antiSpan(req *SecRequest) (err error) {

	_, ok := secKillConf.IdBlackMap[req.UserId]
	if ok {
		err = fmt.Errorf("invalid request")
		logs.Error("userId[%v] is block by id black list", req.UserId)
		return
	}

	_, ok = secKillConf.IpBlackMap[req.ClientAddr]
	if ok {
		err = fmt.Errorf("invalid request")
		logs.Error("userId[%v] ip[%v] is block by id black list", req.UserId, req.ClientAddr)
		return
	}

	secKillConf.SecLimitMgr.lock.Lock()
	//uid频率控制
	limit, ok := secKillConf.SecLimitMgr.UserLimitMap[req.UserId]
	if !ok {
		limit = &Limit{
			secLimit: &SecLimit{},
			minLimit: &MinLimit{},
		}
		secKillConf.SecLimitMgr.UserLimitMap[req.UserId] = limit
	}

	secIdCount := limit.secLimit.Count(req.AccessTime.Unix())
	minIdCount := limit.minLimit.Count(req.AccessTime.Unix())

	//ip频率控制
	limit, ok = secKillConf.SecLimitMgr.IpLimitMap[req.ClientAddr]
	if !ok {
		limit = &Limit{
			secLimit: &SecLimit{},
			minLimit: &MinLimit{},
		}
		secKillConf.SecLimitMgr.IpLimitMap[req.ClientAddr] = limit
	}

	secIpCount := limit.secLimit.Count(req.AccessTime.Unix())
	minIpCount := limit.minLimit.Count(req.AccessTime.Unix())
	secKillConf.SecLimitMgr.lock.Unlock()

	if secIpCount > secKillConf.AccessLimitConf.IPMinAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	if minIpCount > secKillConf.AccessLimitConf.IPMinAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	if secIdCount > secKillConf.AccessLimitConf.UserSecAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	if minIdCount > secKillConf.AccessLimitConf.UserMinAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}
	return
}
