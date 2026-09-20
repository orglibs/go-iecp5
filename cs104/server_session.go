// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

//nolint:lll
package cs104

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

const (
	initial uint32 = iota
	disconnected
	connected
)

// SrvSession the cs104 server session
type SrvSession struct {
	config   *Config
	params   *asdu.Params
	conn     net.Conn
	handler  ServerHandlerInterface
	queue    ServerQueueInterface
	useQueue bool

	rcvASDU  chan []byte // for received asdu
	sendASDU chan []byte // for send asdu
	rcvRaw   chan []byte // for recvLoop raw cs104 frame
	sendRaw  chan []byte // for sendLoop raw cs104 frame

	// see subclass 5.1 — Protection against loss and duplication of messages
	seqNoSend uint16 // sequence number of next outbound I-frame
	ackNoSend uint16 // outbound sequence number yet to be confirmed
	seqNoRcv  uint16 // sequence number of next inbound I-frame
	ackNoRcv  uint16 // inbound sequence number yet to be confirmed
	// maps sendTime I-frames to their respective sequence number
	pending []seqPending
	// seqManage

	status uint32
	rwMux  sync.RWMutex

	onConnection   func(asdu.Connect)
	connectionLost func(asdu.Connect)

	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context

	isActive              bool // connection is active after startdt activated
	stopSessions          chan struct{}
	stopDtResponseWaiting bool
}

// RecvLoop feeds t.rcvRaw.
func (sf *SrvSession) recvLoop() {
	slog.Debug("recvLoop started!")
	receiveLoopContext(sf.ctx, sf.conn, sf.rcvRaw)

	sf.cancel()
	sf.wg.Done()
	slog.Debug("recvLoop stopped!")
}

// sendLoop drains t.sendTime.
//
//nolint:nestif
func (sf *SrvSession) sendLoop() {
	slog.Debug("sendLoop started!")

	defer func() {
		sf.cancel()
		sf.wg.Done()
		slog.Debug("sendLoop stopped!")
	}()

	for {
		select {
		case <-sf.ctx.Done():
			if sf.useQueue {
				sf.emptyChannel(sf.sendASDU)
				sf.emptyChannel(sf.sendRaw)
				if len(sf.pending) > 0 {
					var asduList []asdu.ASDU
					for _, pend := range sf.pending {
						asduPack := asdu.NewEmptyASDU(sf.params)
						if err := asduPack.UnmarshalBinary(pend.asduPending); err != nil {
							slog.Error("trying to requeue unconfirmed asdu failed", "error", err)

							continue
						}

						asduList = append(asduList, *asduPack)
					}

					for i := len(asduList) - 1; i >= 0; i-- {
						err := sf.queue.ReEnqueue(asduList[i])
						if err != nil {
							slog.Error("requeue unconfirmed asdu failed", "error", err)
						}
					}
				}
			}

			return
		case apdu := <-sf.sendRaw:
			slog.Debug("TX Raw", "tx", apdu)

			for wrCnt := 0; len(apdu) > wrCnt; {
				byteCount, err := sf.conn.Write(apdu[wrCnt:])
				if err != nil {
					// See: https://github.com/golang/go/issues/4373
					if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe) ||
						strings.Contains(err.Error(), "use of closed network connection") {
						slog.Error("sendRaw failed", "error", err)

						return
					}

					//nolint:errorlint
					if e, ok := err.(net.Error); !ok || !e.Timeout() {
						slog.Error("sendRaw failed", "error", err)

						if sf.useQueue {
							asduPack := asdu.NewEmptyASDU(sf.params)
							if err := asduPack.UnmarshalBinary(apdu); err != nil {
								slog.Error("trying to resend unconfirmed asdu failed", "error", err)

								continue
							}

							err = sf.queue.ReEnqueue(*asduPack)
							if err != nil {
								slog.Error("enqueue unconfirmed asdu failed", "error", err)
							}
						}

						return
					}
				}
				wrCnt += byteCount
			}
		}
	}
}

