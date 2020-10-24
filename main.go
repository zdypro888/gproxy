package main

import "github.com/zdypro888/daemon"

const (
	name        = "gateproxy"
	description = "gateproxy service"
)

var dependencies = []string{}

func main() {
	if !daemon.Run(name, description, dependencies...) {
		return
	}
	startProxy()
	daemon.WaitNotify()
}
