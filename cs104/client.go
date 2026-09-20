// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/orglibs/go-iecp5/asdu"
)

const (
	inactive = iota
	active
)

// Client is an IEC104 master
type Client struct {
	option  ClientOption
	conn    net.Conn
	handler ClientHandlerInterface

	// channel
	rcvASDU  chan []byte // for received asdu
	sendASDU chan []byte // for send asdu
	rcvRaw   chan []byte // for recvLoop raw cs104 frame
	sendRaw  chan []byte // for sendLoop raw cs104 frame

	// Transmit and Receive Serial Numbers for I-Frames
	seqNoSend uint16 // sequence number of next outbound I-frame
	ackNoSend uint16 // outbound sequence number yet to be confirmed
	seqNoRcv  uint16 // sequence number of next inbound I-frame
	ackNoRcv  uint16 // inbound sequence number yet to be confirmed

	// maps sendTime I-frames to their respective sequence number
	pending []seqPending

	startDtActiveSendSince atomic.Value // The timeout interval to wait for an acknowledgement when sending startDtActive.
	stopDtActiveSendSince  atomic.Value // When stopDtActive is initiated, the timeout to wait for an acknowledgement of a reply is set.

	// connection status
	status   uint32
	rwMux    sync.RWMutex
	isActive uint32

	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closeCancel context.CancelFunc
	done        chan struct{}

	onConnect        func(c *Client)
	onConnectionLost func(c *Client)
}

// NewClient returns an IEC104 master,default config and default asdu.ParamsWide params
func NewClient(handler ClientHandlerInterface, o *ClientOption) *Client {
	return &Client{
		option:           *o,
		handler:          handler,
		rcvASDU:          make(chan []byte, o.config.RecvUnAckLimitW<<4),
		sendASDU:         make(chan []byte, o.config.SendUnAckLimitK<<4),
		rcvRaw:           make(chan []byte, o.config.RecvUnAckLimitW<<5),
		sendRaw:          make(chan []byte, o.config.SendUnAckLimitK<<5), // may not block!
		onConnect:        func(*Client) {},
		onConnectionLost: func(*Client) {},
	}
}

// SetOnConnectHandler set on connect handler
func (sf *Client) SetOnConnectHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onConnect = f
	}

	return sf
}

// SetConnectionLostHandler set connection lost handler
func (sf *Client) SetConnectionLostHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onConnectionLost = f
	}

	return sf
}

// Start start the server,and return quickly,if it nil,the server will disconnected background,other failed
func (sf *Client) Start() error {
	if sf.option.server == nil {
		return errors.New("empty remote server")
	}

	sf.rwMux.Lock()
	if !atomic.CompareAndSwapUint32(&sf.status, initial, disconnected) {
		sf.rwMux.Unlock()
		return errors.New("client already started")
	}
	ctx, cancel := context.WithCancel(context.Background())
	sf.closeCancel = cancel
	sf.done = make(chan struct{})
	sf.rwMux.Unlock()
	go sf.running(ctx)
	return nil
}

// run the client connection
func (sf *Client) running(ctx context.Context) {
	defer close(sf.done)
	defer sf.setConnectStatus(initial)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		slog.Debug("connecting server", "server", sf.option.server)
		var conn net.Conn
		var err error
		if sf.option.DialContext != nil {
			dialCtx, cancel := context.WithTimeout(ctx, sf.option.config.ConnectTimeout0)
			conn, err = sf.option.DialContext(dialCtx, sf.option.server)
			cancel()
		} else {
			conn, err = openConnectionContext(ctx, sf.option.server, sf.option.TLSConfig, sf.option.config.ConnectTimeout0)
		}
		if err != nil {
			slog.Error("connect failed", "error", err)
			if !sf.option.autoReconnect {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(sf.option.reconnectInterval):
			}

			continue
		}

		slog.Debug("connect success")
		if ctx.Err() != nil {
			_ = conn.Close()
			return
		}
		sf.conn = conn
		sf.run(ctx)

		slog.Debug("disconnected server", "server", sf.option.server)
		select {
		case <-ctx.Done():
			return
		default:
			// Random value in the range of 500ms-1s avoid fast retries that can cause invalid connections to the server.
			rnd, err := rand.Int(rand.Reader, big.NewInt(500))
			if err != nil {
				rnd = big.NewInt(0)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond * time.Duration(500+rnd.Int64())):
			}
		}
	}
}

