package cs104_test

import (
	"reflect"
	"testing"

	"gitlab.com/circutor-library/go-iecp5/cs104"
)

func TestIAPCI_String(t *testing.T) {
	tests := []struct {
		name string
		this cs104.IAPCI
		want string
	}{
		{"iFrame", cs104.IAPCI{SendSN: 0x02, RcvSN: 0x02}, "I[sendNO: 2, recvNO: 2]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("APCI.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestSAPCI_String(t *testing.T) {
	tests := []struct {
		name string
		this cs104.SAPCI
		want string
	}{
		{"sFrame", cs104.SAPCI{RcvSN: 123}, "S[recvNO: 123]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("APCI.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestUAPCI_String(t *testing.T) {
	tests := []struct {
		name string
		this cs104.UAPCI
		want string
	}{
		{"uFrame", cs104.UAPCI{Function: cs104.UStartDtActive}, "U[function: StartDtActive]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("APCI.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newIFrame(t *testing.T) {
	type args struct {
		asdu   []byte
		sendSN uint16
		RcvSN  uint16
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"asdu out of range",
			args{asdu: make([]byte, 250)},
			nil,
			true,
		},
		{
			"asdu right",
			args{[]byte{0x01, 0x02}, 0x06, 0x07},
			[]byte{cs104.StartFrame, 0x06, 0x0c, 0x00, 0x0e, 0x00, 0x01, 0x02},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cs104.NewIFrame(tt.args.sendSN, tt.args.RcvSN, tt.args.asdu)
			if (err != nil) != tt.wantErr {
				t.Errorf("newIFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newIFrame() = % x, want % x", got, tt.want)
			}
		})
	}
}

func Test_newSFrame(t *testing.T) {
	type args struct {
		RcvSN uint16
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"", args{0x06}, []byte{cs104.StartFrame, 0x04, 0x01, 0x00, 0x0c, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cs104.NewSFrame(tt.args.RcvSN); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSFrame() = % x, want % x", got, tt.want)
			}
		})
	}
}

func Test_newUFrame(t *testing.T) {
	type args struct {
		which byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"", args{cs104.UStopDtActive}, []byte{cs104.StartFrame, 0x04, 0x13, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cs104.NewUFrame(tt.args.which); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newUFrame() = % x, want % x", got, tt.want)
			}
		})
	}
}

func Test_parse(t *testing.T) {
	type args struct {
		apdu []byte
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 []byte
	}{
		{
			"iAPCI",
			args{[]byte{cs104.StartFrame, 0x04, 0x02, 0x00, 0x03, 0x00}},
			cs104.IAPCI{SendSN: 0x01, RcvSN: 0x01},
			[]byte{},
		},
		{
			"sAPCI",
			args{[]byte{cs104.StartFrame, 0x04, 0x01, 0x00, 0x02, 0x00}},
			cs104.SAPCI{RcvSN: 0x01},
			[]byte{},
		},
		{
			"uAPCI",
			args{[]byte{cs104.StartFrame, 0x04, 0x07, 0x00, 0x00, 0x00}},
			cs104.UAPCI{cs104.UStartDtActive},
			[]byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := cs104.Parse(tt.args.apdu)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
