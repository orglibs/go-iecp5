// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package asdu

// Application service data unit in the control direction parameter

// ParameterNormalInfo: Measured value parameter, normalised value Information body.
type ParameterNormalInfo struct {
	Ioa   InfoObjAddr
	Value Normalize
	Qpm   QualifierOfParameterMV
}

// ParameterNormal Measured value parameter, normalised value, only single information object (SQ = 0)
// [P_ME_NA_1], See companion standard 101, subclass 7.3.5.1
// Cause of transmission (coa) for:
// Control direction
// <6> := Activate
// Surveillance direction:
// <7> := Activation Confirmation
// <20> := Responding to station calls
// <21> := Responding to Group 1 calls
// <22> := Responding to Group 2 calls
// <36> := Responding to Group 16 calls
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown application service data unit public address
// <47> := Unknown information object address
func ParameterNormal(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterNormalInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendNormalize(p.Value)
	u.AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterScaledInfo Measured value parameter, scaled value Infobody
type ParameterScaledInfo struct {
	Ioa   InfoObjAddr
	Value int16
	Qpm   QualifierOfParameterMV
}

// ParameterScaled Measured value parameter, Scaled value, Single information object only(SQ = 0)
// [P_ME_NB_1], See companion standard 101, subclass 7.3.5.2
// Cause of transmission (coa) for:
// Control direction:
// <6> := activate
// Surveillance direction:
// <7> := Activation Confirmation
// <20> := Answering station call
// <21> := Answering group 1 call-outs
// <22> := Answering group 2 call-outs
// <36> := Answering group 16 call-outs
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown application service data unit public address
// <47> := Unknown information object address
func ParameterScaled(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterScaledInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NB_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendScaled(p.Value).AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterFloatInfo Measured parameters, short floating-point numbers InfoBody
type ParameterFloatInfo struct {
	Ioa   InfoObjAddr
	Value float32
	Qpm   QualifierOfParameterMV
}

// ParameterFloat Measured value parameter, short floating-point number, only single information object(SQ = 0)
// [P_ME_NC_1], See companion standard 101, subclass 7.3.5.3
// Cause of transmission (coa) for
// Control direction:
// <6> := activate
//
//	Surveillance direction:
//
// <7> := Activation Confirmation
// <20> := Answering to the call of the station
// <21> := Answering group 1 call-outs
// <22> := Answering group 2 call-outs
// <36> := Answering group 16 call-outs
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown application service data unit public address
// <47> := Unknown information object address
func ParameterFloat(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterFloatInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NC_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendFloat32(p.Value).AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterActivationInfo: Parameter activation format
type ParameterActivationInfo struct {
	Ioa InfoObjAddr
	Qpa QualifierOfParameterAct
}

// ParameterActivation parameter activation, only for a single information objec(SQ = 0)
// [P_AC_NA_1], See companion standard 101, subclass 7.3.5.4
// Cause of transmission (coa) for
// Control direction:
// <6> := activate
// <8> := deactivate
// Surveillance direction:
// <7> := Activation Confirmation
// <9> := Stop Activation Confirmation
// <44> := Unknown type identification
// <45> := Unknown reason for transmission
// <46> := Unknown application service
// <47> := Unknown information object address
func ParameterActivation(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterActivationInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_AC_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendBytes(byte(p.Qpa))
	return c.Send(u)
}

// GetParameterNormal [P_ME_NA_1]，Get the measured value parameter,standard value InfoBody
func (sf *ASDU) GetParameterNormal() ParameterNormalInfo {
	return ParameterNormalInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeNormalize(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterScaled [P_ME_NB_1]，Get the measured value parameter, scaled value InfoBody
func (sf *ASDU) GetParameterScaled() ParameterScaledInfo {
	return ParameterScaledInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeScaled(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterFloat [P_ME_NC_1]，Get measured value parameter, short floating point number Information body
func (sf *ASDU) GetParameterFloat() ParameterFloatInfo {
	return ParameterFloatInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeFloat32(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterActivation [P_AC_NA_1]，Get Parameter Activation value
func (sf *ASDU) GetParameterActivation() ParameterActivationInfo {
	return ParameterActivationInfo{
		sf.DecodeInfoObjAddr(),
		QualifierOfParameterAct(sf.infoObj[0]),
	}
}
