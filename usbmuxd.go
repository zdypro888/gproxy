package main

import (
	"net"
	"time"

	"github.com/zdypro888/usbmuxd"
)

var iControler *usbmuxd.DeviceControler

func startUSBMuxd() {
	iControler = usbmuxd.Service(name, description, dependencies...)
	if iControler == nil {
		return
	}
	iControler.OnPlug = onPlug
	iControler.Listen()
}

func onPlug(device *usbmuxd.USBDevice) bool {
	device.Object = &locker{}
	return true
}

type dialerTimeout struct {
	DialTimeout func(network, address string, t time.Duration) (net.Conn, error)
}

func (dt *dialerTimeout) Dial(network, addr string) (c net.Conn, err error) {
	return dt.DialTimeout(network, addr, 10*time.Second)
}
