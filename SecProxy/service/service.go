package service

import (
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

var(
	//由于无法引用main包中的变量，所以在service包中定义一个secKillConf
	secKillConf *SecSkillConf
)

//根据商品id查询商品信息
func SecInfo(productId int)(data []map[string]interface{},code int,err error){
	secKillConf.RwSecProductLock.RLock()
	defer secKillConf.RwSecProductLock.RUnlock()
	item,code,err := SecInfoById(productId)
	if err != nil {
		return
	}
	data = append(data,item)
	return
}

func SecInfoList()(data []map[string]interface{},code int,err error){
	secKillConf.RwSecProductLock.RLock()
	defer secKillConf.RwSecProductLock.RUnlock()

	for _,v := range secKillConf.SecProductInfoMap {
		item,_,err := SecInfoById(v.ProductId)
		if err != nil {
			logs.Error("product_id[%d] failed,err:%v",v.ProductId,err.Error())
			continue
		}
		data = append(data,item)
	}
	return
}

func SecInfoById(productId int)(data map[string]interface{},code int,err error){
	v,ok := secKillConf.SecProductInfoMap[productId]
	if !ok {
		code = ErrNotFoundProducrTd
		err = fmt.Errorf("not found product_id:%d",productId)
		return
	}

	start := false
	end := false
	status := "success"
	now := time.Now().Unix()

	if now - v.StartTime < 0{
		end = false
		start = false
		status = "sec kill is not start"
		code = ErrActiveNotStart
	}

	if now - v.StartTime > 0 {
		start = true
	}

	if now - v.EndTime > 0 {
		end = true
		start = false
		status = "sec kill is already end"
		code = ErrActiveAlreadyEnd
	}

	if v.Status == ProductStatusForceSaleOut || v.Status == ProductStatusSaleOut{
		start = false
		end  = true
		status = "Product is sale out"
		code = ErrActiveSaleOut
	}

	data = make(map[string]interface{})
	data["product_id"] = productId
	data["start"] = start
	data["end"] = end
	data["status"] = status
	return
}

func userCheck(req *SecRequest)(err error){

	found := false
	for _,refer := range secKillConf.ReferWiteList{
		if refer == req.ClientRefence {
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("inbalid request")
		logs.Warn("user[%d] is reject by refer,req[%v]",req.UserId,req)
		return
	}
	authData := fmt.Sprintf("%d:%s",req.UserId,secKillConf.CookieSecretKey)
	authSign := fmt.Sprintf("%x",md5.Sum([]byte(authData)))
	if authSign != req.UserAuthSign {
		err = fmt.Errorf("invalid user cookie auth")
		return
	}
	return
}

func SecKill(req *SecRequest)(data map[string]interface{},code int,err error){
	secKillConf.RwSecProductLock.RLock()
	defer secKillConf.RwSecProductLock.RUnlock()

	err = userCheck(req)
	if err != nil {
		code = ErrUserCheckAuthFailed
		logs.Warn("userId[%d] invalid,check failed,req[%v]",req.UserId,req)
	}
	err = antiSpan(req)
	if err != nil {
		code = ErrUserServiceBusy
		logs.Warn("service busy")
		return
	}
	data,code,err = SecInfoById(req.ProductId)
	if err != nil {
		logs.Warn("userId[%d] invalid,check failed,req[%v]",req.UserId,req)
		return
	}

	if code != 0 {
		logs.Warn("userId[%d] secInfoById failed,code[%d] req[%v]",req.UserId,code,req)
		return
	}

	secKillConf.SecReqChan <- req
	return
}

