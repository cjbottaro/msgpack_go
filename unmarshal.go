package msgpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

func unmarshalAny(rv reflect.Value, reader *bytes.Reader) error {
	b, err := reader.ReadByte()
	if err != nil {
		return err
	}

	// Handle pointers
	if rv.Kind() == reflect.Pointer {
		if b == 0xc0 {
			rv.SetZero()
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem())) // Allocate memory for the pointer
		}
		rv = rv.Elem() // Dereference to actual value
	}

	switch {
	case b == 0xc2 || b == 0xc3:
		return unmarshalBool(b, rv, reader)
	case (b & 0b11100000) == 0b11100000:
		return unmarshalIntFixNeg(b, rv, reader)
	case (b & 0b10000000) == 0b00000000:
		return unmarshalIntFixPos(b, rv, reader)
	case b == 0xd0:
		return unmarshalInt8(b, rv, reader)
	case b == 0xd1:
		return unmarshalInt16(b, rv, reader)
	case b == 0xd2:
		return unmarshalInt32(b, rv, reader)
	case b == 0xd3:
		return unmarshalInt64(b, rv, reader)
	case b == 0xcc:
		return unmarshalUint8(b, rv, reader)
	case b == 0xcd:
		return unmarshalUint16(b, rv, reader)
	case b == 0xce:
		return unmarshalUint32(b, rv, reader)
	case b == 0xcf:
		return unmarshalUint64(b, rv, reader)
	case b == 0xca:
		return unmarshalFloat32(b, rv, reader)
	case b == 0xcb:
		return unmarshalFloat64(b, rv, reader)
	case (b & 0b10100000) == 0b10100000:
		return unmarshalStrFix(b, rv, reader)
	case b == 0xd9:
		return unmarshalStr8(b, rv, reader)
	case b == 0xda:
		return unmarshalStr16(b, rv, reader)
	case b == 0xdb:
		return unmarshalStr32(b, rv, reader)
	case b == 0xc4:
		return unmarshalBin8(b, rv, reader)
	case b == 0xc5:
		return unmarshalBin16(b, rv, reader)
	case b == 0xc6:
		return unmarshalBin32(b, rv, reader)
	case (b & 0b10010000) == 0b10010000:
		return unmarshalArrayFix(b, rv, reader)
	case b == 0xdc:
		return unmarshalArray16(b, rv, reader)
	case b == 0xdd:
		return unmarshalArray32(b, rv, reader)
	case (b & 0b10000000) == 0b10000000:
		return unmarshalMapFix(b, rv, reader)
	case b == 0xde:
		return unmarshalMap16(b, rv, reader)
	case b == 0xdf:
		return unmarshalMap32(b, rv, reader)
	default:
		return errors.New("unknown msgpack type")
	}
}

func unmarshalBool(b byte, rv reflect.Value, _ *bytes.Reader) error {
	if rv.Kind() != reflect.Bool {
		return fmt.Errorf("msgpack: cannot unmarshal boolean into Go value of type %v", rv.Type())
	}
	rv.SetBool(b == 0xc3)
	return nil
}

func unmarshalIntFixNeg(b byte, rv reflect.Value, _ *bytes.Reader) error {
	return unmarshalInt(int64(int8(b)), rv)
}

func unmarshalIntFixPos(b byte, rv reflect.Value, _ *bytes.Reader) error {
	return unmarshalInt(int64(b), rv)
}

func unmarshalInt8(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n int8
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalInt(int64(n), rv)
}

func unmarshalInt16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n int16
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalInt(int64(n), rv)
}

func unmarshalInt32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n int32
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalInt(int64(n), rv)
}

func unmarshalInt64(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n int64
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalInt(n, rv)
}