func (sf *Client) recvLoop() {
	slog.Debug("recvLoop started!")
	receiveLoopContext(sf.ctx, sf.conn, sf.rcvRaw)

	sf.cancel()
	sf.wg.Done()
	slog.Debug("recvLoop stopped!")
}

func (sf *Client) sendLoop() {
	slog.Debug("sendLoop started")

	defer func() {
		sf.cancel()
		sf.wg.Done()
		slog.Debug("sendLoop stopped")
	}()

	for {
		select {
		case <-sf.ctx.Done():
			return
		case apdu := <-sf.sendRaw:
			slog.Debug("TX Raw", "tx", apdu)

			for wrCnt := 0; len(apdu) > wrCnt; {
				byteCount, err := sf.conn.Write(apdu[wrCnt:])
				if err != nil {
					// See: https://github.com/golang/go/issues/4373
					if errors.Is(err, io.EOF) && errors.Is(err, io.ErrClosedPipe) ||
						strings.Contains(err.Error(), "use of closed network connection") {
						slog.Error("sendRaw failed", "error", err)

						return
					}

					//nolint:errorlint
					if e, ok := err.(net.Error); !ok || !e.Timeout() {
						slog.Error("sendRaw failed", "error", err)

						return
					} // temporary error may be recoverable
				}
				wrCnt += byteCount
			}
			sf.traceAPDU(true, apdu)
		}
	}
}

// traceAPDU 只在安装观察器时复制，防止日志回调修改接收队列中的原始协议数据。
func (sf *Client) traceAPDU(outbound bool, apdu []byte) {
	if sf.option.OnAPDU != nil {
		sf.option.OnAPDU(outbound, append([]byte(nil), apdu...))
	}
}

