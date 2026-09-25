package cs104

import (
	"errors"
	"github.com/orglibs/go-iecp5/asdu"
	"net"
	"sync"
	"testing"
	"time"
)

type memoryListener struct {
	connections chan net.Conn
	closed      chan struct{}
	once        sync.Once
}

func newMemoryListener() *memoryListener {
	return &memoryListener{connections: make(chan net.Conn), closed: make(chan struct{})}
}
func (l *memoryListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.connections:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}
func (l *memoryListener) Close() error   { l.once.Do(func() { close(l.closed) }); return nil }
func (l *memoryListener) Addr() net.Addr { return &net.TCPAddr{} }

func TestOwnedListenerCloseRacesServe(t *testing.T) {
	for i := 0; i < 50; i++ {
		srv := NewServer(nil, nil, false)
		listener := newMemoryListener()
		done := make(chan error, 1)
		go func() { done <- srv.Serve(listener) }()
		if err := srv.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, net.ErrClosed) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("Serve failed to stop")
		}
		select {
		case <-listener.closed:
		default:
			t.Fatal("listener leaked")
		}
	}
}

func TestOwnedListenerClosesConnectedSession(t *testing.T) {
	srv := NewServer(nil, nil, false)
	listener := newMemoryListener()
	done := make(chan error, 1)
	connected := make(chan struct{})
	srv.SetOnConnectionHandler(func(asdu.Connect) { close(connected) })
	go func() { done <- srv.Serve(listener) }()
	peer, conn := net.Pipe()
	defer peer.Close()
	select {
	case listener.connections <- conn:
	case <-time.After(time.Second):
		t.Fatal("not accepting")
	}
	select {
	case <-connected:
	case <-time.After(time.Second):
		t.Fatal("not connected")
	}
	closed := make(chan struct{})
	go func() { _ = srv.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close stuck")
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type orderedGIHandler struct {
	ServerHandlerInterface
	t      *testing.T
	called bool
}

func (h *orderedGIHandler) InterrogationHandler(c asdu.Connect, a *asdu.ASDU, q asdu.QualifierOfInterrogation) (asdu.CauseOfTransmission, int) {
	return asdu.CauseOfTransmission{Cause: asdu.ActivationCon}, int(a.CommonAddr)
}
func (h *orderedGIHandler) AfterInterrogationHandler(c asdu.Connect, a *asdu.ASDU, q asdu.QualifierOfInterrogation) error {
	sess := c.(*SrvSession)
	select {
	case raw := <-sess.sendASDU:
		ack := asdu.NewEmptyASDU(asdu.ParamsWide)
		if err := ack.UnmarshalBinary(raw); err != nil {
			h.t.Fatal(err)
		}
		if ack.Coa.Cause != asdu.ActivationCon {
			h.t.Fatal("data preceded ACT_CON")
		}
	default:
		h.t.Fatal("ACT_CON not queued before hook")
	}
	h.called = true
	return nil
}
