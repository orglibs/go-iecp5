// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

// ClientOption Configuration
type ClientOption struct {
	config            Config
	params            asdu.Params
	server            *url.URL      // Connected server side
	autoReconnect     bool          // if true, reconnection automatically
	reconnectInterval time.Duration // Reconnection interval
	TLSConfig         *tls.Config   // tls configuration
	// DialContext 可选的传输建立函数；nil 使用内置 TCP/TLS 拨号。
	// 回调须遵守 ctx 取消，并返回已完成 TLS 等握手的连接；用于代理或内存链路测试。
	DialContext func(ctx context.Context, remote *url.URL) (net.Conn, error)
	// OnAPDU 是本客户端独立的可选报文观察器；nil 不生成观察器快照。
	// outbound=true 表示整帧已写入连接（不代表远端确认），false 表示完整接收帧进入状态机，
	// 接收通知早于 APCI/ASDU 校验，因此可观察被拒绝的畸形帧；不包含截断帧和重同步丢弃字节。
	// apdu 是独立副本，回调可保留但不能依赖它修改协议处理；收发回调可能并发执行，
	// 实现必须并发安全、及时返回，不得阻塞等待本客户端的关闭、发送或业务确认。
	// 必须在 NewClient 前设置；自动重连沿用该观察器，不改变全局 slog 配置。
	OnAPDU func(outbound bool, apdu []byte)
}

// NewOption with default config and default asdu.ParamsWide params
func NewOption() *ClientOption {
	return &ClientOption{
		config:            DefaultConfig(),
		params:            *asdu.ParamsWide,
		server:            nil,
		autoReconnect:     true,
		reconnectInterval: DefaultReconnectInterval,
		TLSConfig:         nil,
	}
}

// SetConfig sets the config if config is not valid it will use DefaultConfig()
func (sf *ClientOption) SetConfig(cfg Config) *ClientOption {
	if err := cfg.ValidConfigServer(); err != nil {
		sf.config = DefaultConfig()
	} else {
		sf.config = cfg
	}

	return sf
}

// SetParams set asdu params if params is not valid it will use asdu.ParamsWide
func (sf *ClientOption) SetParams(p *asdu.Params) *ClientOption {
	if err := p.Valid(); err != nil {
		sf.params = *asdu.ParamsWide
	} else {
		sf.params = *p
	}

	return sf
}

// SetReconnectInterval set tcp  reconnect the host interval when connect failed after try
func (sf *ClientOption) SetReconnectInterval(t time.Duration) *ClientOption {
	if t > 0 {
		sf.reconnectInterval = t
	}

	return sf
}

// SetAutoReconnect enable auto reconnect
func (sf *ClientOption) SetAutoReconnect(b bool) *ClientOption {
	sf.autoReconnect = b

	return sf
}

// SetTLSConfig set tls config
func (sf *ClientOption) SetTLSConfig(t *tls.Config) *ClientOption {
	sf.TLSConfig = t

	return sf
}

// AddRemoteServer adds a broker URI to the list of brokers to be used.
// The format should be scheme://host:port
// Default values for hostname is "127.0.0.1", for schema is "tcp://".
// An example broker URI would look like: tcp://foobar.com:1204
func (sf *ClientOption) AddRemoteServer(server string) error {
	if len(server) > 0 && server[0] == ':' {
		server = "127.0.0.1" + server
	}

	if !strings.Contains(server, "://") {
		server = "tcp://" + server
	}

	remoteURL, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("parse error:%w", err)
	}

	sf.server = remoteURL

	return nil
}
