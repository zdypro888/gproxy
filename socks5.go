package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/zdypro888/go-socks5"
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

var outgoingLock sync.Mutex
var outgoingCount int
var outgoingLocking bool

func outLock(lock bool) error {
	outgoingLock.Lock()
	defer outgoingLock.Unlock()
	if outgoingLocking {
		return errors.New("outgoing tunnel is locking")
	}
	if lock {
		if outgoingCount > 0 {
			return fmt.Errorf("outgoing has client: %d", outgoingCount)
		}
		outgoingLocking = true
	}
	outgoingCount++
	return nil
}

func outUnlock() {
	outgoingLock.Lock()
	defer outgoingLock.Unlock()
	outgoingLocking = false
	outgoingCount--
}

type outgoingConn struct {
	net.Conn
	Change bool
}

//Close 关闭
func (conn *outgoingConn) Close() error {
	if conn.Change {
		go changeIP()
	} else {
		outUnlock()
	}
	return conn.Conn.Close()
}

var localChangeIP func()

func changeIP() {
	//更换IP
	if localChangeIP != nil {
		localChangeIP()
	}
	outUnlock()
}

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	request := ctx.Value(contextKey("request")).(*socks5.Request)
	username := request.AuthContext.Payload["username"]
	password := request.AuthContext.Payload["password"]
	log.Printf("[%s]连接(%s) %s://%s", username, password, network, addr)
	auto, must, change, lock := commands(password)
	if auto || must {
		//应该是 switch 进程
	} else {
		if err := outLock(lock); err != nil {
			return nil, err
		}
		conn, err := net.Dial(network, addr)
		if err != nil {
			outUnlock()
			return nil, err
		}
		return &outgoingConn{Conn: conn, Change: change}, nil
	}
	return nil, nil
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
