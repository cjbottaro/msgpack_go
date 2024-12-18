package msgpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	_extRegistryByType = make(map[reflect.Type]extHandler)
	_extRegistryById   = make(map[int8]extHandler)
	_anyType           = reflect.TypeOf((*any)(nil)).Elem()
	_anyMapType        = reflect.TypeOf(map[any]any{})
)

type ExtMarshalFn func(any) ([]byte, error)
type ExtUnmarshalFn func([]byte) (any, error)

type extHandler struct {
	typeId      int8
	marshalFn   ExtMarshalFn
	unmarshalFn ExtUnmarshalFn
}

func Marshal(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	buf := new(bytes.Buffer)

	if err := marshalAny(rv, buf); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("msgpack: Unmarshal(non-pointer %s)", rv.Type().String())
	}

	if rv.IsNil() {
		return fmt.Errorf("msgpack: Unmarshal(nil %s)", rv.Type().String())
	}

	rv = rv.Elem() // value that we will fill
	reader := bytes.NewReader(data)
	return unmarshalAny(rv, reader)
}

func MustMarshal(v any) []byte {
	data, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func MustUnmarshal(data []byte, v any) {
	if err := Unmarshal(data, v); err != nil {
		panic(err)
	}
}

func RegisterExt(v any, typeId int8, marshalFn ExtMarshalFn, unmarshalFn ExtUnmarshalFn) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	_extRegistryByType[t] = extHandler{
		typeId:      typeId,
		marshalFn:   marshalFn,
		unmarshalFn: unmarshalFn,
	}

	_extRegistryById[typeId] = extHandler{
		typeId:      typeId,
		marshalFn:   marshalFn,
		unmarshalFn: unmarshalFn,
	}
}

func MarshalTimeExt(v any) ([]byte, error) {
	t := v.(time.Time)
	seconds := uint64(t.Unix())
	nanoseconds := uint64(t.Nanosecond())

	if (seconds >> 34) == 0 {
		content := (nanoseconds << 34) | seconds

		if (content & 0xFFFFFFFF00000000) == 0 {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(content))
			return buf[:], nil
		}

		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], content)
		return buf[:], nil
	}

	var buf [12]byte
	binary.BigEndian.PutUint32(buf[0:4], uint32(nanoseconds))
	binary.BigEndian.PutUint64(buf[4:], seconds)
	return buf[:], nil
}

func UnmarshalTimeExt(buf []byte) (any, error) {
	switch len(buf) {
	case 4:
		// Decode a 4-byte buffer
		content := uint64(binary.BigEndian.Uint32(buf))
		nanoseconds := content >> 34
		seconds := content & 0x3FFFFFFFF // Mask the lower 34 bits
		return time.Unix(int64(seconds), int64(nanoseconds)).UTC(), nil

	case 8:
		// Decode an 8-byte buffer
		content := binary.BigEndian.Uint64(buf)
		nanoseconds := content >> 34
		seconds := content & 0x3FFFFFFFF // Mask the lower 34 bits
		return time.Unix(int64(seconds), int64(nanoseconds)).UTC(), nil

	case 12:
		// Decode a 12-byte buffer
		nanoseconds := binary.BigEndian.Uint32(buf[0:4])
		seconds := binary.BigEndian.Uint64(buf[4:])
		return time.Unix(int64(seconds), int64(nanoseconds)).UTC(), nil

	default:
		return nil, errors.New("msgpack: time ext: invalid size")
	}
}

func structFieldName(f reflect.StructField) (name string) {
	if f.PkgPath != "" {
		return ""
	}

	name = f.Tag.Get("msgpack")

	if name == "" {
		f.Tag.Get("json")
	}

	if name == "" {
		name = f.Name
	}

	if name == "-" {
		return ""
	}

	return name
}
