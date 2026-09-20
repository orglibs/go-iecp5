package cs101

import "testing"

func TestFunctionCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"reset link", FccResetRemoteLink, 0},
		{"reset user process", FccResetUserProcess, 1},
		{"test link", FccBalanceTestLink, 2},
		{"confirmed data", FccUserDataWithConfirmed, 3},
		{"unconfirmed data", FccUserDataWithUnconfirmed, 4},
		{"request bit response", FccUnbalanceWithRequestBitResponse, 8},
		{"request link status", FccLinkStatus, 9},
		{"class 1 data", FccUnbalanceLevel1UserData, 10},
		{"class 2 data", FccUnbalanceLevel2UserData, 11},
		{"acknowledgement", FcsConfirmed, 0},
		{"negative acknowledgement", FcsNConfirmed, 1},
		{"response data", FcsUnbalanceResponse, 8},
		{"no data", FcsUnbalanceNegativeResponse, 9},
		{"link status", FcsStatus, 11},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}
