// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package asdu

import (
	"time"
)

// Application service data unit in control direction system information

// InterrogationCmd send a new interrogation command [C_IC_NA_1].  single information object (SQ = 0)
// [C_IC_NA_1] See Companion Standard 101, Subclass 7.3.4.1
// The reason for transmission (coa) is used for it
// control direction:
// <6>: = Activate
// <8>: = Stop activation
// Monitoring direction:
// <7>: = Activate confirmation
// <9>: = Stop activation confirmation
// <10>: = Activate termination
// <44>: = Unknown type logo
// <45>: = Unknown reasons for transmission
// <46>: = Unknown application service data unit public address
// <47>: = Unknown information object address
func InterrogationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qoi QualifierOfInterrogation) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation || coa.Cause == ActivationTerm) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		C_IC_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendBytes(byte(qoi))

	return c.Send(u)
}

// Countrinterrogationcmd Send Counter Interrogation Command [C_CI_NA_1], the number of summoning commands, only a single information object (SQ = 0)
// [C_CI_NA_1] See Companion Standard 101, Subclass 7.3.4.2
// The reason for transmission (coa) is used for it
// control direction:
// <6>: = Activate
// Monitoring direction:
// <7>: = Activate confirmation
// <10>: = Activate termination
// <44>: = Unknown type logo
// <45>: = Unknown reasons for transmission
// <46>: = Unknown application service data unit public address
// <47>: = Unknown information object address
func CounterInterrogationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qcc QualifierCountCall) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	coa.Cause = Activation

	u := NewASDU(c.Params(), Identifier{
		C_CI_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendBytes(qcc.Value())

	return c.Send(u)
}

// ReadCmd send read command [C_RD_NA_1], read command, single message object only (SQ = 0)
// [C_RD_NA_1] see companion standard 101, subclass 7.3.4.3
// The reason for transmission (coa) is used for the
// Control direction:
// <5> := request
// Monitoring direction:
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of information object
func ReadCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, ioa InfoObjAddr) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	coa.Cause = Request

	u := NewASDU(c.Params(), Identifier{
		C_RD_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(ioa); err != nil {
		return err
	}

	return c.Send(u)
}

// ClockSynchronizationCmd send clock sync command [C_CS_NA_1], clock sync command, single message object only (SQ = 0)
// [C_CS_NA_1] See companion standard 101, subclass 7.3.4.4.
// The reason for transmission (coa) is used in the
// Control direction:
// <6> := active
// Monitoring direction:
// <7> := activation confirmation
// <10> := Activation terminated
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of the information object
func ClockSynchronizationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, isValidTime bool, t time.Time) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	coa.Cause = Activation

	u := NewASDU(c.Params(), Identifier{
		C_CS_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendBytes(CP56Time2a(t, u.InfoObjTimeZone, isValidTime)...)

	return c.Send(u)
}

// TestCommand send test command [C_TS_NA_1], test command, only single message object (SQ = 0)
// [C_TS_NA_1] see companion standard 101, subclass 7.3.4.5
// The reason for transmission (coa) is used for the
// Control direction:
// <6> := active
// Monitoring direction:
// <7> := activation confirmation
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of the information object
func TestCommand(c Connect, coa CauseOfTransmission, ca CommonAddr) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	coa.Cause = Activation

	u := NewASDU(c.Params(), Identifier{
		C_TS_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendBytes(byte(FBPTestWord&0xff), byte(FBPTestWord>>8))

	return c.Send(u)
}

// ResetProcessCmd send reset process command [C_RP_NA_1], reset process command, only single message object (SQ = 0)
// [C_RP_NA_1] see companion standard 101, subclass 7.3.4.6
// The reason for transmission (coa) is used for the
// Control direction:
// <6> := active
// Monitoring direction:
// <7> := activation confirmation
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of the information object
func ResetProcessCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qrp QualifierOfResetProcessCmd) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		C_RP_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendBytes(byte(qrp))

	return c.Send(u)
}

// DelayAcquireCommand send delay acquire command [C_CD_NA_1], delay acquire command, single message object only (SQ = 0)
// [C_CD_NA_1] See companion standard 101, subclass 7.3.4.7.
// The reason for transmission (coa) is used to
// Control direction:
// <3> := burst
// <6> := active
// Monitoring direction:
// <7> := activation confirmation
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of the information object
func DelayAcquireCommand(c Connect, coa CauseOfTransmission, ca CommonAddr, msec uint16) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Activation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		C_CD_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendCP16Time2a(msec)

	return c.Send(u)
}

// TestCommandCP56Time2a send test command [C_TS_TA_1], test command, only single message object (SQ = 0)
// Reason for transmission (coa) used for
// Control direction:
// <6> := active
// Monitoring direction:
// <7> := activation confirmation
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of the information object
func TestCommandCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, t time.Time) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	// 控制方向测试命令只允许 Activation，不能沿用调用方任意 COT。
	coa.Cause = Activation
	u := NewASDU(c.Params(), Identifier{
		C_TS_TA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}

	u.AppendUint16(FBPTestWord)
	u.AppendCP56Time2a(t, u.InfoObjTimeZone, true)

	return c.Send(u)
}

// GetInterrogationCmd [C_IC_NA_1] Get total summoning information body (information object address, summoning qualifier)
func (sf *ASDU) GetInterrogationCmd() (InfoObjAddr, QualifierOfInterrogation) {
	return sf.DecodeInfoObjAddr(), QualifierOfInterrogation(sf.InfoObj[0])
}

// GetCounterInterrogationCmd [C_CI_NA_1] Get Counter Interrogation Information Body (Information Object Address, Counter Interrogation Qualifier)
func (sf *ASDU) GetCounterInterrogationCmd() (InfoObjAddr, QualifierCountCall) {
	return sf.DecodeInfoObjAddr(), ParseQualifierCountCall(sf.InfoObj[0])
}

// GetReadCmd [C_RD_NA_1] Get read command message address
func (sf *ASDU) GetReadCmd() InfoObjAddr {
	return sf.DecodeInfoObjAddr()
}

// GetClockSynchronizationCmd [C_CS_NA_1] Get Clock Synchronisation Command Message Body (Message Object Address, Time)
func (sf *ASDU) GetClockSynchronizationCmd() (InfoObjAddr, time.Time) {
	return sf.DecodeInfoObjAddr(), sf.DecodeCP56Time2a()
}

// GetTestCommand [C_TS_NA_1], get the test command message body (address of the message object, whether it is a test word)
func (sf *ASDU) GetTestCommand() (InfoObjAddr, bool) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16() == FBPTestWord
}

// GetResetProcessCmd [C_RP_NA_1] Get the reset process command message body (message object address, reset process command qualifier).
func (sf *ASDU) GetResetProcessCmd() (InfoObjAddr, QualifierOfResetProcessCmd) {
	return sf.DecodeInfoObjAddr(), QualifierOfResetProcessCmd(sf.InfoObj[0])
}

// GetDelayAcquireCommand [C_CD_NA_1] Get Delay Acquire Command message body (address of message object, number of milliseconds of delay).
func (sf *ASDU) GetDelayAcquireCommand() (InfoObjAddr, uint16) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16()
}

// GetTestCommandCP56Time2a [C_TS_TA_1], get the test command message body (address of the message object, whether it is a test word)
func (sf *ASDU) GetTestCommandCP56Time2a() (InfoObjAddr, bool, time.Time) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16() == FBPTestWord, sf.DecodeCP56Time2a()
}