// run is the big fat state machine.
func (sf *SrvSession) run(ctx context.Context) {
	slog.Debug("run started!")
	// before any thing make sure init
	sf.cleanUp()

	sf.ctx, sf.cancel = context.WithCancel(ctx)
	sf.setConnectStatus(connected)
	sf.wg.Add(3)

	go sf.recvLoop()
	go sf.sendLoop()
	go sf.handlerLoop()

	if sf.useQueue {
		go sf.processQueue()
	}

	// default: STOPDT, when connected establish and not enable "data transfer" yet
	sf.isActive = false
	sf.pending = make([]seqPending, 0)
	checkTicker := time.NewTicker(timeoutResolution)

	// transmission timestamps for timeout calculation
	willNotTimeout := time.Now().Add(time.Hour * 24 * 365 * 100)

	unAckRcvSince := willNotTimeout
	idleTimeout3Sine := time.Now()         // Idle interval to initiate testFrAlive
	testFrAliveSendSince := willNotTimeout // The timeout interval to wait for an acknowledgement when initiating a testFrAlive.
	// For server side, no corresponding U-Frame is required No judgement required
	// var startDtActiveSendSince = willNotTimeout
	// var stopDtActiveSendSince = willNotTimeout

	sendSFrame := func(rcvSN uint16) {
		slog.Debug("TX sFrame", "tx sFrame", SAPCI{rcvSN})
		sf.sendRaw <- NewSFrame(rcvSN)
	}

	sendUFrame := func(which byte) {
		slog.Debug("TX uFrame", "tx uFrame", UAPCI{which})
		sf.sendRaw <- NewUFrame(which)
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

	if sf.onConnection != nil {
		sf.onConnection(sf)
	}

	defer func() {
		sf.cancel()
		sf.setConnectStatus(disconnected)
		checkTicker.Stop()
		_ = sf.conn.Close() // Chain trigger cancel
		sf.wg.Wait()
		if sf.connectionLost != nil {
			sf.connectionLost(sf)
		}

		slog.Debug("run stopped!")
	}()

	for {
		if sf.isActive && seqNoCount(sf.ackNoSend, sf.seqNoSend) < sf.config.SendUnAckLimitK {
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
			// check all timeouts
			if now.Sub(testFrAliveSendSince) >= sf.config.SendUnAckTimeout1 {
				// now.Sub(startDtActiveSendSince) >= t.SendUnAckTimeout1 ||
				// now.Sub(stopDtActiveSendSince) >= t.SendUnAckTimeout1 ||
				slog.Error("test frame alive confirm timeout t₁")

				return
			}

			// check oldest unacknowledged outbound
			oldestTime, err := sf.peek()
			if err == nil {
				if sf.ackNoSend != sf.seqNoSend &&
					now.Sub(oldestTime) >= sf.config.SendUnAckTimeout1 {
					sf.ackNoSend++

					slog.Error("fatal transmission timeout t₁")
					if sf.stopDtResponseWaiting {
						// sendUFrame(UStopDtConfirm)
						sf.isActive = false
						sf.stopDtResponseWaiting = false
						slog.Debug("data transfer stopped by remote")
						time.Sleep(10 * time.Millisecond)
					}

					return
				}
			}

			// Determine if the earliest i-Frame sent timed out, and reply to the sFrame if it did.
			if sf.ackNoRcv != sf.seqNoRcv &&
				(now.Sub(unAckRcvSince) >= sf.config.RecvUnAckTimeout2 ||
					now.Sub(idleTimeout3Sine) >= timeoutResolution) {
				sendSFrame(sf.seqNoRcv)
				sf.ackNoRcv = sf.seqNoRcv
			}

			// Send TestFrActive frame when idle time is up.
			if now.Sub(idleTimeout3Sine) >= sf.config.IdleTimeout3 {
				sendUFrame(UTestFrActive)
				testFrAliveSendSince = time.Now()
				idleTimeout3Sine = testFrAliveSendSince
				// idleTimeout3Sine = time.Now()
			}

		case apdu := <-sf.rcvRaw:

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
				if !sf.isActive {
					slog.Warn("station not active")

					return // not active, close connection
				}

				if !sf.updateAckNoOut(head.RcvSN) {
					slog.Error("fatal incoming acknowledge either earlier than previous or later than sendTime")

					return
				}

			case IAPCI:
				slog.Debug("RX iFrame", "rx iFrame", head)
				if !sf.isActive {
					slog.Warn("station not active")

					return // not active, close connection
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
				if seqNoCount(sf.ackNoRcv, sf.seqNoRcv) >= sf.config.RecvUnAckLimitW {
					sendSFrame(sf.seqNoRcv)
					sf.ackNoRcv = sf.seqNoRcv
				}
			case UAPCI:
				slog.Debug("RX uFrame", "rx uFrame", head)
				switch head.Function {
				case UStartDtActive:
					if !sf.isActive {
						// 反向连接从站没有共享单活调度器，nil 通道不应阻塞 STARTDT。
						if sf.stopSessions != nil {
							select {
							case sf.stopSessions <- struct{}{}:
							case <-sf.ctx.Done():
								return
							}
							select {
							case <-time.After(500 * time.Millisecond):
							case <-sf.ctx.Done():
								return
							}
						}
						sf.isActive = true
					}

					sendUFrame(UStartDtConfirm)
				//  case uStartDtConfirm:
				// 	isActive = true
				// 	startDtActiveSendSince = willNotTimeout
				case UStopDtActive:
					if len(sf.pending) > 0 {
						sf.stopDtResponseWaiting = true
					} else {
						sendUFrame(UStopDtConfirm)
						sf.isActive = false
						slog.Debug("data transfer stopped by remote")
					}
				// case uStopDtConfirm:
				// 	isActive = false
				// 	stopDtActiveSendSince = willNotTimeout
				case UTestFrActive:
					sendUFrame(UTestFrConfirm)
				case UTestFrConfirm:
					testFrAliveSendSince = willNotTimeout
				default:
					slog.Error("illegal U-Frame functions ignored", "illegal frame", head.Function)
				}
			}
		}
	}
}

// handlerLoop handler iFrame asdu
func (sf *SrvSession) handlerLoop() {
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
			asduPack := asdu.NewEmptyASDU(sf.params)
			if err := asduPack.UnmarshalBinary(rawAsdu); err != nil {
				slog.Error("asdu UnmarshalBinary failed", "error", err)

				continue
			}

			if err := sf.serverHandler(asduPack); err != nil {
				slog.Error("serverHandler failed", "error", err)
			}
		}
	}
}

