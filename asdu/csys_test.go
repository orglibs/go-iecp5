package asdu_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/orglibs/go-iecp5/asdu"
)

func TestInterrogationCmd(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		qoi asdu.QualifierOfInterrogation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cause not Activation and Deactivation",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				asdu.QOIGroup1},
			true,
		},
		{
			"C_IC_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_IC_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 21}),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.QOIGroup1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.InterrogationCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qoi); (err != nil) != tt.wantErr {
				t.Errorf("InterrogationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCounterInterrogationCmd(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		qcc asdu.QualifierCountCall
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_CI_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_CI_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x01}),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.QualifierCountCall{asdu.QCCGroup1, asdu.QCCFrzRead}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.CounterInterrogationCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qcc); (err != nil) != tt.wantErr {
				t.Errorf("CounterInterrogationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadCmd(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		ioa asdu.InfoObjAddr
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_RD_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_RD_NA_1), 0x01, 0x05, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56}),
				asdu.CauseOfTransmission{Cause: asdu.Request},
				0x1234,
				0x567890},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.ReadCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.ioa); (err != nil) != tt.wantErr {
				t.Errorf("ReadCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClockSynchronizationCmd(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		t   time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_CS_NA_1",
			args{
				newConn(t, append([]byte{byte(asdu.C_CS_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00}, tm0CP56Time2aBytes...)),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				tm0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.ClockSynchronizationCmd(tt.args.c, tt.args.coa, tt.args.ca, true, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("ClockSynchronizationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestCommand(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_TS_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_TS_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0xaa, 0x55}),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.TestCommand(tt.args.c, tt.args.coa, tt.args.ca); (err != nil) != tt.wantErr {
				t.Errorf("TestCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResetProcessCmd(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		qrp asdu.QualifierOfResetProcessCmd
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_RP_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_RP_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x01}),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				asdu.QPRGeneralRest},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.ResetProcessCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qrp); (err != nil) != tt.wantErr {
				t.Errorf("ResetProcessCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelayAcquireCommand(t *testing.T) {
	type args struct {
		c    asdu.Connect
		coa  asdu.CauseOfTransmission
		ca   asdu.CommonAddr
		msec uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cause not act and spont",
			args{
				newConn(t, nil),
				asdu.CauseOfTransmission{Cause: asdu.Unused},
				0x1234,
				10000},
			true,
		},
		{
			"C_CD_NA_1",
			args{
				newConn(t, []byte{byte(asdu.C_CD_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x10, 0x27}),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				10000},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.DelayAcquireCommand(tt.args.c, tt.args.coa, tt.args.ca, tt.args.msec); (err != nil) != tt.wantErr {
				t.Errorf("DelayAcquireCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestCommandCP56Time2a(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		t   time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_TS_TA_1",
			args{
				newConn(t, append([]byte{byte(asdu.C_TS_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0xaa, 0x55}, tm0CP56Time2aBytes...)),
				asdu.CauseOfTransmission{Cause: asdu.Activation},
				0x1234,
				tm0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.TestCommandCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("TestCommandCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASDU_GetInterrogationCmd(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  asdu.QualifierOfInterrogation
	}{
		{
			"C_IC_NA_1",
			fields{asdu.ParamsWide, []byte{0x00, 0x00, 0x00, 21}},
			0,
			asdu.QOIGroup1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetInterrogationCmd()
			if got != tt.want {
				t.Errorf("ASDU.GetInterrogationCmd() QOI = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ASDU.GetInterrogationCmd() InfoObjAddr = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetCounterInterrogationCmd(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  asdu.QualifierCountCall
	}{
		{
			"C_CI_NA_1",
			fields{asdu.ParamsWide, []byte{0x00, 0x00, 0x00, 0x01}},
			0,
			asdu.QualifierCountCall{asdu.QCCGroup1, asdu.QCCFrzRead},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetCounterInterrogationCmd()
			if got != tt.want {
				t.Errorf("ASDU.GetQuantityInterrogationCmd() InfoObjAddr = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ASDU.GetQuantityInterrogationCmd() QCC = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetReadCmd(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
	}{
		{
			"C_RD_NA_1",
			fields{asdu.ParamsWide, []byte{0x90, 0x78, 0x56}},
			0x567890,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got := this.GetReadCmd()
			if got != tt.want {
				t.Errorf("ASDU.GetReadCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASDU_GetClockSynchronizationCmd(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  time.Time
	}{
		{
			"C_CS_NA_1",
			fields{asdu.ParamsWide, append([]byte{0x00, 0x00, 0x00}, tm0CP56Time2aBytes...)},
			0,
			tm0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetClockSynchronizationCmd()
			if got != tt.want {
				t.Errorf("ASDU.GetClockSynchronizationCmd() InfoObjAddr = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ASDU.GetClockSynchronizationCmd() time = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetTestCommand(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  bool
	}{
		{
			"C_CS_NA_1",
			fields{asdu.ParamsWide, []byte{0x00, 0x00, 0x00, 0xaa, 0x55}},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetTestCommand()
			if got != tt.want {
				t.Errorf("ASDU.GetTestCommand() InfoObjAddr = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ASDU.GetTestCommand() bool = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetResetProcessCmd(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  asdu.QualifierOfResetProcessCmd
	}{
		{
			"C_RP_NA_1",
			fields{asdu.ParamsWide, []byte{0x00, 0x00, 0x00, 0x01}},
			0,
			asdu.QPRGeneralRest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetResetProcessCmd()
			if got != tt.want {
				t.Errorf("ASDU.GetResetProcessCmd() InfoObjAddr = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ASDU.GetResetProcessCmd() QOP = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetDelayAcquireCommand(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  uint16
	}{
		{
			"C_CD_NA_1",
			fields{asdu.ParamsWide, []byte{0x00, 0x00, 0x00, 0x10, 0x27}},
			0,
			10000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetDelayAcquireCommand()
			if got != tt.want {
				t.Errorf("ASDU.GetDelayAcquireCommand() InfoObjAddr = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ASDU.GetDelayAcquireCommand() msec = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestASDU_GetTestCommandCP56Time2a(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  bool
		want2  time.Time
	}{
		{
			"C_CS_TA_1",
			fields{asdu.ParamsWide, append([]byte{0x00, 0x00, 0x00, 0xaa, 0x55}, tm0CP56Time2aBytes...)},
			0,
			true,
			tm0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1, got2 := sf.GetTestCommandCP56Time2a()
			if got != tt.want {
				t.Errorf("GetTestCommandCP56Time2a() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetTestCommandCP56Time2a() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetTestCommandCP56Time2a() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
