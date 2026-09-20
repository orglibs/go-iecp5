// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"fmt"

	"github.com/orglibs/go-iecp5/asdu"
)

const StartFrame byte = 0x68 // start frame

// APDU form Max size 255
//
//	|              APCI                   |       ASDU         |
//	| start | APDU length | control field |       ASDU         |
//	                 |          APDU field size(253)           |
//
// bytes|    1  |    1   |        4           |                    |
const (
	APCICtlFiledSize = 4 // control filed(4)

	APDUSizeMax      = 255                                 // start(1) + length(1) + control field(4) + ASDU
	APDUFieldSizeMax = APCICtlFiledSize + asdu.ASDUSizeMax // control field(4) + ASDU
)

// U-frame Control Field Function
const (
	UStartDtActive  byte = 4 << iota // Startup activation 0x04
	UStartDtConfirm                  // Startup Confirmation 0x08
	UStopDtActive                    // Stop activation 0x10
	UStopDtConfirm                   // Stop Confirmation 0x20
	UTestFrActive                    // Test activation 0x40
	UTestFrConfirm                   // Test Confirmation 0x80
)

// I-frame Contains apci and asdu information frames. Information transmission for numbering information
type IAPCI struct {
	SendSN, RcvSN uint16
}

func (sf IAPCI) String() string {
	return fmt.Sprintf("I[sendNO: %d, recvNO: %d]", sf.SendSN, sf.RcvSN)
}

// S-frames are used primarily to confirm the correct transmission of frames, and are called supervisory by the protocol.
type SAPCI struct {
	RcvSN uint16
}

func (sf SAPCI) String() string {
	return fmt.Sprintf("S[recvNO: %d]", sf.RcvSN)
}

// U-frame apci only Unnumbered control information
type UAPCI struct {
	Function byte // bit8 Test confirmation
}

func (sf UAPCI) String() string {
	var s string
	switch sf.Function {
	case UStartDtActive:
		s = "StartDtActive"
	case UStartDtConfirm:
		s = "StartDtConfirm"
	case UStopDtActive:
		s = "StopDtActive"
	case UStopDtConfirm:
		s = "StopDtConfirm"
	case UTestFrActive:
		s = "TestFrActive"
	case UTestFrConfirm:
		s = "TestFrConfirm"
	default:
		s = "Unknown"
	}

	return fmt.Sprintf("U[function: %s]", s)
}

// newIFrame Creates an I-frame and returns his corresponding apdu.
func NewIFrame(sendSN, rcvSN uint16, asdus []byte) ([]byte, error) {
	if len(asdus) > asdu.ASDUSizeMax {
		return nil, fmt.Errorf("ASDU filed large than max %d", asdu.ASDUSizeMax)
	}

	b := make([]byte, len(asdus)+6)

	b[0] = StartFrame
	b[1] = byte(len(asdus) + 4)
	b[2] = byte(sendSN << 1)
	b[3] = byte(sendSN >> 7)
	b[4] = byte(rcvSN << 1)
	b[5] = byte(rcvSN >> 7)
	copy(b[6:], asdus)

	return b, nil
}

// newSFrame createSFrame and returns his apdu
func NewSFrame(rcvSN uint16) []byte {
	return []byte{StartFrame, 4, 0x01, 0x00, byte(rcvSN << 1), byte(rcvSN >> 7)}
}

// newUFrame Creates a U-frame and returns his apdu.
func NewUFrame(which byte) []byte {
	return []byte{StartFrame, 4, which | 0x03, 0x00, 0x00, 0x00}
}

// apci application protocol control information
type APCI struct {
	start                  byte
	apduFiledLen           byte // control + asdu lengths
	ctr1, ctr2, ctr3, ctr4 byte
}

// Parse 保留 circutor 原有双返回值接口；非法报文返回 nil, nil。
// 需要错误原因的调用方应使用 ParseChecked。任何失败都不能驱动链路状态变化。
func Parse(apdu []byte) (interface{}, []byte) {
	header, data, err := ParseChecked(apdu)
	if err != nil {
		return nil, nil
	}
	return header, data
}