func (sf *SrvSession) setConnectStatus(status uint32) {
	sf.rwMux.Lock()
	atomic.StoreUint32(&sf.status, status)
	sf.rwMux.Unlock()
}

func (sf *SrvSession) connectStatus() uint32 {
	sf.rwMux.RLock()
	status := atomic.LoadUint32(&sf.status)
	sf.rwMux.RUnlock()

	return status
}

func (sf *SrvSession) cleanUp() {
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

// Wrap-around mechanism
func seqNoCount(nextAckNo, nextSeqNo uint16) uint16 {
	if nextAckNo > nextSeqNo {
		nextSeqNo += 32768
	}

	return nextSeqNo - nextAckNo
}

func (sf *SrvSession) updateAckNoOut(ackNo uint16) (ok bool) {
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

	if len(sf.pending) == 0 && sf.stopDtResponseWaiting {
		sf.sendRaw <- NewUFrame(UStopDtConfirm)
		sf.isActive = false
		sf.stopDtResponseWaiting = false
		slog.Debug("data transfer stopped by remote")
		time.Sleep(10 * time.Millisecond)

		return true
	}

	sf.ackNoSend = ackNo

	return true
}

//nolint:maintidx
func (sf *SrvSession) serverHandler(asduPack *asdu.ASDU) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("server handler", "err", err)
		}
	}()

	slog.Debug("ASDU", "asdu", asduPack)

	// circutor 的 Get* 会消费 InfoObj，先保存原始载荷再解码。
	// 确认必须回显选择位、值和时间标签，不能重新构建时丢掉这些字段。
	original := asduPack.Clone()
	reply := func(cot asdu.CauseOfTransmission, ca asdu.CommonAddr) *asdu.ASDU {
		out := original.Clone()
		out.Coa = cot
		out.CommonAddr = ca
		return out
	}
	switch asduPack.Identifier.Type {
	case asdu.C_SC_NA_1: // Single Command without timeStamp
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SC_NA_1, err.Error())
		}

		cmd := asduPack.GetSingleCmd()

		resp := sf.handler.SingleCommandHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SC_NA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_SC_NA_1)
		}

		return nil
	case asdu.C_SC_TA_1: // Single Command with timeStamp
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SC_TA_1, err.Error())
		}

		cmd := asduPack.GetSingleCmd()

		resp := sf.handler.SingleCommandHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SC_TA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_SC_TA_1)
		}

		return nil
	case asdu.C_DC_NA_1: // Double Command without timeStamp
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_DC_NA_1, err.Error())
		}

		cmd := asduPack.GetDoubleCmd()

		resp := sf.handler.DoubleCommandHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_DC_NA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_DC_NA_1)
		}

		return nil
	case asdu.C_DC_TA_1: // Double Command with timeStamp
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_DC_TA_1, err.Error())
		}

		cmd := asduPack.GetDoubleCmd()

		resp := sf.handler.DoubleCommandHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_DC_TA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_DC_TA_1)
		}

		return nil
	case asdu.C_RC_NA_1: // Step Position Command
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %v", asdu.C_RC_NA_1, err.Error())
		}

		cmd := asduPack.GetStepCmd()

		resp := sf.handler.StepPositionCommandHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			slog.Warn("error sending C_RC_NA_1 reply", "error", err)

			return fmt.Errorf("error with %s type, %v", asdu.C_RC_NA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type", asdu.C_RC_NA_1)
		}

		return nil
	case asdu.C_SE_NC_1: // Set Point Command Short
		err := replyError(asduPack, sf)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SE_NC_1, err.Error())
		}

		cmd := asduPack.GetSetpointFloatCmd()

		resp := sf.handler.SetPointCommandFloatHandler(sf, asduPack, cmd, cmd.Ioa)
		actConRep := reply(resp, asduPack.CommonAddr)
		err = sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_SE_NC_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_SE_NB_1)
		}

		return nil
	case asdu.C_IC_NA_1: // Interrogation Command
		if !(asduPack.Identifier.Coa.Cause == asdu.Activation ||
			asduPack.Identifier.Coa.Cause == asdu.Deactivation) {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCOT)

			return fmt.Errorf("error with %s cause of transmission", string(asduPack.Identifier.Coa.Cause))
		}

		if asduPack.CommonAddr == asdu.InvalidCommonAddr {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCA)

			return fmt.Errorf("error with %s common address", fmt.Sprint(asduPack.CommonAddr))
		}

		ioa, qoi := asduPack.GetInterrogationCmd()
		if ioa != asdu.InfoObjAddrIrrelevant {
			_ = asduPack.SendReplyError(sf, asdu.UnknownIOA)

			return fmt.Errorf("error with %s type", asdu.C_IC_NA_1)
		}

		resp, ca := sf.handler.InterrogationHandler(sf, asduPack, qoi)
		actConRep := reply(resp, asdu.CommonAddr(ca))
		err := sf.Send(actConRep)
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_IC_NA_1, err.Error())
		}

		if resp.IsNegative {
			return fmt.Errorf("error with %s type, negative response", asdu.C_IC_NA_1)
		}

		return nil
	case asdu.C_RD_NA_1: // Read Command
		if asduPack.Identifier.Coa.Cause != asdu.Request {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCOT)

			return fmt.Errorf("error with %s cause of transmission", string(asduPack.Identifier.Coa.Cause))
		}

		if asduPack.CommonAddr == asdu.InvalidCommonAddr {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCA)

			return fmt.Errorf("error with %s common address", fmt.Sprint(asduPack.CommonAddr))
		}

		err := sf.handler.ReadHandler(sf, asduPack, asduPack.GetReadCmd())
		if err != nil {
			return fmt.Errorf("error with %s type, %s", asdu.C_RD_NA_1, err.Error())
		}

		return nil
	case asdu.C_RP_NA_1: // Reset Process Command
		if asduPack.Identifier.Coa.Cause != asdu.Activation {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCOT)

			return fmt.Errorf("error with %s cause of transmission", string(asduPack.Identifier.Coa.Cause))
		}

		if asduPack.CommonAddr == asdu.InvalidCommonAddr {
			_ = asduPack.SendReplyError(sf, asdu.UnknownCA)

			return fmt.Errorf("error with %s type", asdu.C_RP_NA_1)
		}

		ioa, qrp := asduPack.GetResetProcessCmd()
		if ioa != asdu.InfoObjAddrIrrelevant {
			_ = asduPack.SendReplyError(sf, asdu.UnknownIOA)

			return fmt.Errorf("error with %s type", asdu.C_RP_NA_1)
		}

		resp := sf.handler.ResetProcessHandler(sf, asduPack, qrp)
		if resp.IsNegative {
			actConRep := reply(resp, asduPack.CommonAddr)
			err := sf.Send(actConRep)
			if err != nil {
				return fmt.Errorf("error with %s type, %s", asdu.C_RP_NA_1, err.Error())
			}
		}

		return nil
		/*
			// Currently unused commands

			case asdu.C_CS_NA_1: // Clock Synchronization Command
				if asduPack.Identifier.Coa.Cause != asdu.Activation {
					err := asduPack.SendReplyError(sf, asdu.UnknownCOT)

					return fmt.Errorf("error with %s cause of transmission, %w", string(asduPack.Identifier.Coa.Cause), err)
				}

				if asduPack.CommonAddr == asdu.InvalidCommonAddr {
					err := asduPack.SendReplyError(sf, asdu.UnknownCA)

					return fmt.Errorf("error with %s common address, %w", fmt.Sprint(asduPack.CommonAddr), err)
				}

				ioa, tm := asduPack.GetClockSynchronizationCmd()
				if ioa != asdu.InfoObjAddrIrrelevant {
					err := asduPack.SendReplyError(sf, asdu.UnknownIOA)

					return fmt.Errorf("error with %s type, %s", asdu.C_CS_NA_1, err)
				}

				resp := sf.handler.ClockSyncHandler(sf, asduPack, tm)
				actConRep := reply(resp, asduPack.CommonAddr)
				err := sf.Send(actConRep)
				if err != nil {
					return fmt.Errorf("error with %s type, %s", asdu.C_CI_NA_1, err)
				}

				if !resp.IsNegative {
					actConRep := reply(asdu.CauseOfTransmission{Cause:asdu.ActivationTerm}, asduPack.CommonAddr)

					err = sf.Send(actConRep)
					if err != nil {
						return fmt.Errorf("error with %s type, %s", asdu.C_CS_NA_1, err)
					}
				} else {
					return fmt.Errorf("error with %s type, %s", asdu.C_CS_NA_1, err)
				}

				return nil
			case asdu.C_TS_NA_1: // Test Command
				if asduPack.Identifier.Coa.Cause != asdu.Activation {
					err := asduPack.SendReplyError(sf, asdu.UnknownCOT)

					return fmt.Errorf("error with %s cause of transmission, %w", string(asduPack.Identifier.Coa.Cause), err)
				}

				if asduPack.CommonAddr == asdu.InvalidCommonAddr {
					err := asduPack.SendReplyError(sf, asdu.UnknownCA)

					return fmt.Errorf("error with %s common address, %w", fmt.Sprint(asduPack.CommonAddr), err)
				}

				ioa, _ := asduPack.GetTestCommand()
				if ioa != asdu.InfoObjAddrIrrelevant {
					err := asduPack.SendReplyError(sf, asdu.UnknownIOA)

					return fmt.Errorf("error with %s type, %s", asdu.C_TS_NA_1, err)
				}

				err := asduPack.SendReplyMirror(sf, asdu.ActivationCon)
				if err != nil {
					return fmt.Errorf("error with %s type, %s", asdu.C_TS_NA_1, err)
				}

				return nil
		*/
	}

	if err := sf.handler.ASDUHandler(sf, asduPack); err != nil {
		err := asduPack.SendReplyError(sf, asdu.UnknownTypeID)

		return fmt.Errorf("error with %s type, %s", asduPack.Identifier.Type, err.Error())
	}

	return nil
}

