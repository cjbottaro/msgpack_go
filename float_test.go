package msgpack_test

import (
	"math"
	"testing"

	msgpack "github.com/cjbottaro/msgpack_go"
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

func TestUnmarshalFloatInputOutputCombos(t *testing.T) {
	data32 := msgpack.MustMarshal(float32(3.14))
	data64 := msgpack.MustMarshal(float64(3.14))

	assert.EqualValues(t, 5, len(data32))
	assert.EqualValues(t, 9, len(data64))

	var f32 float32
	var f64 float64

	t.Run("32 bit input, 32 bit output", func(t *testing.T) {
		msgpack.MustUnmarshal(data32, &f32)
		assert.Equal(t, float32(3.14), f32)
	})

	t.Run("64 bit input, 32 bit output", func(t *testing.T) {
		msgpack.MustUnmarshal(data64, &f32)
		assert.Equal(t, float32(3.14), f32)
	})

	t.Run("32 bit input, 64 bit output", func(t *testing.T) {
		msgpack.MustUnmarshal(data32, &f64)
		assert.InEpsilon(t, 3.14, f64, 0.000001)
	})

	t.Run("64 bit input, 64 bit output", func(t *testing.T) {
		msgpack.MustUnmarshal(data64, &f64)
		assert.Equal(t, 3.14, f64)
	})
}

func TestUnmarshalFloatIntoAny(t *testing.T) {
	data32 := msgpack.MustMarshal(float32(3.14))
	data64 := msgpack.MustMarshal(float64(3.14))

	var v any

	t.Run("32 bit input, nil output", func(t *testing.T) {
		msgpack.MustUnmarshal(data32, &v)
		assert.IsType(t, float64(3.14), v)
		assert.InDelta(t, 3.14, v, 0.000001)
	})

	t.Run("64 bit input, nil output", func(t *testing.T) {
		v = nil
		msgpack.MustUnmarshal(data64, &v)
		assert.Equal(t, float64(3.14), v)
	})

	t.Run("32 bit input, 32 bit pointer output", func(t *testing.T) {
		var f float32
		v = &f
		msgpack.MustUnmarshal(data64, &v)
		assert.Equal(t, float32(3.14), *(v.(*float32)))
		assert.Equal(t, float32(3.14), f)
	})
}

func TestUnmarshalFloatIntoWrongType(t *testing.T) {
	var i int64
	var v any

	data := msgpack.MustMarshal(float64(3.14))

	err := msgpack.Unmarshal(data, &i)
	assert.EqualError(t, err, "msgpack: cannot unmarshal float into Go type of int64")

	v = &i
	err = msgpack.Unmarshal(data, &v)
	assert.EqualError(t, err, "msgpack: cannot unmarshal float into Go type of int64")

	v = i
	err = msgpack.Unmarshal(data, &v)
	assert.EqualError(t, err, "msgpack: cannot assign float to unaddressable value")
}
