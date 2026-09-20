package cs104

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestSessionActivationAndWatcherShutdown(t *testing.T) {
	srv := NewServer(nil, nil, false)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); srv.watchActiveSessions(ctx) }()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("session watcher did not stop")
		}
	}()

	var sessions []*SrvSession
	for i := 0; i < 2; i++ {
		conn, peer := net.Pipe()
		defer conn.Close()
		defer peer.Close()
		session := &SrvSession{conn: conn, status: connected}
		sessions = append(sessions, session)
		srv.sessions[session] = struct{}{}
	}

	// Activation requests may arrive together. Exactly one session must win.
	var workers sync.WaitGroup
	for _, session := range sessions {
		workers.Add(1)
		go func() {
			defer workers.Done()
			activated := make(chan struct{})
			select {
			case srv.stopSessions <- activationRequest{session, activated}:
			case <-ctx.Done():
				return
			}
			select {
			case <-activated:
			case <-ctx.Done():
			}
		}()
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		for i := 0; i < 100; i++ {
			temporary := &SrvSession{}
			srv.mux.Lock()
			srv.sessions[temporary] = struct{}{}
			srv.mux.Unlock()
			_ = sessions[0].IsActive()
			srv.mux.Lock()
			delete(srv.sessions, temporary)
			srv.mux.Unlock()
		}
	}()
	completed := make(chan struct{})
	go func() { workers.Wait(); close(completed) }()
	select {
	case <-completed:
	case <-time.After(time.Second):
		t.Fatal("activation blocked")
	}
	if sessions[0].IsActive() == sessions[1].IsActive() {
		t.Fatal("exactly one session must remain active")
	}
	for _, session := range sessions {
		if !session.IsActive() {
			_ = session.conn.SetWriteDeadline(time.Now().Add(time.Second))
			if _, err := session.conn.Write([]byte{0}); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatal("previous active connection was not closed")
			}
		}
	}
}
