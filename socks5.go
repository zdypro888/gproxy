package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/zdypro888/go-socks5"
	zhttp "github.com/zdypro888/http"
	netproxy "golang.org/x/net/proxy"
)

//StaticBoolCredentials 认证通过
type StaticBoolCredentials bool

//Valid 认证
func (s StaticBoolCredentials) Valid(user, password string) bool {
	return bool(s)
}

//UserPassAuthenticator 认证
type UserPassAuthenticator struct {
	impl *socks5.UserPassAuthenticator
}

//GetCode 取得代码
func (auth *UserPassAuthenticator) GetCode() uint8 {
	return auth.impl.GetCode()
}

//Authenticate 取得认证信息
func (auth *UserPassAuthenticator) Authenticate(reader io.Reader, writer io.Writer) (*socks5.AuthContext, error) {
	actx, err := auth.impl.Authenticate(reader, writer)
	if err != nil {
		return nil, err
	}
	// if username, ok := actx.Payload["Username"]; ok {
	// 	useraddr, err := getAddress(username, false)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	actx.Payload["Address"] = useraddr
	// 	return actx, nil
	// }
	// return nil, socks5.UserAuthFailed
	return actx, nil
}

type addressInfo struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID         string `json:"ID"`
		Type       string `json:"ProxyType"`
		ExportIPV6 bool   `json:"ExportIPV6"`
		Address    string `json:"ProxyAddress"`
		UserName   string `json:"ProxyUserName"`
		Password   string `json:"ProxyPassword"`
		UDID       string `json:"UDID"`
		RerenceID  string `json:"RerenceID"`
	} `json:"Response"`
}

func getAddress(url string, withNew bool) (*addressInfo, error) {
	response, err := http.Get(fmt.Sprintf("%s&new=%v", url, withNew))
	if err != nil {
		return nil, err
	}
	jsonData, err := zhttp.ReadResponse(response)
	if err != nil {
		return nil, err
	}
	info := &addressInfo{}
	if err := json.Unmarshal(jsonData, info); err != nil {
		return nil, err
	}
	return info, nil
}

func dialRemote(ctx context.Context, network, addr string) (net.Conn, error) {
	username := ctx.Value("Username").(string)
	// password := ctx.Value("Password").(string)
	var err error
	var withNew bool
	var addrinfo *addressInfo
	for i := 0; i < 10; i++ {
		if addrinfo, err = getAddress(username, withNew); err != nil {
			log.Printf("请求地址错误: %v", err)
		} else {
			break
		}
	}
	//连续10次取得错误, 返回最后错误
	if addrinfo == nil {
		return nil, err
	}
	switch addrinfo.Data.Type {
	case "local":
		ifaddrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, err
		}
		for _, ifaddr := range ifaddrs {
			if ifaddr.String() == addrinfo.Data.Address {
				dialer := new(net.Dialer)
				dialer.LocalAddr = ifaddr
				return net.Dial(network, addr)
			}
		}
		return nil, &net.AddrError{Err: "can not find", Addr: addrinfo.Data.Address}
	case "socks5":
		var auth *netproxy.Auth
		if addrinfo.Data.UserName != "" {
			auth = &netproxy.Auth{User: addrinfo.Data.UserName, Password: addrinfo.Data.Password}
		}
		dialer, err := netproxy.SOCKS5("tcp", addrinfo.Data.Address, auth, netproxy.Direct)
		if err != nil {
			return nil, err
		}
		return dialer.Dial(network, addr)
	case "http":
		conn, err := net.Dial("tcp", addrinfo.Data.Address)
		if err != nil {
			return nil, err
		}
		hdr := make(http.Header)
		if addrinfo.Data.UserName != "" {
			auth := addrinfo.Data.UserName + ":" + addrinfo.Data.Password
			hdr.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
		}
		connectReq := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: hdr,
		}
		if err = connectReq.Write(conn); err != nil {
			return nil, err
		}
		br := bufio.NewReader(conn)
		response, err := http.ReadResponse(br, connectReq)
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("connect http tunnel faild: %d", response.StatusCode)
		}
		return conn, nil
	}

	return nil, err
}

//startProxy 开启代理
func startProxy() error {
	auth := &UserPassAuthenticator{
		impl: &socks5.UserPassAuthenticator{
			Credentials: StaticBoolCredentials(true),
		},
	}
	socks5conf := &socks5.Config{
		Dial:        dialRemote,
		AuthMethods: []socks5.Authenticator{auth},
	}
	socks5srv, err := socks5.New(socks5conf)
	if err != nil {
		return err
	}
	socks5lis, err := net.Listen("tcp", ":1080")
	if err != nil {
		return err
	}
	go socks5srv.Serve(socks5lis)
	return nil
}
