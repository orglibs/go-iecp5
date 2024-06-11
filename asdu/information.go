// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package asdu

// Application Service Data Units - Information Objects

// InfoObjAddr is the information object address.
// See companion standard 101, subclass 7.2.5.
// The width is controlled by Params.InfoObjAddrSize.
// <0>: Address of irrelevant information objects
// - width 1: <1..255>
// - width 2: <1..65535>
// - width 3: <1..16777215>
type InfoObjAddr uint

// InfoObjAddrIrrelevant Zero means that the information object address is irrelevant.
const InfoObjAddrIrrelevant InfoObjAddr = 0

// SinglePoint is a measured value of a switch.
// See companion standard 101, subclass 7.2.6.1.
type SinglePoint byte

// SinglePoint defined
const (
	SPIOff SinglePoint = iota
	SPIOn
)

// Value single point to byte
func (sf SinglePoint) Value() byte {
	return byte(sf & 0x01)
}

// DoublePoint is a measured value of a determination aware switch.
// See companion standard 101, subclass 7.2.6.2.
type DoublePoint byte

// DoublePoint defined
const (
	DPIIndeterminateOrIntermediate DoublePoint = iota
	DPIDeterminedOff
	DPIDeterminedOn
	DPIIndeterminate
)

// Value double point to byte
func (sf DoublePoint) Value() byte {
	return byte(sf & 0x03)
}

// QualityDescriptor Quality descriptor flags attribute measured values.
// See companion standard 101, subclass 7.2.6.3.
type QualityDescriptor byte

// QualityDescriptor defined.
const (
	// QDSOverflow marks whether the value is beyond a predefined range.
	QDSOverflow QualityDescriptor = 1 << iota
	_                             // reserve
	_                             // reserve
	_                             // reserve
	// QDSBlocked flags that the value is blocked for transmission; the
	// value remains in the state that was acquired before it was blocked.
	QDSBlocked
	// QDSSubstituted flags that the value was provided by the input of
	// an operator (dispatcher) instead of an automatic source.
	QDSSubstituted
	// QDSNotTopical flags that the most recent update was unsuccessful.
	QDSNotTopical
	// QDSInvalid flags that the value was incorrectly acquired.
	QDSInvalid

	// QDSGood means no flags, no problems.
	QDSGood QualityDescriptor = 0
)

// QualityDescriptorProtection  Quality descriptor Protection Equipment flags attribute.
// See companion standard 101, subclass 7.2.6.4.
type QualityDescriptorProtection byte

// QualityDescriptorProtection defined.
const (
	_ QualityDescriptorProtection = 1 << iota // reserve
	_                                         // reserve
	_                                         // reserve
	// QDPElapsedTimeInvalid flags that the elapsed time was incorrectly acquired.
	QDPElapsedTimeInvalid
	// QDPBlocked flags that the value is blocked for transmission; the
	// value remains in the state that was acquired before it was blocked.
	QDPBlocked
	// QDPSubstituted flags that the value was provided by the input of
	// an operator (dispatcher) instead of an automatic source.
	QDPSubstituted
	// QDPNotTopical flags that the most recent update was unsuccessful.
	QDPNotTopical
	// QDPInvalid flags that the value was incorrectly acquired.
	QDPInvalid

	// QDPGood means no flags, no problems.
	QDPGood QualityDescriptorProtection = 0
)

// StepPosition is a measured value with transient state indication.
// Measured values with transient status indication for transformer step positions or other step positions
// See companion standard 101, subclass 7.2.6.5.
// Val range <-64..63>
// bit[0-5]: <-64..63>
// NOTE: bit6 is a sign bit
// bit7: 0: Device is not in transient state 1: Device is in transient state
type StepPosition struct {
	Val          int
	HasTransient bool
}

// Value returns step position value.
func (sf StepPosition) Value() byte {
	p := sf.Val & 0x7f
	if sf.HasTransient {
		p |= 0x80
	}
	return byte(p)
}

// ParseStepPosition parse byte to StepPosition.
func ParseStepPosition(b byte) StepPosition {
	step := StepPosition{HasTransient: (b & 0x80) != 0}
	if b&0x40 == 0 {
		step.Val = int(b & 0x3f)
	} else {
		step.Val = int(b) | (-1 &^ 0x3f)
	}
	return step
}

// Normalize is a 16-bit normalized value in[-1, 1 − 2⁻¹⁵]..
// Normalised value f normalised = 32768 * f true / full code value
// See companion standard 101, subclass 7.2.6.6.
type Normalize int16

