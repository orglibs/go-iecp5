// 发送背压处理移植自 github.com/riclolsen/go-iecp5 v0.4.4，适配 circutor 接口。
package cs104

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

// waitingConn 只替换一个会话的 Send，其余 circutor Connect 方法委托原连接。
type waitingConn struct {
	asdu.Connect
	// ctx 必须由调用方设置截止时间，防止对端不确认时永久等待。
	ctx context.Context
}

// Waiting 为单个会话提供可取消的发送队列等待，适合大点表总召。
// 仅 ErrBufferFull 可重试，断线和编码错误立即返回。广播必须使用 Server.SendWait，
// 不能用循环重试 Server.Send，否则会重复发送给已经入队成功的会话。
// circutor 的 Server 本身不实现 Connect，因此此处保留会话级接口而不强行扩展 Server。
func Waiting(ctx context.Context, c asdu.Connect) asdu.Connect {
	return waitingConn{Connect: c, ctx: ctx}
}

func (w waitingConn) Send(a *asdu.ASDU) error {
	for {
		// 即使队列尚有空间，也不能在取消后继续接收新的发送请求。
		if err := w.ctx.Err(); err != nil {
			return err
		}
		err := w.Connect.Send(a)
		if !errors.Is(err, ErrBufferFull) {
			return err
		}
		timer := time.NewTimer(2 * time.Millisecond)
		select {
		case <-w.ctx.Done():
			timer.Stop()
			return fmt.Errorf("send buffer full: %w", w.ctx.Err())
		case <-timer.C:
		}
	}
}

// SendWait 对调用时的会话快照逐一发送。只重试未入队的会话，保证同次调用中
// 已接收成功的主站不会收到重复数据；失败结果支持 errors.Is 检查原因。
// Send/SendWait 成功都只表示进入发送队列，不代表远端已完成链路或业务确认。
func (sf *Server) SendWait(ctx context.Context, a *asdu.ASDU) error {
	sessions := sf.sessionSnapshot()
	if len(sessions) > 0 {
		if _, err := a.MarshalBinary(); err != nil {
			return err
		}
	}
	var failures []error
	for _, session := range sessions {
		var frame *asdu.ASDU
		if a != nil {
			frame = a.Clone()
		}
		if err := Waiting(ctx, session).Send(frame); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
