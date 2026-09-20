// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"sync"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

// timeoutResolution is seconds according to companion standard 104,
// subclass 6.9, caption "Definition of time outs". However, ten
// of a second make this system much more responsive i.c.w. S-frames.
const timeoutResolution = 100 * time.Millisecond

const ErrQueueEmpty = "queue empty"

// Server struct type for the common server
type Server struct {
	config    Config
	params    asdu.Params
	handler   ServerHandlerInterface
	queue     ServerQueueInterface
	useQueue  bool
	TLSConfig *tls.Config
	mux       sync.Mutex
	sessions  map[*SrvSession]struct{}

	listen         net.Listener
	onConnection   func(asdu.Connect)
	connectionLost func(asdu.Connect)

	wg           sync.WaitGroup
	stopSessions chan struct{}
}

// NewServer starts a new server instance, default config and default asdu.ParamsWide params are used.
func NewServer(handler ServerHandlerInterface, queue ServerQueueInterface, useQueue bool) *Server {
	server104 := &Server{
		config:       DefaultConfig(),
		params:       *asdu.ParamsWideLocal,
		handler:      handler,
		queue:        queue,
		useQueue:     useQueue,
		sessions:     make(map[*SrvSession]struct{}),
		stopSessions: make(chan struct{}),
	}

	return server104
}

// SetConfig set the server configuration. If configuration is not valid it will use DefaultConfig() values for the server
func (sf *Server) SetConfig(cfg Config) *Server {
	if err := cfg.ValidConfigServer(); err != nil {
		sf.config = DefaultConfig()
	} else {
		sf.config = cfg
	}

	return sf
}

// SetParams set asdu params if params is valid it will use asdu.ParamsWide
func (sf *Server) SetParams(p *asdu.Params) *Server {
	if err := p.Valid(); err != nil {
		sf.params = *asdu.ParamsWide
	} else {
		sf.params = *p
	}

	return sf
}

// ListenAndServer run the server
func (sf *Server) ListenAndServer(addr string) {
	// TLSConfig 原有字段现在实际参与监听；nil 保持普通 TCP。
	var listen net.Listener
	var err error
	if sf.TLSConfig != nil {
		listen, err = tls.Listen("tcp", addr, sf.TLSConfig)
	} else {
		listen, err = net.Listen("tcp", addr)
	}
	if err != nil {
		slog.Error("server run failed", "error", err)

		return
	}

	sf.mux.Lock()
	sf.listen = listen
	sf.mux.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		_ = sf.Close()

		slog.Debug("server stop")
	}()

	slog.Debug("server run")

	go sf.watchActiveSessions()

	for {
		conn, err := listen.Accept()
		if err != nil {
			slog.Error("server run failed", "error", err)

			return
		}

		sf.wg.Add(1)

		slog.Info("new connection accepted", "remote", conn.RemoteAddr().String())

		go func() {
			sess := &SrvSession{
				config:   &sf.config,
				params:   &sf.params,
				handler:  sf.handler,
				queue:    nil,
				useQueue: sf.useQueue,
				conn:     conn,
				rcvASDU:  make(chan []byte, sf.config.RecvUnAckLimitW),
				sendASDU: make(chan []byte, sf.config.SendUnAckLimitK), // <<4),
				rcvRaw:   make(chan []byte, sf.config.RecvUnAckLimitW),
				sendRaw:  make(chan []byte, sf.config.SendUnAckLimitK), // <<5),

				onConnection:   sf.onConnection,
				connectionLost: sf.connectionLost,

				stopSessions: sf.stopSessions,
			}

			if sf.useQueue {
				sess.queue = sf.queue
			}

			sf.mux.Lock()
			sf.sessions[sess] = struct{}{}
			sf.mux.Unlock()
			sess.run(ctx)
			sf.mux.Lock()
			delete(sf.sessions, sess)
			sf.mux.Unlock()
			sf.wg.Done()
		}()
	}
}

// Close close the server
func (sf *Server) Close() error {
	var err error

	sf.mux.Lock()

	if sf.listen != nil {
		err = sf.listen.Close()
		sf.listen = nil
	}

	sf.mux.Unlock()
	sf.wg.Wait()

	return err
}

// sessionSnapshot 缩短持锁时间：网络发送或外部队列调用期间不持有会话表锁。
func (sf *Server) sessionSnapshot() []*SrvSession {
	sf.mux.Lock()
	defer sf.mux.Unlock()
	sessions := make([]*SrvSession, 0, len(sf.sessions))
	for session := range sf.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Send 广播到所有会话并汇总错误。一个会话失败不阻止其他会话尝试发送，
// 但不再静默返回成功；需要等待队列空间时使用 SendWait，避免整批重试产生重复。
func (sf *Server) Send(a *asdu.ASDU) error {
	var failures []error
	sessions := sf.sessionSnapshot()
	if len(sessions) > 0 {
		if _, err := a.MarshalBinary(); err != nil {
			return err
		}
	}
	for _, session := range sessions {
		var frame *asdu.ASDU
		if a != nil {
			frame = a.Clone()
		}
		if err := session.Send(frame); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// Params imp interface Connect
func (sf *Server) Params() *asdu.Params { return &sf.params }

// UnderlyingConn imp interface Connect
func (sf *Server) UnderlyingConn() net.Conn { return nil }

// SetInfoObjTimeZone set info object time zone
func (sf *Server) SetInfoObjTimeZone(zone *time.Location) {
	sf.params.InfoObjTimeZone = zone
}

// SetOnConnectionHandler set on connect handler
func (sf *Server) SetOnConnectionHandler(f func(asdu.Connect)) {
	sf.onConnection = f
}

// SetConnectionLostHandler set connect lost handler
func (sf *Server) SetConnectionLostHandler(f func(asdu.Connect)) {
	sf.connectionLost = f
}

func (sf *Server) watchActiveSessions() {
	for {
		<-sf.stopSessions
		for sess := range sf.sessions {
			if sess.isActive {
				slog.Info("new active session detected, stopping others")
				// sess.sendRaw <- NewUFrame(UStopDtActive)
				sess.isActive = false
				sess.conn.Close()
			}
		}
	}
}