func replyError(asduPack *asdu.ASDU, sf *SrvSession) error {
	if asduPack.Identifier.Coa.Cause != asdu.Activation {
		err := asduPack.SendReplyError(sf, asdu.UnknownCOT)

		return fmt.Errorf("error with %s cause of transmission, %w", string(asduPack.Identifier.Coa.Cause), err)
	}
	if asduPack.CommonAddr == asdu.InvalidCommonAddr {
		err := asduPack.SendReplyError(sf, asdu.UnknownCA)

		return fmt.Errorf("error with %s common address, %w", fmt.Sprint(asduPack.CommonAddr), err)
	}

	return nil
}

// IsConnected get server session connected state
func (sf *SrvSession) IsConnected() bool {
	return sf.connectStatus() == connected
}

// Params get params
func (sf *SrvSession) Params() *asdu.Params {
	return sf.params
}

// Send asdu frame
func (sf *SrvSession) Send(u *asdu.ASDU) error {
	// 外部队列模式也必须先校验报文，不能入队后才发现帧永远无法编码。
	data, err := u.MarshalBinary()
	if err != nil {
		return fmt.Errorf("encode ASDU: %w", err)
	}
	if sf.useQueue {
		if sf.queue == nil {
			return errors.New("missing ASDU queue")
		}
		return sf.queue.Enqueue(*u.Clone())
	}
	if !sf.IsConnected() {
		return ErrUseClosedConnection
	}
	select {
	case sf.sendASDU <- data:
		return nil
	default:
		return ErrBufferFull
	}
}

