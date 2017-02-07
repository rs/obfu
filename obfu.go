// Package obfu implements a basic domain safe obfuscation algorithm with randomness
// so the same data obfuscated twice won't give the same result.
package obfu

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// standard base32 encoder using lowercase
const stdEncode = "abcdefghijklmnopqrstuvwxyz234567"

var stdEncoding = base32.NewEncoding(stdEncode)

// EncodeString generates an obfuscated version of the input string with randomness
// to make the output unique.
func EncodeString(src string) string {
	return string(Encode([]byte(src)))
}

// Encode generates an obfuscated version of the input byte with randomness to make the
// output unique.
func Encode(src []byte) []byte {
	// Create a random 32 bits seed.
	seed := rand.Int31()
	// Get custom encoder from seed
	enc := newSeedEncoding(int64(seed))
	// Compute encoded len of seed + input.
	seedLen := 8
	encLen := enc.EncodedLen(len(src))
	dst := make([]byte, encLen+seedLen-1) // -1 to skip the seed padding
	// Encode the seed using standard base32 lowercase on the first
	// bytes of the output.
	seedBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seedBytes, uint32(seed))
	stdEncoding.Encode(dst, seedBytes)
	// Encode the source with the custom base32 encoder and append
	// it to the output. (seedLen-1 is to skip the seed's padding)
	enc.Encode(dst[seedLen-1:], src)
	// Remove padding
	dst = bytes.TrimRight(dst, "=")
	return dst
}

func newSeedEncoding(seed int64) *base32.Encoding {
	// Use the seeded rand to compute a custom base32 encoder.
	rnd := rand.New(rand.NewSource(int64(seed)))
	perm := rnd.Perm(len(stdEncode))
	encode := make([]byte, len(stdEncode))
	for i, v := range perm {
		encode[i] = stdEncode[v]
	}
	return base32.NewEncoding(string(encode))
}

// DecodeString revert an obfuscated string
func DecodeString(src string) (string, error) {
	dst, err := Decode([]byte(src))
	return string(dst), err
}

// Decode revert an obfuscated byte array
func Decode(src []byte) ([]byte, error) {
	// Source must be at least the size of the base32 32 bits seed
	if len(src) < 7 {
		return nil, errors.New("invalid input size")
	}
	// Extract the seed and append the missing padding
	seedEnc := make([]byte, 8)
	copy(seedEnc, src[:7])
	seedEnc[7] = '='
	seedBytes := make([]byte, 5)
	n, err := stdEncoding.Decode(seedBytes, seedEnc)
	if err != nil {
		return nil, fmt.Errorf("seed decode error: %v", err)
	}
	seed := binary.BigEndian.Uint32(seedBytes[:n])
	enc := newSeedEncoding(int64(seed))
	src = src[7:]
	if pad := 8 - len(src)%8; pad != 8 {
		for i := 0; i < pad; i++ {
			src = append(src, '=')
		}
	}
	dst := make([]byte, enc.DecodedLen(len(src)))
	n, err = enc.Decode(dst, src)
	if err != nil {
		return nil, fmt.Errorf("payload decode error: %v", err)
	}
	return dst[:n], nil
}
