package utils

import "testing"

func TestAddressEquality(t *testing.T) {
	a1 := NewLegionAddress("localhost", 1234)
	a2 := NewLegionAddress("localhost", 1234)

	if a1 != a2 {
		t.Error("addresses not equal")
	}
}
