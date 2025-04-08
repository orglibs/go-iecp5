package asdu_test

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"testing"
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

type conn struct {
	p    *asdu.Params
	want []byte
	t    *testing.T
}

func newConn(t *testing.T, want []byte) *conn {
	t.Helper()

	return &conn{asdu.ParamsWide, want, t}
}

func (sf *conn) Params() *asdu.Params     { return sf.p }
func (sf *conn) UnderlyingConn() net.Conn { return nil }

// Send
func (sf *conn) Send(u *asdu.ASDU) error {
	data, err := u.MarshalBinary()
	if err != nil {
		return fmt.Errorf("Send asdu failed:%w", err)
	}

	if !reflect.DeepEqual(sf.want, data) {
		sf.t.Errorf("Send() out = % x, want % x", data, sf.want)
	}

	return nil
}

func (sf *conn) SendQueuedASDU(u *asdu.ASDU) error {
	return nil
}

func TestSingleCmd(t *testing.T) {
	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.SingleCommandInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SingleCommandInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_SC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.SingleCommandInfo{}},
			true},
		{
			"C_SC_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_SC_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x05}),
				asdu.C_SC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SingleCommandInfo{
					0x567890,
					true,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					time.Time{}}},
			false},
		{
			"C_SC_TA_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_SC_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x04}, tm0CP56Time2aBytes...)),
				asdu.C_SC_TA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SingleCommandInfo{
					0x567890, false,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					tm0}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SingleCmd(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("SingleCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDoubleCmd(t *testing.T) {
	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.DoubleCommandInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.DoubleCommandInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_DC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.DoubleCommandInfo{}},
			true},
		{
			"C_DC_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_DC_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x05}),
				asdu.C_DC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.DoubleCommandInfo{
					0x567890,
					asdu.DCOOn,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					time.Time{}}},
			false},
		{
			"C_DC_TA_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_DC_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x06}, tm0CP56Time2aBytes...)),
				asdu.C_DC_TA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.DoubleCommandInfo{
					0x567890,
					asdu.DCOOff,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.DoubleCmd(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("DoubleCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStepCmd(t *testing.T) {
	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.StepCommandInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.StepCommandInfo{}},
			true},
		{
			"cause not Activation and Deactivation", args{
				newConn(t, nil),
				asdu.C_RC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.StepCommandInfo{}},
			true},
		{
			"C_RC_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_RC_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x05}),
				asdu.C_RC_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.StepCommandInfo{
					0x567890,
					asdu.SCOStepDown,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					time.Time{}}},
			false},
		{
			"C_RC_TA_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_RC_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x06}, tm0CP56Time2aBytes...)),
				asdu.C_RC_TA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.StepCommandInfo{
					0x567890,
					asdu.SCOStepUP,
					asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.StepCmd(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("StepCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetpointCmdNormal(t *testing.T) {
	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.SetpointCommandNormalInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandNormalInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_SE_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.SetpointCommandNormalInfo{}},
			true},
		{
			"C_SE_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_SE_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x01}),
				asdu.C_SE_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandNormalInfo{
					0x567890,
					100,
					asdu.QualifierOfSetpointCmd{1, false},
					time.Time{}}},
			false},
		{
			"C_SE_TA_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_SE_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x01}, tm0CP56Time2aBytes...)),
				asdu.C_SE_TA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandNormalInfo{
					0x567890, 100,
					asdu.QualifierOfSetpointCmd{1, false},
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SetpointCmdNormal(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("SetpointCmdNormal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetpointCmdScaled(t *testing.T) {
	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.SetpointCommandScaledInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandScaledInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_SE_NB_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.SetpointCommandScaledInfo{}},
			true},
		{
			"C_SE_NB_1",
			args{
				newConn(t, []byte{byte(asdu.C_SE_NB_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x01}),
				asdu.C_SE_NB_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandScaledInfo{
					0x567890,
					100,
					asdu.QualifierOfSetpointCmd{1, false},
					time.Time{}}},
			false},
		{
			"C_SE_TB_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_SE_TB_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x01}, tm0CP56Time2aBytes...)),
				asdu.C_SE_TB_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandScaledInfo{
					0x567890, 100,
					asdu.QualifierOfSetpointCmd{1, false},
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SetpointCmdScaled(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("SetpointCmdScaled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetpointCmdFloat(t *testing.T) {
	bits := math.Float32bits(100)

	type args struct {
		c      asdu.Connect
		typeID asdu.TypeID
		coa    asdu.CauseOfTransmission
		ca     asdu.CommonAddr
		cmd    asdu.SetpointCommandFloatInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandFloatInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_SE_NC_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.SetpointCommandFloatInfo{}},
			true},
		{
			"C_SE_NC_1",
			args{
				newConn(t, []byte{byte(asdu.C_SE_NC_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24), 0x01}),
				asdu.C_SE_NC_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandFloatInfo{
					0x567890,
					100,
					asdu.QualifierOfSetpointCmd{1, false},
					time.Time{}}},
			false},
		{
			"C_SE_TC_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_SE_TC_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24), 0x01}, tm0CP56Time2aBytes...)),
				asdu.C_SE_TC_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.SetpointCommandFloatInfo{
					0x567890, 100,
					asdu.QualifierOfSetpointCmd{1, false},
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.SetpointCmdFloat(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.ca, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("SetpointCmdFloat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBitsString32Cmd(t *testing.T) {
	type args struct {
		c          asdu.Connect
		typeID     asdu.TypeID
		coa        asdu.CauseOfTransmission
		commonAddr asdu.CommonAddr
		cmd        asdu.BitsString32CommandInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid type id",
			args{
				newConn(t, nil),
				0,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.BitsString32CommandInfo{}},
			true},
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.C_BO_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.BitsString32CommandInfo{}},
			true},
		{
			"C_BO_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_BO_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x00, 0x00}),
				asdu.C_BO_NA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.BitsString32CommandInfo{
					0x567890,
					100,
					time.Time{}}},
			false},
		{
			"C_BO_TA_1 CP56Time2a",
			args{
				newConn(t, append([]byte{byte(asdu.C_BO_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x64, 0x00, 0x00, 0x00}, tm0CP56Time2aBytes...)),
				asdu.C_BO_TA_1,
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.BitsString32CommandInfo{
					0x567890, 100,
					tm0}},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.BitsString32Cmd(tt.args.c, tt.args.typeID, tt.args.coa, tt.args.commonAddr, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("BitsString32Cmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASDU_GetSingleCmd(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.SingleCommandInfo
	}{
		{
			"C_SC_NA_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SC_NA_1},
				[]byte{0x90, 0x78, 0x56, 0x05}},
			asdu.SingleCommandInfo{
				0x567890,
				true,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				time.Time{}},
		},
		{
			"C_SC_TA_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SC_TA_1},
				append([]byte{0x90, 0x78, 0x56, 0x04}, tm0CP56Time2aBytes...)},
			asdu.SingleCommandInfo{
				0x567890,
				false,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetSingleCmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetSingleCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetDoubleCmd(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.DoubleCommandInfo
	}{
		{
			"C_DC_NA_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_DC_NA_1},
				[]byte{0x90, 0x78, 0x56, 0x05}},
			asdu.DoubleCommandInfo{
				0x567890,
				asdu.DCOOn,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				time.Time{},
			},
		},
		{
			"C_DC_TA_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_DC_TA_1},
				append([]byte{0x90, 0x78, 0x56, 0x06}, tm0CP56Time2aBytes...)},
			asdu.DoubleCommandInfo{
				0x567890,
				asdu.DCOOff,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetDoubleCmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetDoubleCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetStepCmd(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.StepCommandInfo
	}{
		{
			"C_RC_NA_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_RC_NA_1},
				[]byte{0x90, 0x78, 0x56, 0x05}},
			asdu.StepCommandInfo{
				0x567890,
				asdu.SCOStepDown,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				time.Time{}},
		},
		{
			"C_RC_TA_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_RC_TA_1},
				append([]byte{0x90, 0x78, 0x56, 0x06}, tm0CP56Time2aBytes...)},
			asdu.StepCommandInfo{
				0x567890,
				asdu.SCOStepUP,
				asdu.QualifierOfCommand{asdu.QOCShortPulseDuration, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetStepCmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetStepCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetSetpointNormalCmd(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.SetpointCommandNormalInfo
	}{
		{
			"C_SE_NA_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_NA_1},
				[]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x01}},
			asdu.SetpointCommandNormalInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				time.Time{}},
		},
		{
			"C_SE_TA_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_TA_1},
				append([]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x01}, tm0CP56Time2aBytes...)},
			asdu.SetpointCommandNormalInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetSetpointNormalCmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetSetpointNormalCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetSetpointCmdScaled(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.SetpointCommandScaledInfo
	}{
		{
			"C_SE_NB_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_NB_1},
				[]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x01}},
			asdu.SetpointCommandScaledInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				time.Time{}},
		},
		{
			"C_SE_TB_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_TB_1},
				append([]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x01}, tm0CP56Time2aBytes...)},
			asdu.SetpointCommandScaledInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetSetpointCmdScaled()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetSetpointCmdScaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetSetpointFloatCmd(t *testing.T) {
	bits := math.Float32bits(100)

	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.SetpointCommandFloatInfo
	}{
		{
			"C_SE_NC_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_NC_1},
				[]byte{0x90, 0x78, 0x56, byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24), 0x01}},
			asdu.SetpointCommandFloatInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				time.Time{}},
		},
		{
			"C_SE_TC_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_SE_TC_1},
				append([]byte{0x90, 0x78, 0x56, byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24), 0x01}, tm0CP56Time2aBytes...)},
			asdu.SetpointCommandFloatInfo{
				0x567890,
				100,
				asdu.QualifierOfSetpointCmd{1, false},
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetSetpointFloatCmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetSetpointFloatCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetBitsString32Cmd(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		infoObj    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.BitsString32CommandInfo
	}{
		{
			"C_BO_NA_1",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_BO_NA_1},
				[]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x00, 0x00}},
			asdu.BitsString32CommandInfo{
				0x567890,
				100,
				time.Time{}},
		},
		{
			"C_BO_TA_1 CP56Time2a",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{Type: asdu.C_BO_TA_1},
				append([]byte{0x90, 0x78, 0x56, 0x64, 0x00, 0x00, 0x00}, tm0CP56Time2aBytes...)},
			asdu.BitsString32CommandInfo{
				0x567890,
				100,
				tm0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.infoObj,
			}
			got := sf.GetBitsString32Cmd()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.GetBitsString32Cmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
