package cs104

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/url"
	"testing"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

// 全部走真实收发循环，使用内存链路而不是直接调用业务处理器；不依赖监听权限。
func wireWrite(t *testing.T, c net.Conn, raw []byte) {
	t.Helper()
	_ = c.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := c.Write(raw); err != nil {
		t.Fatal(err)
	}
}
func wireRead(t *testing.T, c net.Conn) []byte {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(time.Second))
	head := make([]byte, 2)
	if _, err := io.ReadFull(c, head); err != nil {
		t.Fatal(err)
	}
	rest := make([]byte, int(head[1]))
	if _, err := io.ReadFull(c, rest); err != nil {
		t.Fatal(err)
	}
	return append(head, rest...)
}
func wireSilent(t *testing.T, c net.Conn) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	var b [1]byte
	_, err := c.Read(b[:])
	if e, ok := err.(net.Error); !ok || !e.Timeout() {
		t.Fatalf("expected no response, got %x err=%v", b, err)
	}
}
func wireI(t *testing.T, c net.Conn, send, ack uint16, raw []byte) {
	t.Helper()
	frame, err := NewIFrame(send, ack, raw)
	if err != nil {
		t.Fatal(err)
	}
	wireWrite(t, c, frame)
}

type mirrorHandler struct {
	ServerHandlerInterface
	commands chan asdu.SingleCommandInfo // 仅记录真正抵达应用层的有效控制命令。
}

func (h *mirrorHandler) SingleCommandHandler(_ asdu.Connect, _ *asdu.ASDU, cmd asdu.SingleCommandInfo, _ asdu.InfoObjAddr) asdu.CauseOfTransmission {
	h.commands <- cmd
	return asdu.CauseOfTransmission{Cause: asdu.ActivationCon}
}
func (h *mirrorHandler) ASDUHandler(asdu.Connect, *asdu.ASDU) error { return nil }

func TestServerRejectsMalformedAndMirrorsSelectedTimedCommand(t *testing.T) {
	peer, transport := net.Pipe()
	defer peer.Close()
	handler := &mirrorHandler{commands: make(chan asdu.SingleCommandInfo, 4)}
	cfg := DefaultConfig()
	session := &SrvSession{config: &cfg, params: asdu.ParamsWide, conn: transport, handler: handler, rcvASDU: make(chan []byte, 8), sendASDU: make(chan []byte, 8), rcvRaw: make(chan []byte, 16), sendRaw: make(chan []byte, 16)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); session.run(ctx) }()
	defer func() {
		cancel()
		_ = transport.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("session did not stop")
		}
	}()
	wireWrite(t, peer, []byte{0x68, 4, 7, 0xff, 0xff, 0xff})
	wireSilent(t, peer)
	wireWrite(t, peer, NewUFrame(UStartDtActive))
	if got := wireRead(t, peer); !bytes.Equal(got, NewUFrame(UStartDtConfirm)) {
		t.Fatalf("bad STARTDT reply %x", got)
	}
	// APCI 本身合法但 ASDU 多一个尾字节，不能触发实际遥控。
	wireI(t, peer, 0, 0, []byte{45, 1, 6, 7, 1, 0, 100, 0, 0, 1, 0xaa})
	wireSilent(t, peer)
	select {
	case <-handler.commands:
		t.Fatal("padded command executed")
	default:
	}
	// 异常 STOPDT 不能改变已经激活的链路，否则下面的合法 I 帧无法处理。
	wireWrite(t, peer, []byte{0x68, 4, 0x13, 1, 0, 0})
	wireSilent(t, peer)
	stamp := time.Date(2026, 9, 20, 8, 30, 0, 0, time.UTC)
	raw := append([]byte{58, 1, 6, 7, 1, 0, 100, 0, 0, 0x81}, asdu.CP56Time2a(stamp, time.UTC, true)...)
	wireI(t, peer, 1, 0, raw)
	response := wireRead(t, peer)
	reply := asdu.NewEmptyASDU(asdu.ParamsWide)
	if err := reply.UnmarshalBinary(response[6:]); err != nil {
		t.Fatal(err)
	}
	if reply.OrigAddr != 7 || reply.Coa.Cause != asdu.ActivationCon || !bytes.Equal(reply.InfoObj, raw[6:]) {
		t.Fatalf("selection/value/time not mirrored: %x", response)
	}
	select {
	case cmd := <-handler.commands:
		if !cmd.Qoc.InSelect || !cmd.Time.Equal(stamp) {
			t.Fatal("bad parsed command")
		}
	default:
		t.Fatal("command not dispatched")
	}
}

func TestClientRejectsMalformedActivationAndHonorsK(t *testing.T) {
	peer, transport := net.Pipe()
	defer peer.Close()
	option := NewOption()
	_ = option.AddRemoteServer("tcp://memory:2404")
	cfg := DefaultConfig()
	cfg.SendUnAckLimitK = 1
	cfg.RecvUnAckLimitW = 1
	option.SetConfig(cfg)
	option.DialContext = func(context.Context, *url.URL) (net.Conn, error) { return transport, nil }
	client := NewClient(nil, option)
	client.SetOnConnectHandler(func(c *Client) { c.SendStartDt() })
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close(); client.Wait() }()
	if got := wireRead(t, peer); !bytes.Equal(got, NewUFrame(UStartDtActive)) {
		t.Fatalf("bad STARTDT request %x", got)
	}
	// 用后续 TESTFR 应答作为处理屏障，避免 sleep 后直接读状态导致测试竞态。
	barrier := func() {
		wireWrite(t, peer, NewUFrame(UTestFrActive))
		if got := wireRead(t, peer); !bytes.Equal(got, NewUFrame(UTestFrConfirm)) {
			t.Fatalf("bad TESTFR reply %x", got)
		}
	}
	wireWrite(t, peer, []byte{0x68, 4, 0x0b, 1, 0, 0})
	barrier()
	if client.IsActive() {
		t.Fatal("malformed STARTDT confirmation activated client")
	}
	wireWrite(t, peer, NewUFrame(UStartDtConfirm))
	barrier()
	if !client.IsActive() {
		t.Fatal("valid STARTDT confirmation did not activate client")
	}
	for i := 0; i < 2; i++ {
		if err := client.Send(testSample()); err != nil {
			t.Fatal(err)
		}
	}
	first := wireRead(t, peer)
	if first[2]&1 != 0 {
		t.Fatalf("expected I frame: %x", first)
	}
	wireSilent(t, peer) // k=1，第一帧尚未确认，第二帧必须留在发送队列中。
	wireWrite(t, peer, NewSFrame(1))
	second := wireRead(t, peer)
	if binary.LittleEndian.Uint16(second[2:4])>>1 != 1 {
		t.Fatalf("second frame sequence wrong: %x", second)
	}
}
