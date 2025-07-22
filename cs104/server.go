// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
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
	config         Config
	params         asdu.Params
	handler        ServerHandlerInterface
	qm             ServerQueueManagerInterface
	useQueue       bool
	TLSConfig      *tls.Config
	mux            sync.Mutex
	sessions       map[*SrvSession]struct{}
	listen         net.Listener
	onConnection   func(asdu.Connect)
	connectionLost func(asdu.Connect)

	wg sync.WaitGroup
}

// NewServer starts a new server instance, default config and default asdu.ParamsWide params are used.
func NewServer(handler ServerHandlerInterface, qm ServerQueueManagerInterface, useQueue bool) *Server {
	server104 := &Server{
		config:   DefaultConfig(),
		params:   *asdu.ParamsWide,
		handler:  handler,
		qm:       qm,
		useQueue: useQueue,
		sessions: make(map[*SrvSession]struct{}),
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
	listen, err := net.Listen("tcp", addr)
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

	for {
		conn, err := listen.Accept()
		if err != nil {
			slog.Error("server run failed", "error", err)

			return
		}

		sf.wg.Add(1)

		go func() {
			sess := &SrvSession{
				config:   &sf.config,
				params:   &sf.params,
				handler:  sf.handler,
				queue:    nil,
				useQueue: sf.useQueue,
				conn:     conn,
				rcvASDU:  make(chan []byte, sf.config.RecvUnAckLimitW<<4),
				sendASDU: make(chan []byte, sf.config.SendUnAckLimitK<<4),
				rcvRaw:   make(chan []byte, sf.config.RecvUnAckLimitW<<5),
				sendRaw:  make(chan []byte, sf.config.SendUnAckLimitK<<5), // may not block!

				onConnection:   sf.onConnection,
				connectionLost: sf.connectionLost,
			}

			if sf.useQueue {
				sess.queue = sf.qm.GetQueue()
				go sess.processQueue()
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

// Send imp interface Connect
func (sf *Server) Send(a *asdu.ASDU) error {
	sf.mux.Lock()

	for k := range sf.sessions {
		_ = k.Send(a.Clone())
	}

	sf.mux.Unlock()

	return nil
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
