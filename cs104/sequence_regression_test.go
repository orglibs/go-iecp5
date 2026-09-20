package cs104

import "testing"

// 发送序号属于 15 位环，而非 uint16 的 16 位环。确认号 0 必须清除序号 32767。
// 同时验证跨回绕的批量确认及非法确认，防止只修一个边界却误清其他待确认帧。
func TestAcknowledgementWrap(t *testing.T) {
	for _, tc := range []struct {
		name                string
		ack, next, received uint16
		pending             []uint16
		want                int
		valid               bool
	}{
		{"wrap", 32767, 0, 0, []uint16{32767}, 0, true},
		{"cross wrap", 32766, 2, 1, []uint16{32766, 32767, 0, 1}, 1, true},
		{"duplicate", 32767, 0, 32767, []uint16{32767}, 1, true},
		{"future", 32767, 0, 1, []uint16{32767}, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pending := make([]seqPending, len(tc.pending))
			for i, seq := range tc.pending {
				pending[i].seq = seq
			}
			client := &Client{ackNoSend: tc.ack, seqNoSend: tc.next, pending: pending}
			if ok := client.updateAckNoOut(tc.received); ok != tc.valid || len(client.pending) != tc.want {
				t.Fatalf("client valid=%v pending=%d", ok, len(client.pending))
			}
			server := &SrvSession{ackNoSend: tc.ack, seqNoSend: tc.next, pending: pending}
			if ok := server.updateAckNoOut(tc.received); ok != tc.valid || len(server.pending) != tc.want {
				t.Fatalf("server valid=%v pending=%d", ok, len(server.pending))
			}
		})
	}
}

func TestParseCompatibilityRejectsMalformed(t *testing.T) {
	for _, raw := range [][]byte{nil, {0x68, 4, 7}, {0x68, 4, 7, 0xff, 0xff, 0xff}, {0x69, 4, 7, 0, 0, 0}} {
		head, data := Parse(raw)
		if head != nil || data != nil {
			t.Fatalf("legacy Parse accepted %x", raw)
		}
	}
}
