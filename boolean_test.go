package msgpack_test

import (
	"testing"

	"github.com/cjbottaro/msgpack"
	"github.com/stretchr/testify/assert"
)

func TestMarshalBoolean(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    []byte
		expectError bool
	}{
		{
			name:     "Marshal true",
			input:    true,
			expected: []byte{0xc3}, // MessagePack encoding for true
		},
		{
			name:     "Marshal false",
			input:    false,
			expected: []byte{0xc2}, // MessagePack encoding for false
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := msgpack.Marshal(tc.input)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, output)
			}
		})
	}
}

func TestUnmarshalBoolean(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    any
		expectError bool
	}{
		{
			name:     "Unmarshal true",
			input:    []byte{0xc3},
			expected: true,
		},
		{
			name:     "Unmarshal false",
			input:    []byte{0xc2},
			expected: false,
		},
		{
			name:        "Unmarshal invalid data",
			input:       []byte{0xff}, // Invalid MessagePack boolean
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output bool
			err := msgpack.Unmarshal(tc.input, &output)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, output)
			}
		})
	}
}

func TestRoundTripBoolean(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{
			name:     "Round-trip true",
			input:    true,
			expected: true,
		},
		{
			name:     "Round-trip false",
			input:    false,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			marshaled, err := msgpack.Marshal(tc.input)
			assert.NoError(t, err)

			var unmarshaled bool
			err = msgpack.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, unmarshaled)
		})
	}
}