// Float64 returns the value in [-1, 1 − 2⁻¹⁵].
func (sf Normalize) Float64() float64 {
	return float64(sf) / 32768
}

// BinaryCounterReading is binary counter reading
// See companion standard 101, subclass 7.2.6.9.
// CounterReading: Counter reading [bit0...bit31]. .bit31]
// SeqNumber: Sequential notation [bit32... .bit40]. .bit40]
// SQ: Sequence number [bit32... .bit36]. .bit36]
// CY: Carry forward [bit37].
// CA: Count is adjusted
// IV: Invalid
type BinaryCounterReading struct {
	CounterReading int32
	SeqNumber      byte
	HasCarry       bool
	IsAdjusted     bool
	IsInvalid      bool
}

// SingleEvent is single event
// See companion standard 101, subclass 7.2.6.10.
type SingleEvent byte

// SingleEvent dSequenceNotationefined
const (
	SEIndeterminateOrIntermediate SingleEvent = iota // Uncertain or intermediate status
	SEDeterminedOff                                  // Determine status on
	SEDeterminedOn                                   // Determine status off
	SEIndeterminate                                  // Uncertain or intermediate status
)

// StartEvent Start event protection
type StartEvent byte

// StartEvent defined
// See companion standard 101, subclass 7.2.6.11.
const (
	SEPGeneralStart          StartEvent = 1 << iota // General startup
	SEPStartL1                                      // A-phase protection activation
	SEPStartL2                                      // B-phase protection activation
	SEPStartL3                                      // C-phase protection activation
	SEPStartEarthCurrent                            // Ground current protection activation
	SEPStartReverseDirection                        // Reverse protection activation
	// other reserved
)

// OutputCircuitInfo output command information
// See companion standard 101, subclass 7.2.6.12.
type OutputCircuitInfo byte

// OutputCircuitInfo defined
const (
	OCIGeneralCommand OutputCircuitInfo = 1 << iota // General command output to output circuit
	OCICommandL1                                    // A-phase protection command output to output circuit
	OCICommandL2                                    // B-phase protection command output to output circuit
	OCICommandL3                                    // C-phase protection command output to output circuit
	// other reserved
)

// FBPTestWord test special value
// See companion standard 101, subclass 7.2.6.14.
const FBPTestWord uint16 = 0x55aa

// SingleCommand Single command
// See companion standard 101, subclass 7.2.6.15.
type SingleCommand byte

// SingleCommand defined
const (
	SCOOn SingleCommand = iota
	SCOOff
)

// DoubleCommand double command
// See companion standard 101, subclass 7.2.6.16.
type DoubleCommand byte

// DoubleCommand defined
const (
	DCONotAllow0 DoubleCommand = iota
	DCOOn
	DCOOff
	DCONotAllow3
)

// StepCommand step command
// See companion standard 101, subclass 7.2.6.17.
type StepCommand byte

// StepCommand defined
const (
	SCONotAllow0 StepCommand = iota
	SCOStepDown
	SCOStepUP
	SCONotAllow3
)

// COICause Initialization reason
// See companion standard 101, subclass 7.2.6.21.
type COICause byte

// COICause defined
// 0: Local power on
// 1: Local manual reset
// 2: Remote reset
// <3...31>: Standard definitions reserved for the preparation of this distribution standard
// <32... .127>: Reserved for specific use. .127>: Reserved for specific use
const (
	COILocalPowerOn COICause = iota
	COILocalHandReset
	COIRemoteReset
)

// CauseOfInitial cause of initial
// Cause:  see COICause
// IsLocalChange: false - initialisation of local parameters not changed
//
// true - initialisation after changing local parameters
type CauseOfInitial struct {
	Cause         COICause
	IsLocalChange bool
}

// ParseCauseOfInitial parse byte to cause of initial
func ParseCauseOfInitial(b byte) CauseOfInitial {
	return CauseOfInitial{
		Cause:         COICause(b & 0x7f),
		IsLocalChange: b&0x80 == 0x80,
	}
}

// Value CauseOfInitial to byte
func (sf CauseOfInitial) Value() byte {
	if sf.IsLocalChange {
		return byte(sf.Cause | 0x80)
	}
	return byte(sf.Cause)
}

// QualifierOfInterrogation Qualifier Of Interrogation
// See companion standard 101, subclass 7.2.6.22.
type QualifierOfInterrogation byte