// run is the big fat state machine.
func (sf *Client) run(ctx context.Context) {
	slog.Debug("run started!")
	// before any thing make sure init
	sf.cleanUp()

	sf.rwMux.Lock()
	sf.ctx, sf.cancel = context.WithCancel(ctx)
	sf.rwMux.Unlock()
	sf.setConnectStatus(connected)
	sf.wg.Add(3)
	go sf.recvLoop()
	go sf.sendLoop()
	go sf.handlerLoop()

	checkTicker := time.NewTicker(timeoutResolution)

	// transmission timestamps for timeout calculation
	willNotTimeout := time.Now().Add(time.Hour * 24 * 365 * 100)

	unAckRcvSince := willNotTimeout
	idleTimeout3Sine := time.Now()         // Idle interval initiates testFrAlive
	testFrAliveSendSince := willNotTimeout // The timeout interval to wait for an acknowledgement when initiating a testFrAlive.

	sf.startDtActiveSendSince.Store(willNotTimeout)
	sf.stopDtActiveSendSince.Store(willNotTimeout)

	sendSFrame := func(rcvSN uint16) {
		slog.Debug("TX sFrame", "tx sFrame", SAPCI{rcvSN})
		sf.sendRaw <- NewSFrame(rcvSN)
	}

	sendIFrame := func(asdu1 []byte) {
		seqNo := sf.seqNoSend

		iframe, err := NewIFrame(seqNo, sf.seqNoRcv, asdu1)
		if err != nil {
			return
		}

		sf.ackNoRcv = sf.seqNoRcv
		sf.seqNoSend = (seqNo + 1) & 32767
		sf.pending = append(sf.pending, seqPending{seqNo & 32767, time.Now(), asdu1})

		slog.Debug("TX iFrame", "tx iFrame", IAPCI{seqNo, sf.seqNoRcv})
		sf.sendRaw <- iframe
	}

	defer func() {
		// 无论是取消、协议错误还是超时，都先取消本次会话，再等待后台任务退出。
		sf.cancel()
		// default: STOPDT, when a connection is established and "data transfer" is not enabled.
		atomic.StoreUint32(&sf.isActive, inactive)
		sf.setConnectStatus(disconnected)
		checkTicker.Stop()
		_ = sf.conn.Close() // Chain trigger cancel
		sf.wg.Wait()
		sf.onConnectionLost(sf)
		slog.Debug("run stopped!")
	}()

	sf.onConnect(sf)

	for {
		if atomic.LoadUint32(&sf.isActive) == active && seqNoCount(sf.ackNoSend, sf.seqNoSend) < sf.option.config.SendUnAckLimitK {
			select {
			case o := <-sf.sendASDU:
				sendIFrame(o)
				idleTimeout3Sine = time.Now()

				continue
			case <-sf.ctx.Done():
				return
			default: // make no block
			}
		}

		select {
		case <-sf.ctx.Done():
			return
		case now := <-checkTicker.C:
			// define timeOuts
			timeOut1 := now.Sub(testFrAliveSendSince)
			timeOut2, ok := sf.startDtActiveSendSince.Load().(time.Time)
			if !ok {
				slog.Warn("startDt is not a time.Time type", "startDt", sf.startDtActiveSendSince)
			}

			timeOut3, ok := sf.stopDtActiveSendSince.Load().(time.Time)
			if !ok {
				slog.Warn("stopDt is not a time.Time type", "stopDt", sf.stopDtActiveSendSince)
			}

			// check all timeouts
			if timeOut1 >= sf.option.config.SendUnAckTimeout1 ||
				now.Sub(timeOut2) >= sf.option.config.SendUnAckTimeout1 ||
				now.Sub(timeOut3) >= sf.option.config.SendUnAckTimeout1 {
				slog.Error("test frame alive confirm timeout t₁")

				return
			}

			// check oldest unacknowledged outbound
			oldestTime, err := sf.peek()
			if err != nil {
				slog.Debug("no pending frames found", "error", err)
			}

			if sf.ackNoSend != sf.seqNoSend &&
				now.Sub(oldestTime) >= sf.option.config.SendUnAckTimeout1 {
				sf.ackNoSend++

				slog.Error("fatal transmission timeout t₁")

				return
			}

			// Determine if the earliest i-Frame sent timed out, and reply to the sFrame if it did.
			if sf.ackNoRcv != sf.seqNoRcv &&
				(now.Sub(unAckRcvSince) >= sf.option.config.RecvUnAckTimeout2 ||
					now.Sub(idleTimeout3Sine) >= timeoutResolution) {
				sendSFrame(sf.seqNoRcv)
				sf.ackNoRcv = sf.seqNoRcv
			}

			// Send TestFrActive frame when idle time is up.
			if now.Sub(idleTimeout3Sine) >= sf.option.config.IdleTimeout3 {
				sf.sendUFrame(UTestFrActive)
				testFrAliveSendSince = time.Now()
				idleTimeout3Sine = testFrAliveSendSince
			}

		case apdu := <-sf.rcvRaw:
			sf.traceAPDU(false, apdu)
			apci, asduVal, err := ParseChecked(apdu)
			if err != nil {
				slog.Warn("discard malformed APCI", "error", err)
				continue
			}
			// 只有结构合法的帧才刷新空闲时间；异常输入不能延长链路寿命。
			idleTimeout3Sine = time.Now()
			switch head := apci.(type) {
			case SAPCI:
				slog.Debug("RX sFrame", "rx sFrame", head)
				if !sf.updateAckNoOut(head.RcvSN) {
					slog.Error("fatal incoming acknowledge either earlier than previous or later than sendTime")

					return
				}

			case IAPCI:
				slog.Debug("RX iFrame", "rx iFrame", head)
				if atomic.LoadUint32(&sf.isActive) == inactive {
					slog.Warn("station not active")

					break // not active, discard apdu
				}

				if !sf.updateAckNoOut(head.RcvSN) || head.SendSN != sf.seqNoRcv {
					slog.Error("fatal incoming acknowledge either earlier than previous or later than sendTime")

					return
				}

				select {
				case sf.rcvASDU <- asduVal:
				case <-sf.ctx.Done():
					return
				}
				if sf.ackNoRcv == sf.seqNoRcv { // first unacked
					unAckRcvSince = time.Now()
				}

				sf.seqNoRcv = (sf.seqNoRcv + 1) & 32767
				if seqNoCount(sf.ackNoRcv, sf.seqNoRcv) >= sf.option.config.RecvUnAckLimitW {
					sendSFrame(sf.seqNoRcv)
					sf.ackNoRcv = sf.seqNoRcv
				}

			case UAPCI:
				slog.Debug("RX uFrame", "rx", head)
				switch head.Function {
				// case uStartDtActive:
				//	sf.sendUFrame(uStartDtConfirm)
				//	atomic.StoreUint32(&sf.isActive, active)
				case UStartDtConfirm:
					atomic.StoreUint32(&sf.isActive, active)
					sf.startDtActiveSendSince.Store(willNotTimeout)
				// case uStopDtActive:
				//	sf.sendUFrame(uStopDtConfirm)
				//	atomic.StoreUint32(&sf.isActive, inactive)
				case UStopDtConfirm:
					atomic.StoreUint32(&sf.isActive, inactive)
					sf.stopDtActiveSendSince.Store(willNotTimeout)
				case UTestFrActive:
					sf.sendUFrame(UTestFrConfirm)
				case UTestFrConfirm:
					testFrAliveSendSince = willNotTimeout
				default:
					slog.Error("illegal U-Frame functions ignored", "illegal uFrame", head.Function)
				}
			}
		}
	}
}

