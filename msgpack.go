package msgpack

import (
	"bytes"
	"fmt"
	"reflect"
)

var (
	_extRegistryByType = make(map[reflect.Type]extHandler)
	_extRegistryById   = make(map[byte]extHandler)
)

type ExtMarshalFn func(reflect.Value) ([]byte, error)
type ExtUnmarshalFn func([]byte, any) error

type extHandler struct {
	typeId      byte
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

func RegisterExt(v any, typeId byte, marshalFn ExtMarshalFn, unmarshalFn ExtUnmarshalFn) {
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
