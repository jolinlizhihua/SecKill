package main

import (
	_ "SecKill/SecProxy/router"
	"github.com/astaxie/beego"
)

func main() {

	err := initConf()
	if err != nil {
		panic(err)
		return
	}

	err = initSec()
	if err != nil {
		panic(err)
		return
	}
	beego.Run()
}