// QualifierOfInterrogation defined
const (
	// <1...19>: Reservations for standard definitions
	QOIStation QualifierOfInterrogation = 20 + iota // interrogated by station interrogation
	QOIGroup1                                       // interrogated by group 1 interrogation
	QOIGroup2                                       // interrogated by group 2 interrogation
	QOIGroup3                                       // interrogated by group 3 interrogation
	QOIGroup4                                       // interrogated by group 4 interrogation
	QOIGroup5                                       // interrogated by group 5 interrogation
	QOIGroup6                                       // interrogated by group 6 interrogation
	QOIGroup7                                       // interrogated by group 7 interrogation
	QOIGroup8                                       // interrogated by group 8 interrogation
	QOIGroup9                                       // interrogated by group 9 interrogation
	QOIGroup10                                      // interrogated by group 10 interrogation
	QOIGroup11                                      // interrogated by group 11 interrogation
	QOIGroup12                                      // interrogated by group 12 interrogation
	QOIGroup13                                      // interrogated by group 13 interrogation
	QOIGroup14                                      // interrogated by group 14 interrogation
	QOIGroup15                                      // interrogated by group 15 interrogation
	QOIGroup16                                      // interrogated by group 16 interrogation

	// <37..63>: reserved for standard definitions
	// <64..255>: reserved for specific use

	// 0: not used
	QOIUnused QualifierOfInterrogation = 0
)

// QCCRequest [bit0...bit5]. .bit5]
// See companion standard 101, subclass 7.2.6.23.
type QCCRequest byte

// QCCFreeze freeze [bit6,bit7]
// See companion standard 101, subclass 7.2.6.23.
type QCCFreeze byte

// QCCRequest and QCCFreeze defined
const (
	QCCUnused QCCRequest = iota
	QCCGroup1
	QCCGroup2
	QCCGroup3
	QCCGroup4
	QCCTotal
	// <6..31>: Defined for standard use.
	// <32..63>: Reserved for specific use.
	QCCFrzRead          QCCFreeze = 0x00 // Read (no freeze or reset)
	QCCFrzFreezeNoReset QCCFreeze = 0x40 // Count freeze without reset (frozen values are cumulative)
	QCCFrzFreezeReset   QCCFreeze = 0x80 // Count freeze with reset (frozen values are incremental information)
	QCCFrzReset         QCCFreeze = 0xc0 // Count reset
)

// QualifierCountCall
// See companion standard 101, subclass 7.2.6.23.
type QualifierCountCall struct {
	Request QCCRequest
	Freeze  QCCFreeze
}

// ParseQualifierCountCall parse byte to QualifierCountCall
func ParseQualifierCountCall(b byte) QualifierCountCall {
	return QualifierCountCall{
		Request: QCCRequest(b & 0x3f),
		Freeze:  QCCFreeze(b & 0xc0),
	}
}

// Value QualifierCountCall to byte
func (sf QualifierCountCall) Value() byte {
	return byte(sf.Request&0x3f) | byte(sf.Freeze&0xc0)
}

// QPMCategory
type QPMCategory byte

// QPMCategory defined
const (
	QPMUnused    QPMCategory = iota // 0: not used
	QPMThreshold                    // 1: threshold value
	QPMSmoothing                    // 2: smoothing factor (filter time constant)
	QPMLowLimit                     // 3: low limit for transmission of measured values
	QPMHighLimit                    // 4: high limit for transmission of measured values

	// 5‥31: reserved for standard definitions of sf companion standard (compatible range)
	// 32‥63: reserved for special use (private range)

	QPMChangeFlag      QPMCategory = 0x40 // bit6 marks local parameter change  Local parameter changes
	QPMInOperationFlag QPMCategory = 0x80 // bit7 marks parameter operation parameter in the run
)

// QualifierOfParameterMV Qualifier Of Parameter Of Measured Values
// See companion standard 101, subclass 7.2.6.24.
// QPMCategory : [bit0...bit5] Parameter type
// IsChange : [bit6] local parameter change, false - no change, true - change
// IsInOperation : [bit7] parameter is running, false - running, true - not running
type QualifierOfParameterMV struct {
	Category      QPMCategory
	IsChange      bool
	IsInOperation bool
}

// ParseQualifierOfParamMV parse byte to QualifierOfParameterMV
func ParseQualifierOfParamMV(b byte) QualifierOfParameterMV {
	return QualifierOfParameterMV{
		Category:      QPMCategory(b & 0x3f),
		IsChange:      b&0x40 == 0x40,
		IsInOperation: b&0x80 == 0x80,
	}
}

