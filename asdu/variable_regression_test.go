package asdu

import (
	"errors"
	"testing"
)

func TestSetVariableNumberRejectsOutOfRange(t *testing.T) {
	a := NewEmptyASDU(ParamsWide)
	for _, n := range []int{0, 1, 127} {
		if err := a.SetVariableNumber(n); err != nil || int(a.Variable.Number) != n {
			t.Fatalf("valid count %d: number=%d, err=%v", n, a.Variable.Number, err)
		}
	}
	for _, n := range []int{-256, -1, 128, 256} {
		if err := a.SetVariableNumber(n); !errors.Is(err, ErrInfoObjIndexFit) {
			t.Errorf("count %d: got %v", n, err)
		}
		if a.Variable.Number != 127 {
			t.Fatal("invalid count changed the previous value")
		}
	}
}
