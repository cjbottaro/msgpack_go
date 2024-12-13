package msgpack

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

func marshalAny(rv reflect.Value, buf *bytes.Buffer) (err error) {
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return marshalNil(rv, buf)
		}
		rv = rv.Elem()
	case reflect.Interface:
		rv = rv.Elem()
	}

	if handler, found := _extRegistryByType[rv.Type()]; found {
		return marshalExt(rv, handler, buf)
	}

	switch rv.Kind() {
	case reflect.Bool:
		err = marshalBool(rv, buf)
	case reflect.String:
		err = marshalString(rv, buf)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = marshalUint(rv, buf)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = marshalInt(rv, buf)
	case reflect.Float32, reflect.Float64:
		err = marshalFloat(rv, buf)
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			err = marshalBinary(rv, buf)
		} else {
			err = marshalArray(rv, buf)
		}
	case reflect.Map:
		err = marshalMap(rv, buf)
	case reflect.Struct:
		err = marshalStruct(rv, buf)
	}

	return err
}

func marshalNil(_ reflect.Value, buf *bytes.Buffer) error {
	return buf.WriteByte(0xc0)
}

func marshalExt(rv reflect.Value, handler extHandler, buf *bytes.Buffer) error {
	// Use the custom marshal function to get the data
	data, err := handler.marshalFn(rv.Interface())
	if err != nil {
		return err
	}

	length := len(data)

	// Write ext header
	switch {
	case length == 1: // fixext1
		buf.WriteByte(0xd4)
	case length == 2: // fixext2
		buf.WriteByte(0xd5)
	case length == 4: // fixext4
		buf.WriteByte(0xd6)
	case length == 8: // fixext8
		buf.WriteByte(0xd7)
	case length == 16: // fixext16
		buf.WriteByte(0xd8)
	case length <= 255: // ext8
		buf.WriteByte(0xc7)
		binary.Write(buf, binary.BigEndian, uint8(length))
	case length <= 65535: // ext16
		buf.WriteByte(0xc8)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // ext32
		buf.WriteByte(0xc9)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	// Write type identifier
	buf.WriteByte(handler.typeId)

	// Write the serialized data
	_, err = buf.Write(data)
	return err
}

func marshalBool(rv reflect.Value, buf *bytes.Buffer) error {
	if rv.Bool() {
		return buf.WriteByte(0xc3) // true
	}
	return buf.WriteByte(0xc2) // false
}

func marshalUint(rv reflect.Value, buf *bytes.Buffer) error {
	v := rv.Uint()

	switch {
	case v <= 127: // Positive fixint
		buf.WriteByte(uint8(v))
	case v <= 255: // uint8
		buf.WriteByte(0xcc)
		binary.Write(buf, binary.BigEndian, uint8(v))
	case v <= 65535: // uint16
		buf.WriteByte(0xcd)
		binary.Write(buf, binary.BigEndian, uint16(v))
	case v <= 4294967295: // uint32
		buf.WriteByte(0xce)
		binary.Write(buf, binary.BigEndian, uint32(v))
	default: // uint64
		buf.WriteByte(0xcf)
		binary.Write(buf, binary.BigEndian, uint64(v))
	}

	return nil
}

func marshalInt(rv reflect.Value, buf *bytes.Buffer) error {
	v := rv.Int()

	switch {
	case v >= -32 && v <= -1: // Negative fixint
		buf.WriteByte(uint8((v & 0b00011111) | 0b11100000))
	case v >= 0 && v <= 127: // Positive fixint
		buf.WriteByte(uint8(v))
	case v >= -128 && v <= 127: // int8
		buf.WriteByte(0xd0)
		binary.Write(buf, binary.BigEndian, int8(v))
	case v >= -32768 && v <= 32767: // int16
		buf.WriteByte(0xd1)
		binary.Write(buf, binary.BigEndian, int16(v))
	case v >= -2147483648 && v <= 2147483647: // int32
		buf.WriteByte(0xd2)
		binary.Write(buf, binary.BigEndian, int32(v))
	default: // int64
		buf.WriteByte(0xd3)
		binary.Write(buf, binary.BigEndian, int64(v))
	}

	return nil
}

func marshalFloat(rv reflect.Value, buf *bytes.Buffer) error {
	v := rv.Float()

	if rv.Kind() == reflect.Float32 {
		buf.WriteByte(0xca) // float32
		binary.Write(buf, binary.BigEndian, float32(v))
	} else {
		buf.WriteByte(0xcb) // float64
		binary.Write(buf, binary.BigEndian, float64(v))
	}

	return nil
}

func marshalString(rv reflect.Value, buf *bytes.Buffer) error {
	str := rv.String()
	length := len(str)

	switch {
	case length <= 31: // fixstr
		buf.WriteByte(0xa0 | uint8(length))
	case length <= 255: // str8
		buf.WriteByte(0xd9)
		binary.Write(buf, binary.BigEndian, uint8(length))
	case length <= 65535: // str16
		buf.WriteByte(0xda)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // str32
		buf.WriteByte(0xdb)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	_, err := buf.WriteString(str)
	return err
}

func marshalBinary(rv reflect.Value, buf *bytes.Buffer) error {
	data := rv.Bytes()
	length := len(data)

	switch {
	case length <= 255: // bin8
		buf.WriteByte(0xc4)
		binary.Write(buf, binary.BigEndian, uint8(length))
	case length <= 65535: // bin16
		buf.WriteByte(0xc5)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // bin32
		buf.WriteByte(0xc6)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	_, err := buf.Write(data)
	return err
}

func marshalArray(rv reflect.Value, buf *bytes.Buffer) error {
	length := rv.Len()

	// Write the array header
	switch {
	case length <= 15: // fixarray
		buf.WriteByte(0x90 | uint8(length))
	case length <= 65535: // array16
		buf.WriteByte(0xdc)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // array32
		buf.WriteByte(0xdd)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	// Marshal each element
	for i := 0; i < length; i++ {
		elem := rv.Index(i)
		if err := marshalAny(elem, buf); err != nil {
			return err
		}
	}

	return nil
}

func marshalMap(rv reflect.Value, buf *bytes.Buffer) error {
	length := rv.Len()

	// Write the map header
	switch {
	case length <= 15: // fixmap
		buf.WriteByte(0x80 | uint8(length))
	case length <= 65535: // map16
		buf.WriteByte(0xde)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // map32
		buf.WriteByte(0xdf)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	// Marshal each key-value pair
	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// Marshal key
		if err := marshalAny(key, buf); err != nil {
			return err
		}

		// Marshal value
		if err := marshalAny(value, buf); err != nil {
			return err
		}
	}

	return nil
}

func marshalStruct(rv reflect.Value, buf *bytes.Buffer) error {
	rt := rv.Type()
	length := 0

	// Count fields that should be serialized
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if structFieldName(field) != "" {
			length++
		}
	}

	// Write the map header
	switch {
	case length <= 15: // fixmap
		buf.WriteByte(0x80 | uint8(length))
	case length <= 65535: // map16
		buf.WriteByte(0xde)
		binary.Write(buf, binary.BigEndian, uint16(length))
	default: // map32
		buf.WriteByte(0xdf)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}

	// Yes, we're iterating twice and calling structFieldName twice for each
	// field, but we avoid allocations this way.

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldName := structFieldName(field)
		if fieldName == "" {
			continue // Skip fields without a valid name
		}

		// Marshal the field name as the key
		if err := marshalString(reflect.ValueOf(fieldName), buf); err != nil {
			return err
		}

		// Marshal the field value
		fieldValue := rv.Field(i)
		if err := marshalAny(fieldValue, buf); err != nil {
			return err
		}
	}

	return nil
}
