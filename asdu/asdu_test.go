package asdu_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/orglibs/go-iecp5/asdu"
)

func TestParams_Valid(t *testing.T) {
	tests := []struct {
		name    string
		this    *asdu.Params
		wantErr bool
	}{
		{"invalid", &asdu.Params{}, true},
		{"ParamsNarrow", asdu.ParamsNarrow, false},
		{"ParamsWide", asdu.ParamsWide, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.this.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("Params.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParams_ValidCommonAddr(t *testing.T) {
	type args struct {
		addr asdu.CommonAddr
	}
	tests := []struct {
		name    string
		this    *asdu.Params
		args    args
		wantErr bool
	}{
		{"common address zero", asdu.ParamsNarrow, args{asdu.InvalidCommonAddr}, true},
		{"common address size(1),invalid", asdu.ParamsNarrow, args{256}, true},
		{"common address size(1),valid", asdu.ParamsNarrow, args{255}, false},
		{"common address size(2),valid", asdu.ParamsWide, args{65535}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.this.ValidCommonAddr(tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Params.ValidCommonAddr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParams_IdentifierSize(t *testing.T) {
	tests := []struct {
		name string
		this *asdu.Params
		want int
	}{
		{"ParamsNarrow(4)", asdu.ParamsNarrow, 4},
		{"ParamsWide(6)", asdu.ParamsWide, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.IdentifierSize(); got != tt.want {
				t.Errorf("Params.IdentifierSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_SetVariableNumber(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		InfoObj    []byte
		bootstrap  [asdu.ASDUSizeMax]byte
	}
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.InfoObj,
				Bootstrap:  tt.fields.bootstrap,
			}
			if err := this.SetVariableNumber(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("ASDU.SetVariableNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASDU_Reply(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		InfoObj    []byte
		bootstrap  [asdu.ASDUSizeMax]byte
	}
	type args struct {
		c    asdu.CauseOfTransmission
		addr asdu.CommonAddr
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *asdu.ASDU
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:     tt.fields.Params,
				Identifier: tt.fields.Identifier,
				InfoObj:    tt.fields.InfoObj,
				Bootstrap:  tt.fields.bootstrap,
			}
			if got := this.Reply(tt.args.c, tt.args.addr, 0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ASDU.Reply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_MarshalBinary(t *testing.T) {
	type fields struct {
		Params     *asdu.Params
		Identifier asdu.Identifier
		InfoObj    []byte
	}
	tests := []struct {
		name     string
		fields   fields
		wantData []byte
		wantErr  bool
	}{
		{
			"unused cause",
			fields{
				asdu.ParamsNarrow,
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Unused},
					0,
					0x80},
				nil},
			nil,
			true,
		},
		{
			"invalid cause size",
			fields{
				&asdu.Params{CauseSize: 0, CommonAddrSize: 1, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC},
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					0x80},
				nil},
			nil,
			true,
		},
		{
			"cause size(1),but origAddress not equal zero",
			fields{
				&asdu.Params{CauseSize: 1, CommonAddrSize: 1, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC},
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					1,
					0x80},
				nil},
			nil,
			true,
		},
		{
			"invalid common address",
			fields{
				&asdu.Params{CauseSize: 1, CommonAddrSize: 1, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC},
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					asdu.InvalidCommonAddr},
				nil},
			nil,
			true},
		{
			"invalid common address size",
			fields{
				&asdu.Params{CauseSize: 1, CommonAddrSize: 0, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC},
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					0x80},
				nil},
			nil,
			true,
		},
		{
			"common size(1),but common address equal 255",
			fields{
				&asdu.Params{CauseSize: 1, CommonAddrSize: 1, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC},
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					255},
				nil},
			nil,
			true,
		},
		{
			"ParamsNarrow",
			fields{
				asdu.ParamsNarrow,
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{Number: 2},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					0x80},
				[]byte{0x00, 0x01, 0x02, 0x03}},
			[]byte{0x01, 0x02, 0x06, 0x80, 0x00, 0x01, 0x02, 0x03},
			false,
		},
		{
			"ParamsNarrow global address",
			fields{
				asdu.ParamsNarrow,
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{Number: 2},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					asdu.GlobalCommonAddr},
				[]byte{0x00, 0x01, 0x02, 0x03}},
			[]byte{0x01, 0x02, 0x06, 0xff, 0x00, 0x01, 0x02, 0x03},
			false,
		},
		{
			"ParamsWide",
			fields{
				asdu.ParamsWide,
				asdu.Identifier{
					asdu.M_SP_NA_1,
					asdu.VariableStruct{Number: 1},
					asdu.CauseOfTransmission{Cause: asdu.Activation},
					0,
					0x6080},
				[]byte{0x00, 0x01, 0x02, 0x03}},
			[]byte{0x01, 0x01, 0x06, 0x00, 0x80, 0x60, 0x00, 0x01, 0x02, 0x03},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := asdu.NewASDU(tt.fields.Params, tt.fields.Identifier)
			this.InfoObj = append(this.InfoObj, tt.fields.InfoObj...)

			gotData, err := this.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("ASDU.MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("ASDU.MarshalBinary() = % x, want % x", gotData, tt.wantData)
			}
		})
	}
}

func TestASDU_UnmarshalBinary(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		Params  *asdu.Params
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"invalid param",
			&asdu.Params{},
			args{}, // 125
			[]byte{},
			true,
		},
		{
			"less than data unit identifier size",
			asdu.ParamsWide,
			args{[]byte{0x0b, 0x01, 0x06, 0x80}},
			[]byte{},
			true,
		},
		{
			"type id fix size error",
			asdu.ParamsWide,
			args{[]byte{0x07d, 0x01, 0x06, 0x00, 0x80, 0x60}},
			[]byte{},
			true,
		},

		{
			"ParamsNarrow global address",
			asdu.ParamsNarrow,
			args{[]byte{0x0b, 0x01, 0x06, 0x80, 0x00, 0x01, 0x02, 0x03}},
			[]byte{0x00, 0x01, 0x02, 0x03},
			false,
		},
		{
			"ParamsNarrow",
			asdu.ParamsNarrow,
			args{[]byte{0x0b, 0x01, 0x06, 0xff, 0x00, 0x01, 0x02, 0x03}},
			[]byte{0x00, 0x01, 0x02, 0x03},
			false,
		},
		{
			"ParamsWide",
			asdu.ParamsWide,
			args{[]byte{0x01, 0x01, 0x06, 0x00, 0x80, 0x60, 0x00, 0x01, 0x02, 0x03}},
			[]byte{0x00, 0x01, 0x02, 0x03},
			false,
		},
		{
			"ParamsWide sequence",
			asdu.ParamsWide,
			args{[]byte{0x01, 0x81, 0x06, 0x00, 0x80, 0x60, 0x00, 0x01, 0x02, 0x03}},
			[]byte{0x00, 0x01, 0x02, 0x03},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := asdu.NewEmptyASDU(tt.Params)
			if err := this.UnmarshalBinary(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ASDU.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(this.InfoObj, tt.want) {
				t.Errorf("ASDU.UnmarshalBinary() got % x, want % x", this.InfoObj, tt.want)
			}
		})
	}
}
