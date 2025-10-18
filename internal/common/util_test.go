package common

import (
	"encoding/hex"
	"testing"
)

// ---------- MakeRandHexString ----------

func TestMakeRandHexString_LengthAndHex(t *testing.T) {
	const n = 16
	s, err := MakeRandHexString(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != n*2 {
		t.Fatalf("expected hex length %d, got %d", n*2, len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		t.Fatalf("string is not valid hex: %v", err)
	}
}

func TestMakeRandHexString_ZeroSize(t *testing.T) {
	s, err := MakeRandHexString(0)
	if err != nil {
		t.Fatalf("unexpected error for size=0: %v", err)
	}
	if s != "" {
		t.Fatalf("expected empty string for size=0, got %q", s)
	}
}

func TestMakeRandHexString_EntropyHint(t *testing.T) {
	const n = 32
	a, err := MakeRandHexString(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := MakeRandHexString(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a == b {
		t.Logf("warning: two MakeRandHexString(%d) results are identical; extremely unlikely", n)
	}
}

// ---------- WipeByteArray ----------

func TestWipeByteArray_ZerosBuffer(t *testing.T) {
	buf := []byte{1, 2, 3, 4, 5}
	WipeByteArray(buf)
	for i, v := range buf {
		if v != 0 {
			t.Fatalf("expected buf[%d]==0, got %d", i, v)
		}
	}
}

func TestWipeByteArray_NilSafe(t *testing.T) {
	WipeByteArray(nil)
}

// ---------- GenerateRandByteArray ----------

func TestGenerateRandByteArray_Basic(t *testing.T) {
	const n = 24
	buf := GenerateRandByteArray(n)
	if buf == nil {
		t.Fatalf("expected non-nil slice")
	}
	if len(buf) != n {
		t.Fatalf("expected length %d, got %d", n, len(buf))
	}
}

func TestGenerateRandByteArray_EntropyHint(t *testing.T) {
	const n = 32
	a := GenerateRandByteArray(n)
	b := GenerateRandByteArray(n)

	if len(a) != n || len(b) != n {
		t.Fatalf("unexpected lengths: %d, %d", len(a), len(b))
	}

	identical := true
	for i := range a {
		if a[i] != b[i] {
			identical = false
			break
		}
	}
	if identical {
		t.Logf("warning: two GenerateRandByteArray(%d) results are identical; extremely unlikely", n)
	}
}