func (sf *SrvSession) SendQueuedASDU(u *asdu.ASDU) error {
	if !sf.IsConnected() {
		return ErrUseClosedConnection
	}

	data, err := u.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error %w", err)
	}

	select {
	case sf.sendASDU <- data:
	default:
		return ErrBufferFull
	}

	return nil
}

func (sf *SrvSession) IsActive() bool {
	return sf.isActive
}

func (sf *SrvSession) AreAllMessagesConfirmed() bool {
	slog.Debug("checking if all messages are confirmed", "pending count", len(sf.pending))

	return len(sf.pending) == 0
}

func (sf *SrvSession) RemoteClose() error {
	return sf.conn.Close()
}

// UnderlyingConn got under net.conn
func (sf *SrvSession) UnderlyingConn() net.Conn {
	return sf.conn
}

//nolint:nestif
func (sf *SrvSession) processQueue() {
	var sendData *asdu.ASDU

	for {
		select {
		case <-sf.ctx.Done():
			return
		default:
			if sf.isActive && sf.queue != nil {
				data, err := sf.queue.Dequeue()
				if err != nil {
					if err.Error() == ErrQueueEmpty && sendData != nil {
						err = sf.SendQueuedASDU(sendData)
						if err != nil {
							slog.Warn("queue data send failed", "error", err)
							if errors.Is(err, ErrUseClosedConnection) || errors.Is(err, ErrBufferFull) {
								_ = sf.queue.ReEnqueue(*sendData)
								sendData = nil
								time.Sleep(timeoutResolution)

								continue
							}
						}

						time.Sleep(timeoutResolution)
						sendData = nil
					}

					time.Sleep(timeoutResolution)

					continue
				}

				if sendData == nil {
					sendData = data.Clone()

					continue
				} else {
					if sendData.Identifier.Type == data.Identifier.Type && sendData.Identifier.Coa.Cause == data.Identifier.Coa.Cause &&
						data.Identifier.Variable.Number == 1 && isCombinableCOT(sendData.Identifier.Coa) {
						combinedASDU, err := combineASDUs(*sendData, data)
						if err == nil {
							sendData = combinedASDU

							continue
						}
					}

					err = sf.SendQueuedASDU(sendData)
					if err != nil {
						slog.Warn("queue data send failed", "error", err)
						if errors.Is(err, ErrUseClosedConnection) || errors.Is(err, ErrBufferFull) {
							_ = sf.queue.ReEnqueue(data)
							_ = sf.queue.ReEnqueue(*sendData)
							sendData = nil
							time.Sleep(timeoutResolution)

							continue
						}
					}

					time.Sleep(1 * time.Millisecond)
					sendData = data.Clone()
				}
			}
		}
	}
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}

	return 0
}

