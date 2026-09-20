// 异常控制帧测试移植自 riclolsen v0.4.4。
package cs104

import (
	"errors"
	"testing"
)

func Test_parseRejectsMalformedAPCI(t *testing.T) {
	for _, tt := range []struct {
		name string
		apdu []byte
	}{
		{"truncated", []byte{StartFrame, 0x04, 0x07}},
		{"length field disagrees with the octets read",
			[]byte{StartFrame, 0x08, 0x01, 0x00, 0x02, 0x00}},

		{"U format with reserved octets set",
			[]byte{StartFrame, 0x04, 0x07, 0xFF, 0xFF, 0xFF}},
		{"U format with one reserved octet set",
			[]byte{StartFrame, 0x04, 0x07, 0x00, 0x01, 0x00}},
		{"U format with length 10",
			[]byte{StartFrame, 0x0A, 0x07, 0x00, 0x00, 0x00,
				0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}},
		{"U format with two function bits set",
			[]byte{StartFrame, 0x04, 0x0F, 0x00, 0x00, 0x00}},
		{"U format with no function bit set",
			[]byte{StartFrame, 0x04, 0x03, 0x00, 0x00, 0x00}},

		{"S format with second control octet set",
			[]byte{StartFrame, 0x04, 0x01, 0xFF, 0x02, 0x00}},
		{"S format with unused bits of the first octet set",
			[]byte{StartFrame, 0x04, 0xF1, 0x00, 0x02, 0x00}},
		{"S format with the sequence format bit set",
			[]byte{StartFrame, 0x04, 0x01, 0x00, 0x03, 0x00}},
		{"S format with length 8",
			[]byte{StartFrame, 0x08, 0x01, 0x00, 0x02, 0x00, 0xAA, 0xAA, 0xAA, 0xAA}},

		{"I format with no ASDU",
			[]byte{StartFrame, 0x04, 0x02, 0x00, 0x02, 0x00}},
		{"I format with the sequence format bit set",
			[]byte{StartFrame, 0x05, 0x02, 0x00, 0x03, 0x00, 0x64}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			apci, _, err := ParseChecked(tt.apdu)
			if err == nil {
				t.Fatalf("accepted as %T: % x", apci, tt.apdu)
			}
			if !errors.Is(err, ErrInvalidAPCI) {
				t.Fatalf("wrong error: %v", err)
			}
		})
	}
}
