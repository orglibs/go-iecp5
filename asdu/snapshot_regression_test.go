package asdu

import (
	"bytes"
	"testing"
	"time"
)

// MarshalBinary 不能借用可变 Bootstrap 作为发送队列的数据，否则复用 ASDU 会篡改旧帧。
func TestMarshalOwnsPayloadSnapshot(t *testing.T) {
	a := newTestASDU(M_SP_NA_1, false, 1, []byte{100, 0, 0, 1})
	first, err := a.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first[6:], a.InfoObj) {
		t.Fatal("external InfoObj not copied")
	}
	saved := append([]byte(nil), first...)
	a.InfoObj[3] = 0
	a.CommonAddr = 2
	if _, err := a.MarshalBinary(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, saved) {
		t.Fatal("later marshal changed queued frame")
	}
}

func TestTimeEncodingUsesTargetZone(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	// 同一瞬间在 UTC 是周日，在柏林已经是周一且采用夏令时。
	moment := time.Date(2026, 7, 5, 23, 30, 0, 0, time.UTC)
	raw := CP56Time2a(moment, berlin, true)
	if raw[3]&0x80 == 0 || raw[4]>>5 != 1 {
		t.Fatalf("wrong target-zone SU/day-of-week: %x", raw)
	}
	invalid := CP56Time2a(moment, berlin, false)
	if invalid[2]&0x80 == 0 || !ParseCP56Time2a(invalid, berlin).IsZero() {
		t.Fatal("circutor invalid-time flag lost")
	}
}

// 保留现有接口，只修正带时标测试命令的 COT；不影响 IsTest/IsNegative 标记。
type captureTestCommand struct {
	Connect
	frame *ASDU
}

func (c *captureTestCommand) Params() *Params    { return ParamsWide }
func (c *captureTestCommand) Send(a *ASDU) error { c.frame = a; return nil }
func TestTimedTestCommandForcesActivation(t *testing.T) {
	for _, cause := range []Cause{Request, Spontaneous, ActivationCon, UnknownCA} {
		c := new(captureTestCommand)
		if err := TestCommandCP56Time2a(c, CauseOfTransmission{Cause: cause, IsTest: true}, 1, time.Now()); err != nil {
			t.Fatal(err)
		}
		if c.frame.Coa.Cause != Activation || !c.frame.Coa.IsTest {
			t.Fatalf("bad COT %+v", c.frame.Coa)
		}
		// 验证发送和接收路径均认识带时标测试命令，不能只检查构造出的结构体。
		raw, err := c.frame.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		decoded := NewEmptyASDU(ParamsWide)
		if err := decoded.UnmarshalBinary(raw); err != nil {
			t.Fatal(err)
		}
	}
}

// 类型 38 的 SEP 不带额外 QDP，信息元素应为 SEP + CP16 + CP56。
// 原来的 11 字节定义会拒绝符合协议的保护事件，或误接收多余尾字节。
func TestProtectionEventCP56Length(t *testing.T) {
	raw := []byte{38, 1, 3, 0, 1, 0, 100, 0, 0, 1, 10, 0, 0, 0, 30, 8, 20, 9, 26}
	a := NewEmptyASDU(ParamsWide)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatal(err)
	}
	encoded, err := a.MarshalBinary()
	if err != nil || !bytes.Equal(encoded, raw) {
		t.Fatalf("protection event changed: %x, %v", encoded, err)
	}
	if err := a.UnmarshalBinary(append(raw, 0)); err == nil {
		t.Fatal("padded protection event accepted")
	}
}