func combineASDUs(asduCombined asdu.ASDU, newASDU asdu.ASDU) (*asdu.ASDU, error) {
	isSeq := false

	if !asduCombined.Variable.IsSequence {
		newIoa := newASDU.ReadInfoObjAddr()
		oldIoa := asduCombined.ReadInfoObjAddr()

		if asduCombined.Variable.Number == 1 && newIoa == oldIoa+1 &&
			newASDU.Identifier.Coa.Cause == asdu.InterrogatedByStation {
			isSeq = true
		}
	} else {
		newIoa := newASDU.ReadInfoObjAddr()
		oldIoa := asduCombined.ReadInfoObjAddr()

		if newIoa != oldIoa+asdu.InfoObjAddr(asduCombined.Variable.Number) {
			return nil, errors.New("cannot combine non contiguous IOA in isSequence package")
		}

		isSeq = true
	}

	num := asduCombined.Variable.Number
	a := asdu.NewASDU(asduCombined.Params, asdu.Identifier{
		Type:       asduCombined.Identifier.Type,
		Variable:   asdu.VariableStruct{IsSequence: isSeq},
		Coa:        asduCombined.Identifier.Coa,
		OrigAddr:   0,
		CommonAddr: asduCombined.Identifier.CommonAddr,
	})

	if err := a.SetVariableNumber(int(num + 1)); err != nil {
		return nil, fmt.Errorf("error trying to set variable number: %v", err.Error())
	}

	a.InfoObj = append(a.InfoObj, asduCombined.InfoObj...)
	if isSeq {
		a.InfoObj = append(a.InfoObj, newASDU.InfoObj[newASDU.Params.InfoObjAddrSize:]...)
	} else {
		a.InfoObj = append(a.InfoObj, newASDU.InfoObj...)
	}

	if len(a.InfoObj) > asdu.ASDUSizeMax-a.IdentifierSize() {
		return nil, fmt.Errorf("ASDU size exceeded: size=%d, max=%d", len(a.InfoObj), asdu.ASDUSizeMax-a.IdentifierSize())
	}

	return a, nil
}

func isCombinableCOT(cot asdu.CauseOfTransmission) bool {
	if cot.Cause == asdu.InterrogatedByStation ||
		cot.Cause == asdu.Spontaneous {
		return true
	}

	return false
}

func (sf *SrvSession) peek() (time.Time, error) {
	if len(sf.pending) > 0 {
		return sf.pending[0].sendTime, nil
	}

	return time.Time{}, errors.New("no pending i-frame")
}

func (sf *SrvSession) emptyChannel(ch chan []byte) {
	var asduList []asdu.ASDU
	for {
		select {
		case m := <-ch:
			asduPack := asdu.NewEmptyASDU(sf.params)
			if err := asduPack.UnmarshalBinary(m); err != nil {
				slog.Error("trying to requeue buffered asdu failed", "error", err)

				continue
			}

			asduList = append(asduList, *asduPack)
		default:
			// Reverse and re-enqueue
			for i := len(asduList) - 1; i >= 0; i-- {
				err := sf.queue.ReEnqueue(asduList[i])
				if err != nil {
					slog.Error("requeue buffered asdu failed", "error", err)
				}
			}

			return
		}
	}
}
