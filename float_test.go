package msgpack_test

import (
	"math"
	"testing"

	"github.com/cjbottaro/msgpack"
	"github.com/stretchr/testify/assert"
)

func TestMarshalFloat(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    []byte
		expectError bool
	}{
		{
			name:     "Marshal float32",
			input:    float32(3.14),
			expected: nil, // Exact binary encoding depends on MessagePack implementation; checked in round-trip tests
		},
		{
			name:     "Marshal float64",
			input:    float64(3.141592653589793),
			expected: nil, // Checked in round-trip tests
		},
		{
			name:     "Marshal negative float64",
			input:    float64(-42.42),
			expected: nil, // Checked in round-trip tests
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := msgpack.Marshal(tc.input)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
			}
		})
	}
}

func TestUnmarshalFloat(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    any
		expectError bool
	}{
		{
			name:     "Unmarshal float32",
			input:    []byte{0xca, 0x40, 0x48, 0xf5, 0xc3}, // Example MessagePack encoding for float32(3.14)
			expected: float32(3.14),
		},
		{
			name:     "Unmarshal float64",
			input:    []byte{0xcb, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}, // Example for float64(3.141592653589793)
			expected: float64(3.141592653589793),
		},
		{
			name:        "Unmarshal invalid float data",
			input:       []byte{0xff}, // Invalid MessagePack encoding
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output any
			err := msgpack.Unmarshal(tc.input, &output)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InEpsilon(t, tc.expected, output, 0.000001) // Allow slight precision differences
			}
		})
	}
}

func TestRoundTripFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "Round-trip float32",
			input:    float32(3.14),
			expected: float32(3.14),
		},
		{
			name:     "Round-trip float64",
			input:    float64(3.141592653589793),
			expected: float64(3.141592653589793),
		},
		{
			name:     "Round-trip float64 with large value",
			input:    float64(math.MaxFloat64),
			expected: float64(math.MaxFloat64),
		},
		{
			name:     "Round-trip float64 with small value",
			input:    float64(math.SmallestNonzeroFloat64),
			expected: float64(math.SmallestNonzeroFloat64),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			marshaled, err := msgpack.Marshal(tc.input)
			assert.NoError(t, err)

			var unmarshaled any
			err = msgpack.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.InEpsilon(t, tc.expected, unmarshaled, 0.000001) // Allow slight precision differences
		})
	}
}