func unmarshalInt(v int64, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Int:
		if v > int64(^uint(0)>>1) || v < -int64(^uint(0)>>1)-1 { // Check range for int
			return fmt.Errorf("msgpack: integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetInt(v)
	case reflect.Int8:
		if v > int64(^uint8(0)>>1) || v < -int64(^uint8(0)>>1)-1 { // Check range for int8
			return fmt.Errorf("msgpack: integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetInt(v)
	case reflect.Int16:
		if v > int64(^uint16(0)>>1) || v < -int64(^uint16(0)>>1)-1 { // Check range for int16
			return fmt.Errorf("msgpack: integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetInt(v)
	case reflect.Int32:
		if v > int64(^uint32(0)>>1) || v < -int64(^uint32(0)>>1)-1 { // Check range for int32
			return fmt.Errorf("msgpack: integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetInt(v)
	case reflect.Int64:
		rv.SetInt(v) // int64 can hold any int64 value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v < 0 { // Unsigned integers cannot hold negative values
			return fmt.Errorf("msgpack: integer %d underflows Go type %v", v, rv.Type())
		}
		if v > math.MaxInt64 { // Check for overflow when converting to unsigned
			return fmt.Errorf("msgpack: integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetUint(uint64(v))
	default:
		return fmt.Errorf("msgpack: cannot unmarshal integer into Go value of type %v", rv.Type())
	}
	return nil
}

func unmarshalUint8(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n uint8
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalUint(uint64(n), rv)
}

func unmarshalUint16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n uint16
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalUint(uint64(n), rv)
}

func unmarshalUint32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n uint32
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalUint(uint64(n), rv)
}

func unmarshalUint64(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var n uint64
	if err := binary.Read(reader, binary.BigEndian, &n); err != nil {
		return err
	}
	return unmarshalUint(n, rv)
}

func unmarshalUint(v uint64, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Uint:
		if v > uint64(^uint(0)) { // Max value of uint
			return fmt.Errorf("msgpack: unsigned integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetUint(v)
	case reflect.Uint8:
		if v > uint64(^uint8(0)) { // Max value of uint8
			return fmt.Errorf("msgpack: unsigned integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetUint(v)
	case reflect.Uint16:
		if v > uint64(^uint16(0)) { // Max value of uint16
			return fmt.Errorf("msgpack: unsigned integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetUint(v)
	case reflect.Uint32:
		if v > uint64(^uint32(0)) { // Max value of uint32
			return fmt.Errorf("msgpack: unsigned integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetUint(v)
	case reflect.Uint64:
		rv.SetUint(v) // uint64 can hold any uint64 value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Check if the unsigned value can fit into the signed integer type
		max := int64(^uint64(0) >> 1) // Max signed int64
		if v > uint64(max) {
			return fmt.Errorf("msgpack: unsigned integer %d overflows Go type %v", v, rv.Type())
		}
		rv.SetInt(int64(v))
	default:
		return fmt.Errorf("msgpack: cannot unmarshal unsigned integer into Go value of type %v", rv.Type())
	}
	return nil
}

func unmarshalFloat32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var v float32

	if err := checkFloat(rv); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return err
	}

	rv.SetFloat(float64(v))

	return nil
}

func unmarshalFloat64(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var v float64

	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return err
	}

	rv.SetFloat(v)

	return nil
}

func checkFloat(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return nil
	default:
		return fmt.Errorf("msgpack: cannot unmarshal float into Go value of type %v", rv.Type())
	}
}

func unmarshalStrFix(b byte, rv reflect.Value, reader *bytes.Reader) error {
	l := uint8(b & 0b00011111)
	if l > 31 {
		return fmt.Errorf("msgpack: invalid FixStr length %d", l)
	}
	var buf [31]byte // Avoid heap allocation.
	return unmarshalStr(uint32(l), buf[:l], rv, reader)
}

func unmarshalStr8(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint8
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read string length: %w", err)
	}
	var buf [255]byte
	return unmarshalStr(uint32(l), buf[:l], rv, reader)
}

func unmarshalStr16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint16
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read string length: %w", err)
	}
	return unmarshalStr(uint32(l), nil, rv, reader)
}

func unmarshalStr32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint32
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read string length: %w", err)
	}
	return unmarshalStr(l, nil, rv, reader)
}

func unmarshalStr(length uint32, buf []byte, rv reflect.Value, reader *bytes.Reader) error {
	if rv.Kind() != reflect.String {
		return fmt.Errorf("msgpack: cannot unmarshal string into Go value of type %v", rv.Type())
	}

	if length == 0 {
		rv.SetString("")
		return nil
	}

	// TODO maybe implement practical size limit to prevent malicious messages.

	// TODO maybe use static allocated buffer for strings up to 1k in size.

	if buf == nil {
		buf = make([]byte, length)
	}

	if _, err := io.ReadFull(reader, buf); err != nil {
		return fmt.Errorf("msgpack: unable to read string data: %w", err)
	}

	rv.SetString(string(buf))
	return nil
}

func unmarshalBin8(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint8
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read binary length: %w", err)
	}
	var buf [255]byte // Stack-allocated buffer for small binaries
	return unmarshalBin(uint32(l), buf[:l], rv, reader)
}

func unmarshalBin16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint16
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read binary length: %w", err)
	}
	return unmarshalBin(uint32(l), nil, rv, reader)
}

func unmarshalBin32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var l uint32
	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return fmt.Errorf("msgpack: unable to read binary length: %w", err)
	}
	return unmarshalBin(l, nil, rv, reader)
}

