package asdu_test

import (
	"math"
	"reflect"
	"testing"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

func TestSinglePoint_Value(t *testing.T) {
	tests := []struct {
		name string
		this asdu.SinglePoint
		want byte
	}{
		{"off", asdu.SPIOff, 0x00},
		{"on", asdu.SPIOn, 0x01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Value(); got != tt.want {
				t.Errorf("SinglePoint.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoublePoint_Value(t *testing.T) {
	tests := []struct {
		name string
		this asdu.DoublePoint
		want byte
	}{
		{"IndeterminateOrIntermediate", asdu.DPIIndeterminateOrIntermediate, 0x00},
		{"DeterminedOff", asdu.DPIDeterminedOff, 0x01},
		{"DeterminedOn", asdu.DPIDeterminedOn, 0x02},
		{"Indeterminate", asdu.DPIIndeterminate, 0x03},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Value(); got != tt.want {
				t.Errorf("DoublePoint.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseStepPosition(t *testing.T) {
	type args struct {
		value byte
	}
	tests := []struct {
		name string
		args args
		want asdu.StepPosition
	}{
		{"Value 0xc0 In transient state", args{0xc0}, asdu.StepPosition{-64, true}},
		{"Value 0x40 Not in transient state", args{0x40}, asdu.StepPosition{-64, false}},
		{"Value 0x87 In transient state", args{0x87}, asdu.StepPosition{0x07, true}},
		{"Value 0x07 Not in transient state", args{0x07}, asdu.StepPosition{0x07, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asdu.ParseStepPosition(tt.args.value); got != tt.want {
				t.Errorf("NewStepPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepPosition_Value(t *testing.T) {
	for _, HasTransient := range []bool{false, true} {
		for value := -64; value <= 63; value++ {
			got := asdu.ParseStepPosition(asdu.StepPosition{value, HasTransient}.Value())
			if got.Val != value || got.HasTransient != HasTransient {
				t.Errorf("ParseStepPosition(StepPosition(%d, %t).Value()) = StepPosition(%d, %t)", value, HasTransient, got.Val, got.HasTransient)
			}
		}
	}
}

// TestNormal tests the full value range.
func TestNormal(t *testing.T) {
	v := asdu.Normalize(-1 << 15)
	last := v.Float64()
	if last != -1 {
		t.Errorf("%#04x: got %f, want -1", uint16(v), last)
	}

	for v != 1<<15-1 {
		v++
		got := v.Float64()
		if got <= last || got >= 1 {
			t.Errorf("%#04x: got %f (%#04x was %f)", uint16(v), got, uint16(v-1), last)
		}
		last = got
	}
}

func TestNormalize_Float64(t *testing.T) {
	minim := float64(-1)

	for v := math.MinInt16; v < math.MaxInt16; v++ {
		got := asdu.Normalize(v).Float64()
		if got < minim || got >= 1 {
			t.Errorf("%#04x: got %f (%#04x was %f)", uint16(v), got, uint16(v-1), minim)
		}
		minim = got
	}
}

func TestParseQualifierOfCmd(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want asdu.QualifierOfCommand
	}{
		{"with selects", args{0x84}, asdu.QualifierOfCommand{1, true}},
		{"with executes", args{0x0c}, asdu.QualifierOfCommand{3, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asdu.ParseQualifierOfCommand(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQualifierOfCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQualifierOfSetpointCmd(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want asdu.QualifierOfSetpointCmd
	}{
		{"with selects", args{0x87}, asdu.QualifierOfSetpointCmd{7, true}},
		{"with executes", args{0x07}, asdu.QualifierOfSetpointCmd{7, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asdu.ParseQualifierOfSetpointCmd(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQualifierOfSetpointCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQualifierOfCmd_Value(t *testing.T) {
	type fields struct {
		CmdQ   asdu.QOCQual
		InExec bool
	}
	tests := []struct {
		name   string
		fields fields
		want   byte
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := asdu.QualifierOfCommand{
				Qual:     tt.fields.CmdQ,
				InSelect: tt.fields.InExec,
			}
			if got := this.Value(); got != tt.want {
				t.Errorf("QualifierOfCommand.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQualifierOfSetpointCmd_Value(t *testing.T) {
	type fields struct {
		CmdS   asdu.QOSQual
		InExec bool
	}
	tests := []struct {
		name   string
		fields fields
		want   byte
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := asdu.QualifierOfSetpointCmd{
				Qual:     tt.fields.CmdS,
				InSelect: tt.fields.InExec,
			}
			if got := this.Value(); got != tt.want {
				t.Errorf("QualifierOfSetpointCmd.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQualifierOfParam(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want asdu.QualifierOfParameterMV
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asdu.ParseQualifierOfParamMV(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQualifierOfParamMV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQualifierOfParam_Value(t *testing.T) {
	type fields struct {
		ParamQ        asdu.QPMCategory
		IsChange      bool
		IsInOperation bool
	}
	tests := []struct {
		name   string
		fields fields
		want   byte
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := asdu.QualifierOfParameterMV{
				Category:      tt.fields.ParamQ,
				IsChange:      tt.fields.IsChange,
				IsInOperation: tt.fields.IsInOperation,
			}
			if got := this.Value(); got != tt.want {
				t.Errorf("QualifierOfParameterMV.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