// ParseChecked 严格校验 APCI 长度、保留位和 U 帧功能位。
// 移植自 riclolsen v0.4.4；保持本分支 IAPCI/SAPCI/UAPCI 的导出名称。
func ParseChecked(apdu []byte) (interface{}, []byte, error) {
	if len(apdu) > APDUSizeMax {
		return nil, nil, ErrInvalidAPCI
	}
	if len(apdu) < 6 {
		return nil, nil, fmt.Errorf("%w: %d octets, minimum 6", ErrInvalidAPCI, len(apdu))
	}
	if apdu[0] != StartFrame {
		return nil, nil, ErrInvalidAPCI
	}
	apci := APCI{apdu[0], apdu[1], apdu[2], apdu[3], apdu[4], apdu[5]}

	// The length octet counts the control field and the ASDU, so it must
	// describe exactly what was read.
	if int(apci.apduFiledLen) != len(apdu)-2 {
		return nil, nil, fmt.Errorf("%w: length field %d does not match %d octets read",
			ErrInvalidAPCI, apci.apduFiledLen, len(apdu)-2)
	}

	switch {
	case apci.ctr1&0x01 == 0: // I format
		// An I frame carries an ASDU; one without a payload has nothing to
		// say. The low bit of the third octet is the format bit of the
		// receive sequence number and is always zero.
		if apci.apduFiledLen <= APCICtlFiledSize {
			return nil, nil, fmt.Errorf("%w: I format with no ASDU", ErrInvalidAPCI)
		}
		if apci.ctr3&0x01 != 0 {
			return nil, nil, fmt.Errorf("%w: I format with control octet 3 = 0x%02X, bit 0 must be clear",
				ErrInvalidAPCI, apci.ctr3)
		}
		return IAPCI{
			SendSN: uint16(apci.ctr1)>>1 + uint16(apci.ctr2)<<7,
			RcvSN:  uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:], nil

	case apci.ctr1&0x03 == 0x01: // S format
		if apci.apduFiledLen != APCICtlFiledSize {
			return nil, nil, fmt.Errorf("%w: S format with length %d, must be %d",
				ErrInvalidAPCI, apci.apduFiledLen, APCICtlFiledSize)
		}
		if apci.ctr1 != 0x01 || apci.ctr2 != 0 || apci.ctr3&0x01 != 0 {
			return nil, nil, fmt.Errorf("%w: S format with control field %02X %02X %02X %02X, unused bits must be clear",
				ErrInvalidAPCI, apci.ctr1, apci.ctr2, apci.ctr3, apci.ctr4)
		}
		return SAPCI{
			RcvSN: uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:], nil

	default: // U format, apci.ctr1&0x03 == 0x03
		if apci.apduFiledLen != APCICtlFiledSize {
			return nil, nil, fmt.Errorf("%w: U format with length %d, must be %d",
				ErrInvalidAPCI, apci.apduFiledLen, APCICtlFiledSize)
		}
		if apci.ctr2 != 0 || apci.ctr3 != 0 || apci.ctr4 != 0 {
			return nil, nil, fmt.Errorf("%w: U format with reserved octets %02X %02X %02X, must be zero",
				ErrInvalidAPCI, apci.ctr2, apci.ctr3, apci.ctr4)
		}
		function := apci.ctr1 & 0xfc
		if !isValidUFunction(function) {
			return nil, nil, fmt.Errorf("%w: U format function 0x%02X is not one of the six defined",
				ErrInvalidAPCI, function)
		}
		return UAPCI{Function: function}, apdu[6:], nil
	}
}

func isValidUFunction(function byte) bool {
	switch function {
	case UStartDtActive, UStartDtConfirm, UStopDtActive, UStopDtConfirm, UTestFrActive, UTestFrConfirm:
		return true
	default:
		return false
	}
}
