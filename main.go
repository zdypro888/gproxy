package main

import (
	"log"

	"github.com/zdypro888/daemon"
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
	startProxy()
	daemon.WaitNotify()
}
