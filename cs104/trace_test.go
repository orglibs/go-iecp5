package cs104

import (
	"context"
	"net"
	"net/url"
	"testing"
)

// 观察器拿到副本，即使修改/保留报文也不能干扰链路状态；TCP 分段不能产生多条帧日志。
func TestAPDUObserverOwnsSnapshot(t *testing.T) {
	peer, transport := net.Pipe()
	defer peer.Close()
	option := NewOption().SetAutoReconnect(false)
	_ = option.AddRemoteServer("tcp://memory:2404")
	option.DialContext = func(context.Context, *url.URL) (net.Conn, error) { return transport, nil }
	type event struct {
		outbound bool   // true 为实际写出的完整帧，false 为进入状态机的完整帧。
		raw      []byte // 保留回调给出的独立快照，以检查后续帧是否覆盖它。
	}
	events := make(chan event, 16)
	option.OnAPDU = func(outbound bool, raw []byte) {
		raw[0] = 0xaa // 修改快照；真实发送帧和接收解析仍必须使用 0x68。
		events <- event{outbound, raw}
	}
	client := NewClient(nil, option)
	client.SetOnConnectHandler(func(c *Client) { c.SendStartDt() })
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close(); client.Wait() }()
	if raw := wireRead(t, peer); raw[0] != 0x68 {
		t.Fatal("observer changed transmitted bytes")
	}
	confirm := NewUFrame(UStartDtConfirm)
	wireWrite(t, peer, confirm[:1])
	wireWrite(t, peer, confirm[1:3])
	wireWrite(t, peer, confirm[3:])
	// TESTFR 确认作为屏障，确保激活确认已经处理完。
	wireWrite(t, peer, NewUFrame(UTestFrActive))
	wireRead(t, peer)
	if !client.IsActive() {
		t.Fatal("observer changed received activation frame")
	}
	_ = client.Close()
	client.Wait()
	close(events)
	var retained []event
	for e := range events {
		retained = append(retained, e)
	}
	if len(retained) != 4 {
		t.Fatalf("expected 4 whole-frame events, got %d", len(retained))
	}
	counts := map[byte]int{}
	for _, e := range retained {
		if len(e.raw) != 6 || e.raw[0] != 0xaa {
			t.Fatalf("snapshot reused: %x", e.raw)
		}
		counts[e.raw[2]]++
		if (e.raw[2] == 7 || e.raw[2] == 0x83) != e.outbound {
			t.Fatalf("wrong direction: %+v", e)
		}
	}
	for _, code := range []byte{7, 0x0b, 0x43, 0x83} {
		if counts[code] != 1 {
			t.Fatalf("missing/duplicate frame %x: %v", code, counts)
		}
	}
}
