package cs104

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

func testSample() *asdu.ASDU {
	a := asdu.NewASDU(asdu.ParamsWide, asdu.Identifier{Type: asdu.M_SP_NA_1, Variable: asdu.VariableStruct{Number: 1}, Coa: asdu.CauseOfTransmission{Cause: asdu.Spontaneous}, CommonAddr: 1})
	a.InfoObj = append(a.InfoObj, 100, 0, 0, 1)
	return a
}

// 使用真实 SrvSession 的有界发送队列，验证发送背压而无需开启 TCP 端口。
func testBufferedSession(size int) *SrvSession {
	return &SrvSession{params: asdu.ParamsWide, status: connected, sendASDU: make(chan []byte, size)}
}

func TestBroadcastReportsFailures(t *testing.T) {
	slow, fast := testBufferedSession(1), testBufferedSession(2)
	slow.sendASDU <- []byte{0}
	srv := NewServer(nil, nil, false)
	srv.sessions[slow] = struct{}{}
	srv.sessions[fast] = struct{}{}
	if err := srv.Send(testSample()); !errors.Is(err, ErrBufferFull) {
		t.Fatalf("broadcast hid failure: %v", err)
	}
	if len(fast.sendASDU) != 1 {
		t.Fatal("one failed session prevented delivery to the other")
	}
}

func TestBroadcastWaitDoesNotDuplicate(t *testing.T) {
	slow, fast := testBufferedSession(1), testBufferedSession(2)
	slow.sendASDU <- []byte{0}
	srv := NewServer(nil, nil, false)
	srv.sessions[slow] = struct{}{}
	srv.sessions[fast] = struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 在下一次重试前释放慢会话空间，快速会话只能收到一次广播。
	drained := make(chan struct{})
	go func() { defer close(drained); time.Sleep(20 * time.Millisecond); <-slow.sendASDU }()
	if err := srv.SendWait(ctx, testSample()); err != nil {
		t.Fatal(err)
	}
	<-drained
	if len(slow.sendASDU) != 1 || len(fast.sendASDU) != 1 {
		t.Fatalf("slow=%d fast=%d", len(slow.sendASDU), len(fast.sendASDU))
	}
}

func TestWaitingTimeoutAndCancellation(t *testing.T) {
	full := testBufferedSession(1)
	full.sendASDU <- []byte{0}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := Waiting(ctx, full).Send(testSample()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wrong timeout error: %v", err)
	}
	empty := testBufferedSession(1)
	canceled, stop := context.WithCancel(context.Background())
	stop()
	if err := Waiting(canceled, empty).Send(testSample()); !errors.Is(err, context.Canceled) {
		t.Fatalf("wrong cancellation error: %v", err)
	}
	if len(empty.sendASDU) != 0 {
		t.Fatal("canceled call still enqueued")
	}
}

type rejectingConn struct {
	asdu.Connect
	calls int
}

func (c *rejectingConn) Send(*asdu.ASDU) error { c.calls++; return ErrUseClosedConnection }
func TestWaitingDoesNotRetryDisconnect(t *testing.T) {
	conn := new(rejectingConn)
	if err := Waiting(context.Background(), conn).Send(testSample()); !errors.Is(err, ErrUseClosedConnection) || conn.calls != 1 {
		t.Fatalf("retried permanent failure: %v count=%d", err, conn.calls)
	}
}

// 发送前的严格校验覆盖主站、普通从站、外部队列从站三条路径。
type recordingQueue struct{ count int }

func (q *recordingQueue) Enqueue(asdu.ASDU) error     { q.count++; return nil }
func (q *recordingQueue) ReEnqueue(asdu.ASDU) error   { return nil }
func (q *recordingQueue) Dequeue() (asdu.ASDU, error) { return asdu.ASDU{}, errors.New("empty") }
func TestSendRejectsMalformedBeforeEnqueue(t *testing.T) {
	a := testSample()
	a.Variable.Number = 2
	client := NewClient(nil, NewOption())
	atomic.StoreUint32(&client.status, connected)
	atomic.StoreUint32(&client.isActive, active)
	if err := client.Send(a); !errors.Is(err, asdu.ErrInfoObjSizeMismatch) || len(client.sendASDU) != 0 {
		t.Fatalf("client accepted bad ASDU: %v", err)
	}
	session := testBufferedSession(1)
	if err := session.Send(a); !errors.Is(err, asdu.ErrInfoObjSizeMismatch) || len(session.sendASDU) != 0 {
		t.Fatalf("server accepted bad ASDU: %v", err)
	}
	queue := new(recordingQueue)
	session.useQueue = true
	session.queue = queue
	if err := session.Send(a); !errors.Is(err, asdu.ErrInfoObjSizeMismatch) || queue.count != 0 {
		t.Fatalf("queued bad ASDU: %v", err)
	}
	a = testSample()
	a.InfoObj = make([]byte, 250)
	if err := client.Send(a); !errors.Is(err, asdu.ErrLengthOutOfRange) {
		t.Fatalf("oversized ASDU accepted: %v", err)
	}
}