func (sf *Client) handlerLoop() {
	slog.Debug("handlerLoop started")

	defer func() {
		sf.wg.Done()
		slog.Debug("handlerLoop stopped")
	}()

	for {
		select {
		case <-sf.ctx.Done():
			return
		case rawAsdu := <-sf.rcvASDU:
			asduPack := asdu.NewEmptyASDU(&sf.option.params)
			if err := asduPack.UnmarshalBinary(rawAsdu); err != nil {
				slog.Warn("asdu UnmarshalBinary failed", "error", err)

				continue
			}
			if err := sf.clientHandler(asduPack); err != nil {
				slog.Warn("Falied handling I frame, error:", "error", err)
			}
		}
	}
}

func (sf *Client) setConnectStatus(status uint32) {
	sf.rwMux.Lock()
	atomic.StoreUint32(&sf.status, status)
	sf.rwMux.Unlock()
}

func (sf *Client) connectStatus() uint32 {
	sf.rwMux.RLock()
	status := atomic.LoadUint32(&sf.status)
	sf.rwMux.RUnlock()

	return status
}

func (sf *Client) cleanUp() {
	sf.ackNoRcv = 0
	sf.ackNoSend = 0
	sf.seqNoRcv = 0
	sf.seqNoSend = 0
	sf.pending = nil
	// clear sending chan buffer
loop:
	for {
		select {
		case <-sf.sendRaw:
		case <-sf.rcvRaw:
		case <-sf.rcvASDU:
		case <-sf.sendASDU:
		default:
			break loop
		}
	}
}

func (sf *Client) sendUFrame(which byte) {
	slog.Debug("TX uFrame", "tx uFrame", UAPCI{which})
	sf.sendRaw <- NewUFrame(which)
}

func (sf *Client) updateAckNoOut(ackNo uint16) (ok bool) {
	if ackNo == sf.ackNoSend {
		return true
	}

	// new acks validate， ack cannot precede req seq, error.
	if seqNoCount(sf.ackNoSend, sf.seqNoSend) < seqNoCount(ackNo, sf.seqNoSend) {
		return false
	}

	// confirm reception
	for i, v := range sf.pending {
		if v.seq == (ackNo-1)&32767 {
			sf.pending = sf.pending[i+1:]

			break
		}
	}

	sf.ackNoSend = ackNo

	return true
}

// IsConnected get server session connected state
func (sf *Client) IsConnected() bool {
	return sf.connectStatus() == connected
}

// clientHandler hand response handler
func (sf *Client) clientHandler(asduPack *asdu.ASDU) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("client handler", "error", err)
		}
	}()

	slog.Debug("ASDU", "asdu", asduPack)

	switch asduPack.Identifier.Type {
	case asdu.C_IC_NA_1: // InterrogationCmd
		err := sf.handler.InterrogationHandler(sf, asduPack)

		return err

	case asdu.C_CI_NA_1: // CounterInterrogationCmd
		err := sf.handler.CounterInterrogationHandler(sf, asduPack)

		return err

	case asdu.C_RD_NA_1: // ReadCmd
		err := sf.handler.ReadHandler(sf, asduPack)

		return err

	case asdu.C_CS_NA_1: // ClockSynchronizationCmd
		err := sf.handler.ClockSyncHandler(sf, asduPack)

		return err

	case asdu.C_TS_NA_1: // TestCommand
		err := sf.handler.TestCommandHandler(sf, asduPack)

		return err

	case asdu.C_RP_NA_1: // ResetProcessCmd
		err := sf.handler.ResetProcessHandler(sf, asduPack)

		return err

	case asdu.C_CD_NA_1: // DelayAcquireCommand
		err := sf.handler.DelayAcquisitionHandler(sf, asduPack)

		return err
	default:
		err := sf.handler.ASDUHandler(sf, asduPack)

		return err
	}
}

