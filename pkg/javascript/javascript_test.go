package javascript_test

import (
	"testing"
)

func TestArithmetic(t *testing.T) {
	if 2+2 != 4 {
		t.Fatalf("Arithmetic cannot be counted on")
	}
}
