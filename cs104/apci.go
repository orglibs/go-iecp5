// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"fmt"

	"gitlab.com/circutor-library/go-iecp5/asdu"
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
func NewIFrame(sendSN, RcvSN uint16, asdus []byte) ([]byte, error) {
	if len(asdus) > asdu.ASDUSizeMax {
		return nil, fmt.Errorf("ASDU filed large than max %d", asdu.ASDUSizeMax)
	}

	b := make([]byte, len(asdus)+6)

	b[0] = StartFrame
	b[1] = byte(len(asdus) + 4)
	b[2] = byte(sendSN << 1)
	b[3] = byte(sendSN >> 7)
	b[4] = byte(RcvSN << 1)
	b[5] = byte(RcvSN >> 7)
	copy(b[6:], asdus)

	return b, nil
}

// newSFrame createSFrame and returns his apdu
func NewSFrame(RcvSN uint16) []byte {
	return []byte{StartFrame, 4, 0x01, 0x00, byte(RcvSN << 1), byte(RcvSN >> 7)}
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

// return frame type , APCI, remain data
func Parse(apdu []byte) (interface{}, []byte) {
	apci := APCI{apdu[0], apdu[1], apdu[2], apdu[3], apdu[4], apdu[5]}
	if apci.ctr1&0x01 == 0 {
		return IAPCI{
			SendSN: uint16(apci.ctr1)>>1 + uint16(apci.ctr2)<<7,
			RcvSN:  uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:]
	}
	if apci.ctr1&0x03 == 0x01 {
		return SAPCI{
			RcvSN: uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:]
	}
	// apci.ctrl&0x03 == 0x03
	return UAPCI{
		Function: apci.ctr1 & 0xfc,
	}, apdu[6:]
}
