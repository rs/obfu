package obfu

import (
	"math/rand"
	"net"
	"testing"
)

func TestObfuscateIO(t *testing.T) {
	src := []byte{}
	for i := 0; i < 16; i++ {
		src = append(src, 'a')
		res, err := Decode(Encode(src))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(res), string(src); got != want {
			t.Errorf("encode != decode: got: %q, want %q", got, want)
		}
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		in   []byte
		seed int64
		want string
	}{
		{net.ParseIP("0.0.0.0").To4(), 0, "pd6c76qiiiiiii"},
		{net.ParseIP("254.123.231.213").To4(), 0, "pd6c76qh7dtskv"},
		{net.ParseIP("254.123.231.213").To4(), 1, "jvsyeiira2zi3b"},
		{net.ParseIP("fe80::aede:48ff:fe00:1122"), 0, "pd6c76qh3iiiiiiiiiic6xtl4hhoiiafv"},
	}
	for i, tt := range tests {
		rand.Seed(tt.seed)
		if got := string(Encode(tt.in)); got != tt.want {
			t.Errorf("#%d: invalid encode result: got %q, want %q", i, got, tt.want)
		}
	}
}

func TestDecode(t *testing.T) {
	_, err := DecodeString("")
	if err == nil {
		t.Fatalf("wanted error, got nil")
	}
	if got, want := err.Error(), "invalid input size"; got != want {
		t.Errorf("invalid error: got %q, want %q", got, want)
	}
	_, err = DecodeString("99999999")
	if err == nil {
		t.Fatalf("wanted error, got nil")
	}
	if got, want := err.Error(), "seed decode error: illegal base32 data at input byte 0"; got != want {
		t.Errorf("invalid error: got %q, want %q", got, want)
	}
	_, err = DecodeString("pd6c76q9")
	if got, want := err.Error(), "payload decode error: illegal base32 data at input byte 0"; got != want {
		t.Errorf("invalid error: got %q, want %q", got, want)
	}
}
