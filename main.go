package main

import (
	"log"

	"github.com/zdypro888/daemon"
	"github.com/zdypro888/limit"
)

const (
	name        = "gateproxy"
	description = "gateproxy service"
)

var dependencies = []string{}

func main() {
	if !daemon.Run(name, description, dependencies...) {
		return
	}
	if err := Config.Load(); err != nil {
		log.Printf("读取配置错误: %v", err)
		Config.Save()
	}
	var err error
	if err = limit.Limit(); err != nil {
		log.Printf("limit error: %v", err)
	}
	startUSBMuxd()
	startProxy()
	daemon.WaitNotify()
}
