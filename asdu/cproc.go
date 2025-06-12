// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package asdu

import (
	"fmt"
	"time"
)

// Application service data unit for process information in the control direction

// SingleCommandInfo, Single command Message body
type SingleCommandInfo struct {
	Ioa   InfoObjAddr
	Value bool
	Qoc   QualifierOfCommand
	Time  time.Time
}

// SingleCmd sends a type identification [C_SC_NA_1] or [C_SC_TA_1]. Single command, only a single information object(SQ = 0)
// [C_SC_NA_1] See companion standard 101, subclass 7.3.2.1
// [C_SC_TA_1] See companion standard 101,
// Causes of transmission (coa) for
// Control direction：
// <6> := activate
// <8> := deactivate
// surveillance direction：
// <7> := Confirmation of activation
// <9> := Deactivation Confirmation
// <10> := Termination of activation
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown application service data unit public address
// <47> := Unknown information object address
func SingleCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SingleCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	value := cmd.Qoc.Value()
	if cmd.Value {
		value |= 0x01
	}

	u.AppendBytes(value)
	switch typeID {
	case C_SC_NA_1:
	case C_SC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// DoubleCommandInfo
type DoubleCommandInfo struct {
	Ioa   InfoObjAddr
	Value DoubleCommand
	Qoc   QualifierOfCommand
	Time  time.Time
}

// DoubleCmd sends a type identification [C_DC_NA_1] or [C_DC_TA_1]. Dual command, only a single information object(SQ = 0)
// [C_DC_NA_1] See companion standard 101, subclass 7.3.2.2
// [C_DC_TA_1] See companion standard 101,
// Cause of transmission (coa) for
// Control direction：
// <6> := activate
// <8> := deactivate
// surveillance direction：
// <7> := Activation Confirmation
// <9> := Deactivation Confirmation
// <10> := Activation terminated
// <44> := Unknown Type Identifier
// <45> := Unknown cause of transmission
// <46> := Unknown application service data unit public address
// <47> := Unknown information object address
func DoubleCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd DoubleCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBytes(cmd.Qoc.Value() | byte(cmd.Value&0x03))

	switch typeID {
	case C_DC_NA_1:
	case C_DC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// StepCommandInfo
type StepCommandInfo struct {
	Ioa   InfoObjAddr
	Value StepCommand
	Qoc   QualifierOfCommand
	Time  time.Time
}

// StepCmd sends a type [C_RC_NA_1] or [C_RC_TA_1]. StepCmd sends a type [C_RC_NA_1] or [C_RC_TA_1].
// StepCmd commands, only a single information object (SQ = 0).
// [C_RC_NA_1] See companion standard 101, subclass 7.3.2.3.
// [C_RC_TA_1] See companion standard 101, subclass 7.3.2.3.
// The reason for transmission (coa) is used for the
// control direction:
// <6> := activate
// <8> := deactivate
// Monitoring direction:
// <7> := Activation confirmation
// <9> := Deactivation confirmation
// <10> := Activation terminated
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown public address of application service data unit
// <47> := Unknown address of information object

func StepCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd StepCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBytes(cmd.Qoc.Value() | byte(cmd.Value&0x03))

	switch typeID {
	case C_RC_NA_1:
	case C_RC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// SetpointCommandNormalInfo Setup command, normalised values InfoBody
type SetpointCommandNormalInfo struct {
	Ioa   InfoObjAddr
	Value Normalize
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdNormal sends a type [C_SE_NA_1] or [C_SE_TA_1]. Setting commands, normalised values, single information objects only(SQ = 0)
// [C_SE_NA_1] See companion standard 101, subclass 7.3.2.4
// [C_SE_TA_1] See companion standard 101,
// Cause of transmission (coa) for
// control direction：
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
func SetpointCmdNormal(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandNormalInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendNormalize(cmd.Value).AppendBytes(cmd.Qos.Value())
	switch typeID {
	case C_SE_NA_1:
	case C_SE_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// SetpointCommandScaledInfo
type SetpointCommandScaledInfo struct {
	Ioa   InfoObjAddr
	Value int16
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdScaled sends a type [C_SE_NB_1] or [C_SE_TB_1]. Set command, standardization value, only single information object (SQ = 0)
// [c_se_nb_1] See Companion Standard 101, Subclass 7.3.2.5
// [c_se_tb_1] See Companion Standard 101,
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

func SetpointCmdScaled(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandScaledInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendScaled(cmd.Value).AppendBytes(cmd.Qos.Value())

	switch typeID {
	case C_SE_NB_1:
	case C_SE_TB_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// SetpointCommandFloatInfo Set command, short -floating point digital information body
type SetpointCommandFloatInfo struct {
	Ioa   InfoObjAddr
	Value float32
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdFloat sends a type [C_SE_NC_1] or [C_SE_TC_1].Set command, short floating point number, only a single information object (SQ = 0)
// [C_SE_NC_1] See Companion Standard 101, Subclass 7.3.2.6
// [c_se_tc_1] See Companion Standard 101,
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
func SetpointCmdFloat(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandFloatInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendFloat32(cmd.Value).AppendBytes(cmd.Qos.Value())

	switch typeID {
	case C_SE_NC_1:
	case C_SE_TC_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	if err := c.Send(u); err != nil {
		return fmt.Errorf("SetpointCmdFloat: %w", err)
	}

	return nil
}

// BitsString32CommandInfo Bit strings command information body
type BitsString32CommandInfo struct {
	Ioa   InfoObjAddr
	Value uint32
	Time  time.Time
}

// BitsString32Cmd sends a type [C_BO_NA_1] or [C_BO_TA_1]. Bit string command, only a single information object (SQ = 0)
// [C_BO_NA_1] See Companion Standard 101, Subclass 7.3.2.7
// [C_BO_TA_1] See Companion Standard 101,
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
func BitsString32Cmd(c Connect, typeID TypeID, coa CauseOfTransmission, commonAddr CommonAddr, cmd BitsString32CommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}

	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		commonAddr,
	})
	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBitsString32(cmd.Value)

	switch typeID {
	case C_BO_NA_1:
	case C_BO_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	if err := c.Send(u); err != nil {
		return fmt.Errorf("BitsString32Cmd: %w", err)
	}

	return nil
}

// GetSingleCmd [C_SC_NA_1] or [C_SC_TA_1] Get the order command information body
func (sf *ASDU) GetSingleCmd() SingleCommandInfo {
	var s SingleCommandInfo

	s.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	s.Value = value&0x01 == 0x01
	s.Qoc = ParseQualifierOfCommand(value & 0xfe)

	switch sf.Type {
	case C_SC_NA_1:
	case C_SC_TA_1:
		s.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return s
}

// GetDoubleCmd [C_DC_NA_1] or [C_DC_TA_1] Get double command information body
func (sf *ASDU) GetDoubleCmd() DoubleCommandInfo {
	var cmd DoubleCommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	cmd.Value = DoubleCommand(value & 0x03)
	cmd.Qoc = ParseQualifierOfCommand(value & 0xfc)

	switch sf.Type {
	case C_DC_NA_1:
	case C_DC_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetStepCmd [C_RC_NA_1] or [C_RC_TA_1] Get step adjustment command information body
func (sf *ASDU) GetStepCmd() StepCommandInfo {
	var cmd StepCommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	cmd.Value = StepCommand(value & 0x03)
	cmd.Qoc = ParseQualifierOfCommand(value & 0xfc)

	switch sf.Type {
	case C_RC_NA_1:
	case C_RC_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointNormalCmd [C_SE_NA_1] or [C_SE_TA_1] Get the setting command, the one -time value information body
func (sf *ASDU) GetSetpointNormalCmd() SetpointCommandNormalInfo {
	var cmd SetpointCommandNormalInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeNormalize()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NA_1:
	case C_SE_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointCmdScaled [C_SE_NB_1] or [C_SE_TB_1] Get the setting command, standardized value information body
func (sf *ASDU) GetSetpointCmdScaled() SetpointCommandScaledInfo {
	var cmd SetpointCommandScaledInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeScaled()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NB_1:
	case C_SE_TB_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointFloatCmd [C_SE_NC_1] or [C_SE_TC_1] Get the setting command, short -floating -point number information body
func (sf *ASDU) GetSetpointFloatCmd() SetpointCommandFloatInfo {
	var cmd SetpointCommandFloatInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeFloat32()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NC_1:
	case C_SE_TC_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetBitsString32Cmd [C_BO_NA_1] or [C_BO_TA_1] Get Bitstring Command Message Body
func (sf *ASDU) GetBitsString32Cmd() BitsString32CommandInfo {
	var cmd BitsString32CommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeBitsString32()
	switch sf.Type {
	case C_BO_NA_1:
	case C_BO_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetStepPositionCmd [C_RC_NA_1] or [C_RC_TA_1] Get the setting command, standardized value information body
func (sf *ASDU) GetStepPositionCmd() StepCommandInfo {
	var cmd StepCommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = ParseStepCommand(sf.DecodeByte())
	cmd.Qoc = ParseQualifierOfCommand(sf.DecodeByte())

	switch sf.Type {
	case C_RC_NA_1:
	case C_RC_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}