func unmarshalBin(length uint32, buf []byte, rv reflect.Value, reader *bytes.Reader) error {
	if rv.Kind() != reflect.Slice || rv.Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("msgpack: cannot unmarshal binary into Go value of type %v", rv.Type())
	}

	if length == 0 {
		rv.SetBytes(nil)
		return nil
	}

	// Use a preallocated buffer if provided or create a new one
	if buf == nil {
		buf = make([]byte, length)
	}

	if _, err := io.ReadFull(reader, buf); err != nil {
		return fmt.Errorf("msgpack: unable to read binary data: %w", err)
	}

	rv.SetBytes(buf)
	return nil
}

func unmarshalArrayFix(b byte, rv reflect.Value, reader *bytes.Reader) error {
	length := uint32(b & 0b00001111)
	return unmarshalArray(length, rv, reader)
}

func unmarshalArray16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var length uint16
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("msgpack: unable to read array length: %w", err)
	}
	return unmarshalArray(uint32(length), rv, reader)
}

func unmarshalArray32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("msgpack: unable to read array length: %w", err)
	}
	return unmarshalArray(length, rv, reader)
}

func unmarshalArray(length uint32, rv reflect.Value, reader *bytes.Reader) error {
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("msgpack: cannot unmarshal array into Go value of type %v", rv.Type())
	}

	// Ensure the slice has enough capacity
	if rv.IsNil() || rv.Cap() < int(length) {
		rv.Set(reflect.MakeSlice(rv.Type(), int(length), int(length)))
	} else {
		rv.SetLen(int(length)) // Adjust length without reallocation
	}

	for i := 0; i < int(length); i++ {
		if err := unmarshalAny(rv.Index(i), reader); err != nil {
			return fmt.Errorf("msgpack: unable to unmarshal array element %d: %w", i, err)
		}
	}

	return nil
}

func unmarshalMapFix(b byte, rv reflect.Value, reader *bytes.Reader) error {
	length := uint32(b & 0b00001111)
	return unmarshalMap(length, rv, reader)
}

func unmarshalMap16(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var length uint16
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("msgpack: unable to read map length: %w", err)
	}
	return unmarshalMap(uint32(length), rv, reader)
}

func unmarshalMap32(_ byte, rv reflect.Value, reader *bytes.Reader) error {
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("msgpack: unable to read map length: %w", err)
	}
	return unmarshalMap(length, rv, reader)
}

func unmarshalMap(length uint32, rv reflect.Value, reader *bytes.Reader) error {
	// Handle nil maps or structs
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map:
		return unmarshalIntoMap(length, rv, reader)
	case reflect.Struct:
		return unmarshalIntoStruct(length, rv, reader)
	default:
		return fmt.Errorf("msgpack: cannot unmarshal map into Go value of type %v", rv.Type())
	}
}

func unmarshalIntoMap(length uint32, rv reflect.Value, reader *bytes.Reader) error {
	// Ensure map is initialized
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}

	keyType := rv.Type().Key()
	valueType := rv.Type().Elem()

	for i := uint32(0); i < length; i++ {
		// Unmarshal key
		key := reflect.New(keyType).Elem()
		if err := unmarshalAny(key, reader); err != nil {
			return fmt.Errorf("msgpack: unable to unmarshal map key: %w", err)
		}

		// Unmarshal value
		value := reflect.New(valueType).Elem()
		if err := unmarshalAny(value, reader); err != nil {
			return fmt.Errorf("msgpack: unable to unmarshal map value: %w", err)
		}

		// Set key-value pair in map
		rv.SetMapIndex(key, value)
	}

	return nil
}

func unmarshalIntoStruct(length uint32, rv reflect.Value, reader *bytes.Reader) error {

	// Build the struct field map, excluding any fields that should be skipped via
	// tags, etc.
	rt := rv.Type()
	fieldMap := map[string]int{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if name := structFieldName(field); name != "" {
			fieldMap[name] = i
		}
	}

	for i := uint32(0); i < length; i++ {
		// Unmarshal key
		var key string
		if err := unmarshalAny(reflect.ValueOf(&key).Elem(), reader); err != nil {
			return fmt.Errorf("msgpack: unable to unmarshal struct key: %w", err)
		}

		// Find the corresponding struct field
		fieldIndex, ok := fieldMap[key]
		if !ok {
			var v any
			if err := unmarshalAny(reflect.ValueOf(&v), reader); err != nil {
				return fmt.Errorf("msgpack: unable to skip unknown struct field: %w", err)
			}
			continue
		}

		// Unmarshal value into the field
		field := rv.Field(fieldIndex)
		if !field.CanSet() {
			return fmt.Errorf("msgpack: cannot set field %s in struct %v", key, rv.Type())
		}
		if err := unmarshalAny(field, reader); err != nil {
			return fmt.Errorf("msgpack: unable to unmarshal struct field %s: %w", key, err)
		}
	}

	return nil
}
