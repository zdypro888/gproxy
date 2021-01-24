package main

import (
	"log"
	"os/exec"
)

func init() {
	localChangeIP = rasidalChangeIP
}

func rasidalChangeIP() {
	rasdial := exec.Command("rasdial", Config.Connection, "/disconnect")
	if err := rasdial.Run(); err != nil {
		log.Printf("断开连接失败: %v", err)
	}
	rasdial = exec.Command("rasdial", Config.Connection, Config.UserName, Config.Password)
	if err := rasdial.Run(); err != nil {
		log.Printf("连接失败: %v", err)
	}
}
