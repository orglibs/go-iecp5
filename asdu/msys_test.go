package asdu_test

import (
	"reflect"
	"testing"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

func TestEndOfInitialization(t *testing.T) {
	type args struct {
		c   asdu.Connect
		coa asdu.CauseOfTransmission
		ca  asdu.CommonAddr
		ioa asdu.InfoObjAddr
		coi asdu.CauseOfInitial
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"M_EI_NA_1",
			args{
				newConn(t, []byte{byte(asdu.M_EI_NA_1), 0x01, 0x04, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x01}),
				asdu.CauseOfTransmission{Cause: asdu.Initialized},
				0x1234,
				0x567890,
				asdu.CauseOfInitial{asdu.COILocalHandReset, false}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := asdu.EndOfInitialization(tt.args.c, tt.args.coa, tt.args.ca, tt.args.ioa, tt.args.coi); (err != nil) != tt.wantErr {
				t.Errorf("EndOfInitialization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASDU_GetEndOfInitialization(t *testing.T) {
	type fields struct {
		Params  *asdu.Params
		infoObj []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   asdu.InfoObjAddr
		want1  asdu.CauseOfInitial
	}{
		{
			"M_EI_NA_1",
			fields{asdu.ParamsWide, []byte{0x90, 0x78, 0x56, 0x01}},
			0x567890,
			asdu.CauseOfInitial{asdu.COILocalHandReset, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &asdu.ASDU{
				Params:  tt.fields.Params,
				InfoObj: tt.fields.infoObj,
			}
			got, got1 := this.GetEndOfInitialization()
			if got != tt.want {
				t.Errorf("ASDU.GetEndOfInitialization() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ASDU.GetEndOfInitialization() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
