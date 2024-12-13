package msgpack_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cjbottaro/msgpack"
)

type Atom string
type Date time.Time

func init() {
	msgpack.RegisterExt((*Atom)(nil), 0x01,
		func(v any) ([]byte, error) {
			return []byte(v.(Atom)), nil
		},
		func(buf []byte) (any, error) {
			return Atom(buf), nil
		},
	)

	// msgpack.RegisterExt((*Date)(nil), 0x02,
	// 	func(rv reflect.Value) ([]byte, error) {
	// 		return []byte(rv.Interface().(Atom)), nil
	// 	},
	// 	func(buf []byte) (any, error) {
	// 		return Atom(buf), nil
	// 	},
	// )
}

func TestAtom(t *testing.T) {
	// :hello serialized by Elixir
	data := []byte("\xC7\x05\x01hello")
	item := Atom("hello")

	d, err := msgpack.Marshal(item)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(data, d) {
		fmt.Printf("expected: %v\n", string(data))
		fmt.Printf("  actual: %v\n", string(d))
		t.FailNow()
	}

	var v any
	err = msgpack.Unmarshal(data, &v)
	if err != nil {
		panic(err)
	}

	if item != v.(Atom) {
		fmt.Printf("expected: %v", data)
		fmt.Printf("  actual: %v", v)
		t.FailNow()
	}

	var a Atom
	err = msgpack.Unmarshal(data, &a)
	if err != nil {
		panic(err)
	}

	if item != a {
		fmt.Printf("expected: %v", data)
		fmt.Printf("  actual: %v", a)
		t.FailNow()
	}

	var x int
	err = msgpack.Unmarshal(data, &x)
	if err.Error() != "msgpack: cannot unmarshal msgpack_test.Atom into Go value of type int" {
		t.FailNow()
	}
}

func TestAtomNested(t *testing.T) {
	mi := map[any]any{
		"foo": Atom("bar"),
	}

	data, err := msgpack.Marshal(mi)
	if err != nil {
		panic(err)
	}

	var mo map[any]any
	err = msgpack.Unmarshal(data, &mo)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(mi, mo) {
		fmt.Printf("expected: %+v\n", mi)
		fmt.Printf("  actual: %+v\n", mo)
		t.FailNow()
	}

	{
		type AtomStruct struct {
			Foo Atom `msgpack:"foo"`
		}

		var s AtomStruct
		err = msgpack.Unmarshal(data, &s)
		if err != nil {
			panic(err)
		}

		expected := AtomStruct{
			Foo: Atom("bar"),
		}

		if !reflect.DeepEqual(expected, s) {
			fmt.Printf("expected: %+v\n", expected)
			fmt.Printf("  actual: %+v\n", s)
			t.FailNow()
		}
	}

	{
		type AtomStructPtr struct {
			Foo *Atom `msgpack:"foo"`
		}

		var s AtomStructPtr
		err = msgpack.Unmarshal(data, &s)
		if err != nil {
			panic(err)
		}

		foo := Atom("bar")
		expected := AtomStructPtr{
			Foo: &foo,
		}

		if !reflect.DeepEqual(expected, s) {
			fmt.Printf("expected: %+v\n", expected)
			fmt.Printf("  actual: %+v\n", s)
			t.FailNow()
		}
	}
}
