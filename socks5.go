package main

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"

	"github.com/zdypro888/go-socks5"
	"github.com/zdypro888/usbmuxd"
	"golang.org/x/net/proxy"
)

//StaticBoolCredentials 认证通过
type StaticBoolCredentials bool

//Valid 认证
func (s StaticBoolCredentials) Valid(user, password, userAddr string) bool {
	return bool(s)
}

//AuthSetRule 认证通过
type AuthSetRule struct {
}

type contextKey string

//Allow 是否允许
func (rule *AuthSetRule) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	return context.WithValue(ctx, contextKey("request"), req), true
}

func commands(commands string) (bool, bool, bool, bool) {
	var auto, must, change, lock bool
	controls := strings.Split(commands, "-")
	for _, command := range controls {
		switch command {
		case "auto":
			auto = true
		case "must":
			must = true
		case "change":
			change = true
		case "lock":
			lock = true
		}
	}
	return auto, must, change, lock
}

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	request := ctx.Value(contextKey("request")).(*socks5.Request)
	username := request.AuthContext.Payload["username"]
	password := request.AuthContext.Payload["password"]
	log.Printf("[%s]连接(%s) %s://%s", username, password, network, addr)
	auto, must, change, lock := commands(password)
	if auto || must {
		//应该是 switch 进程
		if iControler.DeviceCount == 0 {
			return nil, errors.New("device not found(none)")
		}
		var minDevice *usbmuxd.USBDevice
		iControler.Devices.Range(func(key, value interface{}) bool {
			device := value.(*usbmuxd.USBDevice)
			if device.Pluged {
				deviceLck := device.Object.(*locker)
				if err := deviceLck.Lock(lock); err == nil {
					if minDevice == nil {
						minDevice = device
						if lock {
							return false
						}
					} else {
						minLocker := minDevice.Object.(*locker)
						if minLocker.Count > deviceLck.Count {
							minLocker.Unlock()
							minDevice = device
						} else {
							deviceLck.Unlock()
						}
					}
				}
			}
			return true
		})
		if minDevice == nil {
			return nil, errors.New("device not found(busy)")
		}
		deviceLocker := minDevice.Object.(*locker)
		var devicePass []string
		if change {
			devicePass = append(devicePass, "change")
		}
		if lock {
			devicePass = append(devicePass, "lock")
		}
		dialer, err := proxy.SOCKS5("usbmuxd", "1080", &proxy.Auth{User: username, Password: strings.Join(devicePass, "-")}, &dialerTimeout{DialTimeout: minDevice.DialTimeout})
		if err != nil {
			deviceLocker.Unlock()
			return nil, err
		}
		conn, err := dialer.Dial(network, addr)
		if err != nil {
			deviceLocker.Unlock()
			return nil, err
		}
		out := &outgoingConn{Conn: conn, Locker: deviceLocker}
		return out, nil
	}
	if err := outgoing.Lock(lock); err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, addr)
	if err != nil {
		outgoing.Unlock()
		return nil, err
	}
	out := &outgoingConn{Conn: conn, Locker: &outgoing}
	if change {
		out.Change = localChangeIP
	}
	return out, nil
}

//startProxy 开启代理
func startProxy() error {
	server := socks5.NewServer(
		// socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
		socks5.WithCredential(StaticBoolCredentials(true)),
		socks5.WithRule(&AuthSetRule{}),
		socks5.WithDial(dial),
	)
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		return err
	}
	go server.Serve(listener)
	return nil
}
