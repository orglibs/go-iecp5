package asdu_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

func Test_checkValid(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		attrsLen   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.CheckValid(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.attrsLen); (err != nil) != tt.wantErr {
				t.Errorf("checkValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_single(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.SinglePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendSingle(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("single() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSingle(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.SinglePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.SinglePointInfo{}},
			true,
		},
		{
			"M_SP_NA_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_SP_NA_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x11, 0x02, 0x00, 0x00, 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.SinglePointInfo{
					{0x000001, true, asdu.QDSBlocked, time.Time{}},
					{0x000002, false, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_SP_NA_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_SP_NA_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x11, 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.SinglePointInfo{
					{0x000001, true, asdu.QDSBlocked, time.Time{}},
					{0x000002, false, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.Single(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("Single() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSingleCP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.SinglePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.SinglePointInfo{}},
			true,
		},
		{
			"M_SP_TA_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_SP_TA_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x11}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.SinglePointInfo{
					{0x000001, true, asdu.QDSBlocked, tm0},
					{0x000002, false, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SingleCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("SingleCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSingleCP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.SinglePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.SinglePointInfo{}},
			true,
		},
		{
			"M_SP_TB_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_SP_TB_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x11}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.SinglePointInfo{
					{0x000001, true, asdu.QDSBlocked, tm0},
					{0x000002, false, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SingleCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("SingleCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_double(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.DoublePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendDouble(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("double() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDouble(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.DoublePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.DoublePointInfo{}},
			true,
		},
		{
			"M_DP_NA_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_DP_NA_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x11}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.DoublePointInfo{
					{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, time.Time{}},
					{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_DP_NA_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_DP_NA_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x12, 0x11}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.DoublePointInfo{
					{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, time.Time{}},
					{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.Double(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("Double() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDoubleCP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.DoublePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.DoublePointInfo{}},
			true,
		},
		{
			"M_DP_TA_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_DP_TA_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x12}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x11}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.DoublePointInfo{
					{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, tm0},
					{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.DoubleCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("DoubleCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDoubleCP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.DoublePointInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.DoublePointInfo{}},
			true,
		},
		{
			"M_DP_TB_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_DP_TB_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x12}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x11}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.DoublePointInfo{
					{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, tm0},
					{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.DoubleCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("DoubleCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_step(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.StepPositionInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendStep(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("step() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStep(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.StepPositionInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.StepPositionInfo{}},
			true,
		},
		{
			"M_ST_NA_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ST_NA_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x10, 0x02, 0x00, 0x00, 0x02, 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.StepPositionInfo{
					{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, time.Time{}},
					{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_ST_NA_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ST_NA_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x10, 0x02, 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.StepPositionInfo{
					{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, time.Time{}},
					{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.Step(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("Step() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStepCP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.StepPositionInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.StepPositionInfo{}},
			true,
		},
		{
			"M_ST_TA_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ST_TA_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x10}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.StepPositionInfo{
					{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, tm0},
					{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.StepCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("StepCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStepCP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.StepPositionInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.StepPositionInfo{}},
			true,
		},
		{
			"M_SP_TB_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ST_TB_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x10}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.StepPositionInfo{
					{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, tm0},
					{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.StepCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("StepCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_bitString32(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.BitString32Info
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendBitString32(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("bitString32() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBitString32(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.BitString32Info
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.BitString32Info{}},
			true,
		},
		{
			"M_BO_NA_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_BO_NA_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.BitString32Info{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_BO_NA_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_BO_NA_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x00, 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.BitString32Info{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.BitString32(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("BitString32() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBitString32CP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.BitString32Info
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.BitString32Info{}},
			true,
		},
		{
			"M_BO_TA_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_BO_TA_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.BitString32Info{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.BitString32CP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("BitString32CP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBitString32CP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.BitString32Info
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.BitString32Info{}},
			true,
		},
		{
			"M_BO_TB_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_BO_TB_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.BitString32Info{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.BitString32CP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("BitString32CP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_measuredValueNormal(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		attrs      []asdu.MeasuredValueNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendMeasuredValueNormal(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.attrs...); (err != nil) != tt.wantErr {
				t.Errorf("measuredValueNormal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueNormal(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{}},
			true,
		},
		{
			"M_ME_NA_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NA_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_ME_NA_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NA_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueNormal(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueNormal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueNormalCP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{}},
			true,
		},
		{
			"M_ME_TA_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TA_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueNormalCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueNormalCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueNormalCP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{}},
			true,
		},
		{
			"M_ME_TD_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TD_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueNormalCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueNormalCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueNormalNoQuality(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{}},
			true,
		},
		{
			"M_ME_ND_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_ND_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x02, 0x00}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSGood, time.Time{}},
					{0x000002, 2, asdu.QDSGood, time.Time{}},
				}},
			false,
		},
		{
			"M_ME_ND_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_ND_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueNormalInfo{
					{0x000001, 1, asdu.QDSGood, time.Time{}},
					{0x000002, 2, asdu.QDSGood, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueNormalNoQuality(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueNormalNoQuality() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_measuredValueScaled(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueScaledInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendMeasuredValueScaled(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("measuredValueScaled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueScaled(t *testing.T) {
	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueScaledInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{}},
			true,
		},
		{
			"M_ME_NB_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NB_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_ME_NB_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NB_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{
					{0x000001, 1, asdu.QDSBlocked, time.Time{}},
					{0x000002, 2, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueScaled(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueScaled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueScaledCP24Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueScaledInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{}},
			true,
		},
		{
			"M_ME_TB_1 CP24Time2a  Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TB_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueScaledCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueScaledCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueScaledCP56Time2a(t *testing.T) {
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueScaledInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{}},
			true,
		},
		{
			"M_ME_TE_1 CP56Time2a Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TE_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueScaledInfo{
					{0x000001, 1, asdu.QDSBlocked, tm0},
					{0x000002, 2, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueScaledCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueScaledCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_measuredValueFloat(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueFloatInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SendMeasuredValueFloat(tt.args.c, tt.args.typeID, tt.args.isSequence, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("measuredValueFloat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueFloat(t *testing.T) {
	bits1 := math.Float32bits(100)
	bits2 := math.Float32bits(101)

	type args struct {
		c          asdu.Connect
		isSequence bool
		coa        asdu.CauseOfTransmission
		ca         asdu.CommonAddr
		infos      []asdu.MeasuredValueFloatInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{}},
			true,
		},
		{
			"M_ME_NC_1 seq = false Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NC_1), 0x02, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10,
					0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}),
				false,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{
					{0x000001, 100, asdu.QDSBlocked, time.Time{}},
					{0x000002, 101, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
		{
			"M_ME_NC_1 seq = true Number = 2",
			args{
				newConn(t, []byte{byte(asdu.M_ME_NC_1), 0x82, 0x02, 0x00, 0x34, 0x12,
					0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10,
					byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}),
				true,
				asdu.CauseOfTransmission{Cause: asdu.Background},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{
					{0x000001, 100, asdu.QDSBlocked, time.Time{}},
					{0x000002, 101, asdu.QDSBlocked, time.Time{}},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueFloat(tt.args.c, tt.args.isSequence, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueFloat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueFloatCP24Time2a(t *testing.T) {
	bits1 := math.Float32bits(100)
	bits2 := math.Float32bits(101)

	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueFloatInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{}},
			true,
		},
		{
			"M_ME_TC_1 seq = false Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TC_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10}, tm0CP24Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}, tm0CP24Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{
					{0x000001, 100, asdu.QDSBlocked, tm0},
					{0x000002, 101, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueFloatCP24Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueFloatCP24Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredValueFloatCP56Time2a(t *testing.T) {
	bits1 := math.Float32bits(100)
	bits2 := math.Float32bits(101)
	type args struct {
		c     asdu.Connect
		coa   asdu.CauseOfTransmission
		ca    asdu.CommonAddr
		infos []asdu.MeasuredValueFloatInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid cause",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{}},
			true,
		},
		{
			"M_ME_TF_1 seq = false Number = 2",
			args{
				newConn(t, append(append([]byte{byte(asdu.M_ME_TF_1), 0x02, 0x03, 0x00, 0x34, 0x12},
					append([]byte{0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10}, tm0CP56Time2aBytes...)...),
					append([]byte{0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}, tm0CP56Time2aBytes...)...)),
				asdu.CauseOfTransmission{Cause: asdu.Spontaneous},
				0x1234,
				[]asdu.MeasuredValueFloatInfo{
					{0x000001, 100, asdu.QDSBlocked, tm0},
					{0x000002, 101, asdu.QDSBlocked, tm0},
				}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.MeasuredValueFloatCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.infos...); (err != nil) != tt.wantErr {
				t.Errorf("MeasuredValueFloatCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASDU_GetSinglePoint(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.SinglePointInfo
	}{
		{
			"M_SP_NA_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_SP_NA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x11, 0x02, 0x00, 0x00, 0x10}},
			[]asdu.SinglePointInfo{
				{0x000001, true, asdu.QDSBlocked, time.Time{}},
				{0x000002, false, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_SP_NA_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_SP_NA_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x11, 0x10}},
			[]asdu.SinglePointInfo{
				{0x000001, true, asdu.QDSBlocked, time.Time{}},
				{0x000002, false, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_SP_TB_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_SP_TB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x11}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.SinglePointInfo{
				{0x000001, true, asdu.QDSBlocked, tm0},
				{0x000002, false, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetSinglePoint()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetSinglePoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetSinglePointCP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.SinglePointInfo
	}{
		{
			"M_SP_TA_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_SP_TA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x11}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.SinglePointInfo{
				{0x000001, true, asdu.QDSBlocked, tm0},
				{0x000002, false, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetSinglePoint()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetDoublePoint(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.DoublePointInfo
	}{
		{
			"M_DP_NA_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_DP_NA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x11}},
			[]asdu.DoublePointInfo{
				{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, time.Time{}},
				{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_DP_NA_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_DP_NA_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x12, 0x11}},
			[]asdu.DoublePointInfo{
				{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, time.Time{}},
				{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_DP_TB_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_DP_TB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x12}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x11}, tm0CP56Time2aBytes...)...)},
			[]asdu.DoublePointInfo{
				{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, tm0},
				{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetDoublePoint(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetDoublePoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetDoublePointCP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.DoublePointInfo
	}{

		{
			"M_DP_TA_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_DP_TA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x12}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x11}, tm0CP24Time2aBytes...)...)},
			[]asdu.DoublePointInfo{
				{0x000001, asdu.DPIDeterminedOn, asdu.QDSBlocked, tm0},
				{0x000002, asdu.DPIDeterminedOff, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetDoublePoint()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetStepPosition(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.StepPositionInfo
	}{
		{
			"M_ST_NA_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ST_NA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x10, 0x02, 0x00, 0x00, 0x02, 0x10}},
			[]asdu.StepPositionInfo{
				{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, time.Time{}},
				{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ST_NA_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ST_NA_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x10, 0x02, 0x10}},
			[]asdu.StepPositionInfo{
				{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, time.Time{}},
				{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ST_TB_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ST_TB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x10}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.StepPositionInfo{
				{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, tm0},
				{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetStepPosition(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetStepPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetStepPositionCP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.StepPositionInfo
	}{

		{
			"M_ST_TA_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ST_TA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x10}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.StepPositionInfo{
				{0x000001, asdu.StepPosition{Val: 0x01}, asdu.QDSBlocked, tm0},
				{0x000002, asdu.StepPosition{Val: 0x02}, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetStepPosition()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetBitString32(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.BitString32Info
	}{
		{
			"M_BO_NA_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_BO_NA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}},
			[]asdu.BitString32Info{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_BO_NA_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_BO_NA_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x00, 0x10}},
			[]asdu.BitString32Info{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_BO_TB_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_BO_TB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.BitString32Info{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetBitString32(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetBitString32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetBitString32CP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.BitString32Info
	}{
		{
			"M_BO_TA_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_BO_TA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.BitString32Info{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetBitString32()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetMeasuredValueNormal(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueNormalInfo
	}{
		{
			"M_ME_NA_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x10}},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_NA_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NA_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x10}},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_TD_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TD_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
		{
			"M_ME_ND_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_ND_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x02, 0x00}},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSGood, time.Time{}},
				{0x000002, 2, asdu.QDSGood, time.Time{}}},
		},
		{
			"M_ME_ND_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_ND_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00}},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSGood, time.Time{}},
				{0x000002, 2, asdu.QDSGood, time.Time{}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetMeasuredValueNormal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetMeasuredValueNormal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetMeasuredValueNormalCP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueNormalInfo
	}{

		{
			"M_ME_TA_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TA_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.MeasuredValueNormalInfo{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetMeasuredValueNormal()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetMeasuredValueScaled(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueScaledInfo
	}{
		{
			"M_ME_NB_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0x02, 0x00, 0x10}},
			[]asdu.MeasuredValueScaledInfo{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_NB_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NB_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10, 0x02, 0x00, 0x10}},
			[]asdu.MeasuredValueScaledInfo{
				{0x000001, 1, asdu.QDSBlocked, time.Time{}},
				{0x000002, 2, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_TE_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TE_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.MeasuredValueScaledInfo{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetMeasuredValueScaled(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetMeasuredValueScaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetMeasuredValueScaledCP24Time2a(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueScaledInfo
	}{
		{
			"M_ME_TB_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TB_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x10}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, 0x02, 0x00, 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.MeasuredValueScaledInfo{
				{0x000001, 1, asdu.QDSBlocked, tm0},
				{0x000002, 2, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetMeasuredValueScaled()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}

func TestASDU_GetMeasuredValueFloat(t *testing.T) {
	bits1 := math.Float32bits(100)
	bits2 := math.Float32bits(101)
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueFloatInfo
	}{
		{
			"M_ME_NC_1 seq = false Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NC_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				[]byte{
					0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10,
					0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}},
			[]asdu.MeasuredValueFloatInfo{
				{0x000001, 100, asdu.QDSBlocked, time.Time{}},
				{0x000002, 101, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_NC_1 seq = true Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_NC_1,
					Variable: asdu.VariableStruct{IsSequence: true, Number: 2}},
				[]byte{
					0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10,
					byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}},
			[]asdu.MeasuredValueFloatInfo{
				{0x000001, 100, asdu.QDSBlocked, time.Time{}},
				{0x000002, 101, asdu.QDSBlocked, time.Time{}}},
		},
		{
			"M_ME_TF_1 CP56Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TF_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10}, tm0CP56Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}, tm0CP56Time2aBytes...)...)},
			[]asdu.MeasuredValueFloatInfo{
				{0x000001, 100, asdu.QDSBlocked, tm0},
				{0x000002, 101, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			if got := this.GetMeasuredValueFloat(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetMeasuredValueFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetMeasuredValueFloatCP24Time2a(t *testing.T) {
	bits1 := math.Float32bits(100)
	bits2 := math.Float32bits(101)
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []asdu.MeasuredValueFloatInfo
	}{
		{
			"M_ME_TC_1 CP24Time2a  Number = 2",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					Type:     asdu.M_ME_TC_1,
					Variable: asdu.VariableStruct{IsSequence: false, Number: 2}},
				append(append([]byte{0x01, 0x00, 0x00, byte(bits1), byte(bits1 >> 8), byte(bits1 >> 16), byte(bits1 >> 24), 0x10}, tm0CP24Time2aBytes...),
					append([]byte{0x02, 0x00, 0x00, byte(bits2), byte(bits2 >> 8), byte(bits2 >> 16), byte(bits2 >> 24), 0x10}, tm0CP24Time2aBytes...)...)},
			[]asdu.MeasuredValueFloatInfo{
				{0x000001, 100, asdu.QDSBlocked, tm0},
				{0x000002, 101, asdu.QDSBlocked, tm0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := this.GetMeasuredValueFloat()
			for i, v := range got {
				isError := false
				if !reflect.DeepEqual(v.Ioa, tt.want[i].Ioa) {
					isError = true
				}
				if !reflect.DeepEqual(v.Value, tt.want[i].Value) {
					isError = true
				}
				if !reflect.DeepEqual(v.Qds, tt.want[i].Qds) {
					isError = true
				}
				if v.Time.Second() != tt.want[i].Time.Second() {
					isError = true
				}
				if v.Time.Nanosecond() != tt.want[i].Time.Nanosecond() {
					isError = true
				}
				if isError {
					t.Errorf("ASDU.GetSinglePoint() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}
