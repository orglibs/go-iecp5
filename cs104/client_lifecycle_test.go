package cs104

import (
	"context"
	"io"
	"net"
	"net/url"
	"testing"
	"time"
)

// TestImmediateClose 验证 Start/Close 不依赖后台 goroutine 的调度先后顺序。
func TestImmediateClose(t *testing.T) {
	for i := 0; i < 30; i++ {
		option := NewOption()
		if err := option.AddRemoteServer("tcp://unused:2404"); err != nil {
			t.Fatal(err)
		}
		option.DialContext = func(ctx context.Context, _ *url.URL) (net.Conn, error) { <-ctx.Done(); return nil, ctx.Err() }
		client := NewClient(nil, option)
		if err := client.Start(); err != nil {
			t.Fatal(err)
		}
		_ = client.Close()
		done := make(chan struct{})
		go func() { client.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Close leaked startup/dial goroutine")
		}
		if client.IsConnected() || client.IsActive() {
			t.Fatal("closed client still active")
		}
	}
}

func TestActiveRequiresStartDT(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	option := NewOption()
	_ = option.AddRemoteServer("tcp://unused:2404")
	option.DialContext = func(context.Context, *url.URL) (net.Conn, error) { return clientConn, nil }
	client := NewClient(nil, option)
	connected := make(chan struct{})
	client.SetOnConnectHandler(func(*Client) { close(connected) })
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close(); client.Wait() }()
	select {
	case <-connected:
	case <-time.After(time.Second):
		t.Fatal("client did not connect")
	}
	if !client.IsConnected() || client.IsActive() {
		t.Fatal("TCP connection must not imply active data transfer")
	}
}

func TestReceiveTruncatedFrameStops(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	done := make(chan struct{})
	go func() { receiveLoop(client, make(chan []byte, 1)); close(done) }()
	// 报文宣称还需要控制域，但对端提前关闭，不能对 UnexpectedEOF 空转。
	_, _ = server.Write([]byte{0x68, 4, 0x07})
	_ = server.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("truncated receive loop leaked")
	}
}

func TestReceiveClosedPipeStops(t *testing.T) {
	client, server := net.Pipe()
	_ = client.Close()
	_ = server.Close()
	done := make(chan struct{})
	go func() { receiveLoop(client, make(chan []byte, 1)); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("closed receive loop leaked")
	}
}

func TestReceiveQueueCancellation(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { receiveLoopContext(ctx, client, make(chan []byte)); close(done) }()
	_, err := server.Write([]byte{0x68, 4, 0x07, 0, 0, 0})
	if err != nil && err != io.ErrClosedPipe {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("full receive queue ignored cancellation")
	}
}