// Params returns params of client
func (sf *Client) Params() *asdu.Params {
	return &sf.option.params
}

// Send send asdu
func (sf *Client) Send(a *asdu.ASDU) error {
	if !sf.IsConnected() {
		return ErrUseClosedConnection
	}

	if atomic.LoadUint32(&sf.isActive) == inactive {
		return ErrNotActive
	}

	data, err := a.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error in data:%w", err)
	}

	select {
	case sf.sendASDU <- data:
	default:
		return ErrBufferFull
	}

	return nil
}

func (sf *Client) SendQueuedASDU(a *asdu.ASDU) error {
	return nil
}

// UnderlyingConn returns underlying conn of client
func (sf *Client) UnderlyingConn() net.Conn {
	return sf.conn
}

func (sf *Client) IsActive() bool {
	return sf.IsConnected() && atomic.LoadUint32(&sf.isActive) == active
}

func (sf *Client) AreAllMessagesConfirmed() bool {
	return true
}

func (sf *Client) RemoteClose() error {
	return nil
}

// Close close all
func (sf *Client) Close() error {
	sf.rwMux.Lock()

	if sf.closeCancel != nil {
		sf.closeCancel()
	}

	sf.rwMux.Unlock()

	return nil
}

// Wait waits for shutdown. Do not call from a client callback.
func (sf *Client) Wait() {
	sf.rwMux.RLock()
	done := sf.done
	sf.rwMux.RUnlock()
	if done != nil {
		<-done
	}
}

// Disconnect abandons an ambiguous command session, allowing automatic reconnect.
func (sf *Client) Disconnect() {
	sf.rwMux.Lock()
	if sf.cancel != nil {
		sf.cancel()
	}
	sf.rwMux.Unlock()
}

// SendStartDt start data transmission on this connection
func (sf *Client) SendStartDt() {
	sf.startDtActiveSendSince.Store(time.Now())
	sf.sendUFrame(UStartDtActive)
}

// SendStopDt stop data transmission on this connection
func (sf *Client) SendStopDt() {
	sf.stopDtActiveSendSince.Store(time.Now())
	sf.sendUFrame(UStopDtActive)
}

// InterrogationCmd wrap asdu.InterrogationCmd
func (sf *Client) InterrogationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qoi asdu.QualifierOfInterrogation) error {
	err := asdu.InterrogationCmd(sf, coa, ca, qoi)
	if err != nil {
		return err
	}

	return nil
}

// CounterInterrogationCmd wrap asdu.CounterInterrogationCmd
func (sf *Client) CounterInterrogationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qcc asdu.QualifierCountCall) error {
	err := asdu.CounterInterrogationCmd(sf, coa, ca, qcc)
	if err != nil {
		return err
	}

	return nil
}

// ReadCmd wrap asdu.ReadCmd
func (sf *Client) ReadCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, ioa asdu.InfoObjAddr) error {
	err := asdu.ReadCmd(sf, coa, ca, ioa)
	if err != nil {
		return err
	}

	return nil
}

// ClockSynchronizationCmd wrap asdu.ClockSynchronizationCmd
func (sf *Client) ClockSynchronizationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, t time.Time) error {
	err := asdu.ClockSynchronizationCmd(sf, coa, ca, true, t)
	if err != nil {
		return err
	}

	return nil
}

// ResetProcessCmd wrap asdu.ResetProcessCmd
func (sf *Client) ResetProcessCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qrp asdu.QualifierOfResetProcessCmd) error {
	err := asdu.ResetProcessCmd(sf, coa, ca, qrp)
	if err != nil {
		return err
	}

	return nil
}

// DelayAcquireCommand wrap asdu.DelayAcquireCommand
func (sf *Client) DelayAcquireCommand(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, msec uint16) error {
	err := asdu.DelayAcquireCommand(sf, coa, ca, msec)
	if err != nil {
		return err
	}

	return nil
}

// TestCommand  wrap asdu.TestCommand
func (sf *Client) TestCommand(coa asdu.CauseOfTransmission, ca asdu.CommonAddr) error {
	err := asdu.TestCommand(sf, coa, ca)
	if err != nil {
		return err
	}

	return nil
}

func (sf *Client) peek() (time.Time, error) {
	if len(sf.pending) > 0 {
		return sf.pending[0].sendTime, nil
	}

	return time.Time{}, errors.New("no pending i-frame")
}
