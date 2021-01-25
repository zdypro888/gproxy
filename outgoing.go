package main

import "net"

var outgoing locker
var localChangeIP func()

type outgoingConn struct {
	net.Conn
	Locker *locker
	Change func()
}

//Close 关闭
func (conn *outgoingConn) Close() error {
	if conn.Change != nil {
		go conn.changeIP()
	} else {
		conn.Locker.Unlock()
	}
	return conn.Conn.Close()
}

func (conn *outgoingConn) changeIP() {
	conn.Change()
	conn.Locker.Unlock()
}
