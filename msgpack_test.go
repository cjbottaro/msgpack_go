package msgpack_test

import (
	"bytes"
	"testing"

	"github.com/cjbottaro/msgpack"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalInt(t *testing.T) {
	values := []int64{-9223372036854775808, -2147483648, -32768, -128, -32, -1, 0, 1, 127, 128, 32767, 2147483647, 9223372036854775807}

	for _, v := range values {
		var out int64

		data, err := msgpack.Marshal(v)
		require.NoError(t, err, "marshal int64 %d", v)

		err = msgpack.Unmarshal(data, &out)
		require.NoError(t, err, "unmarshal int64 %d", v)
		require.Equal(t, v, out, "round-trip int64 %d", v)
	}
}

func TestMarshalUnmarshalUint(t *testing.T) {
	values := []uint64{0, 1, 127, 128, 255, 256, 65535, 65536, 4294967295, 4294967296, 18446744073709551615}

	for _, v := range values {
		var out uint64

		data, err := msgpack.Marshal(v)
		require.NoError(t, err, "marshal uint64 %d", v)

		err = msgpack.Unmarshal(data, &out)
		require.NoError(t, err, "unmarshal uint64 %d", v)
		require.Equal(t, v, out, "round-trip uint64 %d", v)
	}
}

func TestUnmarshalNonPointer(t *testing.T) {
	var data []byte
	data = append(data, 0)              // positive fixint
	err := msgpack.Unmarshal(data, 123) // non-pointer
	require.Error(t, err)
}

func TestUnmarshalNilPointer(t *testing.T) {
	var data []byte
	data = append(data, 0) // positive fixint
	var p *int
	err := msgpack.Unmarshal(data, p)
	require.Error(t, err)
}

func TestUnmarshalOverflow(t *testing.T) {
	data, err := msgpack.Marshal(int64(2147483647)) // fits in int32
	require.NoError(t, err)

	var smallInt int8
	err = msgpack.Unmarshal(data, &smallInt)
	require.Error(t, err, "should overflow int8")

	data, err = msgpack.Marshal(int64(-1)) // negative
	require.NoError(t, err)
	var u uint8
	err = msgpack.Unmarshal(data, &u)
	require.Error(t, err, "should not store negative into unsigned")
}

func TestUnmarshalUnsupportedType(t *testing.T) {
	data := []byte{0xa0} // fixstr empty (not supported in current implementation)
	var x int
	err := msgpack.Unmarshal(data, &x)
	require.Error(t, err)
}

func TestRoundTripBytes(t *testing.T) {
	// round-trip a large value and make sure the binary is correct
	v := uint64(18446744073709551615)
	data, err := msgpack.Marshal(v)
	require.NoError(t, err)

	var out uint64
	err = msgpack.Unmarshal(data, &out)
	require.NoError(t, err)
	require.Equal(t, v, out)

	// Also verify we used the right encoding for max uint64 (0xcf)
	require.Equal(t, byte(0xcf), data[0])
	require.Equal(t, 9, len(data))
}

func TestMarshalUnmarshalZeroValue(t *testing.T) {
	var i int
	data, err := msgpack.Marshal(i)
	require.NoError(t, err)
	var out int
	err = msgpack.Unmarshal(data, &out)
	require.NoError(t, err)
	require.Equal(t, i, out)
}

func TestMarshalUnmarshalUnsignedWithinSignedRange(t *testing.T) {
	var i uint16 = 32
	data, err := msgpack.Marshal(i)
	require.NoError(t, err)
	var out int8
	err = msgpack.Unmarshal(data, &out)
	require.NoError(t, err)
	require.Equal(t, int8(i), out)
}

func TestMarshalUnmarshalSignedWithinUnsignedRange(t *testing.T) {
	var i int16 = 255
	data, err := msgpack.Marshal(i)
	require.NoError(t, err)
	var out uint16
	err = msgpack.Unmarshal(data, &out)
	require.NoError(t, err)
	require.Equal(t, uint16(i), out)
}

func TestUnmarshalIntoWrongType(t *testing.T) {
	data, err := msgpack.Marshal(int64(-1))
	require.NoError(t, err)

	var u uint64
	err = msgpack.Unmarshal(data, &u)
	require.Error(t, err, "cannot store negative in unsigned")
}

func TestMarshalUnmarshalByteSequence(t *testing.T) {
	// This ensures that we can round-trip data through the buffer
	v := int64(42)
	data, err := msgpack.Marshal(v)
	require.NoError(t, err)

	buf := bytes.NewBuffer(data)
	var out int64
	err = msgpack.Unmarshal(buf.Bytes(), &out)
	require.NoError(t, err)
	require.Equal(t, v, out)
}

func TestUnmarshalNestedStruct(t *testing.T) {
	type NestedStruct struct {
		Value string
	}

	type TestStruct struct {
		PtrField *NestedStruct
	}

	data, err := msgpack.Marshal(map[string]any{
		"PtrField": map[string]any{
			"Value": "Hello",
		},
	})
	if err != nil {
		panic(err)
	}

	var result TestStruct
	err = msgpack.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate result
	if result.PtrField == nil {
		t.Fatalf("PtrField was not allocated")
	}

	if result.PtrField.Value != "Hello" {
		t.Fatalf("Unexpected value in PtrField.Value: %v", result.PtrField.Value)
	}
}