// Value QualifierOfParameterMV to byte
func (sf QualifierOfParameterMV) Value() byte {
	v := byte(sf.Category) & 0x3f
	if sf.IsChange {
		v |= 0x40
	}
	if sf.IsInOperation {
		v |= 0x80
	}
	return v
}

// QualifierOfParameterAct Qualifier Of Parameter Activation
// See companion standard 101, subclass 7.2.6.25.
type QualifierOfParameterAct byte

// QualifierOfParameterAct defined
const (
	QPAUnused QualifierOfParameterAct = iota
	// Activate/deactivate the parameters loaded before this (address of information object = 0)
	QPADeActPrevLoadedParameter
	// Activate/deactivate the parameters of the addressed information object.
	QPADeActObjectParameter
	// Activate/deactivate the addressed information object for continuous cyclic or periodic transmission
	QPADeActObjectTransmission
	// 4‥127: reserved for standard definitions of sf companion standard (compatible range)
	// 128‥255: reserved for special use (private range)
)

// QOCQual the qualifier of qual.
// See companion standard 101, subclass 7.2.6.26.
type QOCQual byte

// QOCQual defined
const (
	// 0: no additional definition
	// No additional definitions
	QOCNoAdditionalDefinition QOCQual = iota
	// 1: short pulse duration (circuit-breaker), duration determined by a system parameter in the outstation
	// short pulse duration (circuit breaker), duration determined by system parameters in the controlled station
	QOCShortPulseDuration
	// 2: long pulse duration, duration determined by a system parameter in the outstation
	// Long pulse duration, which is determined by the system parameters in the controlled station.
	QOCLongPulseDuration
	// 3: persistent output
	// Continuous output
	QOCPersistentOutput
	//	4‥8: reserved for standard definitions of sf companion standard
	//	9‥15: reserved for the selection of other predefined functions
	//	16‥31: reserved for special use (private range)
)

// QualifierOfCommand is a  qualifier of command. imperative qualifier
// See companion standard 101, subclass 7.2.6.26.
// See section 5, subclass 6.8.
// InSelect: true - selects, false - executes.
type QualifierOfCommand struct {
	Qual     QOCQual
	InSelect bool
}

// ParseQualifierOfCommand parse byte to QualifierOfCommand
func ParseQualifierOfCommand(b byte) QualifierOfCommand {
	return QualifierOfCommand{
		Qual:     QOCQual((b >> 2) & 0x1f),
		InSelect: b&0x80 == 0x80,
	}
}

// Value QualifierOfCommand to byte
func (sf QualifierOfCommand) Value() byte {
	v := (byte(sf.Qual) & 0x1f) << 2
	if sf.InSelect {
		v |= 0x80
	}
	return v
}

// QualifierOfResetProcessCmd reset process command qualifier
// See companion standard 101, subclass 7.2.6.27.
type QualifierOfResetProcessCmd byte

// QualifierOfResetProcessCmd defined
const (
	// Not used
	QRPUnused QualifierOfResetProcessCmd = iota
	// Total reset of the process
	QPRGeneralRest
	// Reset event buffer for messages with timestamps to be processed
	QPRResetPendingInfoWithTimeTag
	// <3..127>: reserved for standard use
	// <128..255>: Reserved for specific use.
)

/*
TODO: file Documentation related undefined
*/

// QOSQual is the qualifier of a set-point command qual.
// See companion standard 101, subclass 7.2.6.39.
//
//	0: default
//	0‥63: reserved for standard definitions of sf companion standard (compatible range)
//	64‥127: reserved for special use (private range)
type QOSQual uint

// QualifierOfSetpointCmd is a qualifier of command. Setting command qualifiers
// See section 5, subclass 6.8.
// InSelect: true - selects, false - executes.
type QualifierOfSetpointCmd struct {
	Qual     QOSQual
	InSelect bool
}

// ParseQualifierOfSetpointCmd parse byte to QualifierOfSetpointCmd
func ParseQualifierOfSetpointCmd(b byte) QualifierOfSetpointCmd {
	return QualifierOfSetpointCmd{
		Qual:     QOSQual(b & 0x7f),
		InSelect: b&0x80 == 0x80,
	}
}

// Value QualifierOfSetpointCmd to byte
func (sf QualifierOfSetpointCmd) Value() byte {
	v := byte(sf.Qual) & 0x7f
	if sf.InSelect {
		v |= 0x80
	}
	return v
}

// StatusAndStatusChangeDetection Status and status change detection
// See companion standard 101, subclass 7.2.6.40.
type StatusAndStatusChangeDetection uint32
