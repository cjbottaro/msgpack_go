package msgpack_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	msgpack "github.com/cjbottaro/msgpack_go"
)

type Atom string
type Date time.Time

func init() {
	msgpack.RegisterExt((*time.Time)(nil), -0x01, msgpack.MarshalTimeExt, msgpack.UnmarshalTimeExt)

	msgpack.RegisterExt((*Atom)(nil), 0x01,
		func(v any) ([]byte, error) {
			return []byte(v.(Atom)), nil
		},
		func(buf []byte) (any, error) {
			return Atom(buf), nil
		},
	)

	msgpack.RegisterExt((*Date)(nil), 0x02,
		func(v any) ([]byte, error) {
			t := time.Time(v.(Date))
			year, month, day := t.Date()

			if year < -16384 || year > 16383 {
				return nil, fmt.Errorf("year of out of range")
			}

			val := (year << 9) | (int(month) << 5) | day

			var enc [3]byte
			enc[0] = byte((val >> 16) & 0xFF)
			enc[1] = byte((val >> 8) & 0xFF)
			enc[2] = byte(val & 0xFF)

			return enc[:], nil
		},
		func(buf []byte) (any, error) {
			if len(buf) != 3 {
				return nil, fmt.Errorf("invalid length for date ext data")
			}
			val := (int(buf[0]) << 16) | (int(buf[1]) << 8) | int(buf[2])

			year := ((val >> 9) & 0x7FFF)
			month := (val >> 5) & 0xF
			day := val & 0x1F

			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			return Date(t), nil
		},
	)
}

func TestTime(test *testing.T) {
	// ~U[2024-11-25 02:19:12.033203Z] packed by Elixir
	data := []byte("\xD7\xFF\a\xEA\x8C\xE0gCޠ")
	expected, err := time.Parse("2006-01-02T15:04:05.000000Z", "2024-11-25T02:19:12.033203Z")
	if err != nil {
		panic(err)
	}

	d, err := msgpack.Marshal(expected)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(data, d) {
		fmt.Printf("expected: %v\n", data)
		fmt.Printf("  actual: %v\n", d)
		test.FailNow()
	}

	var v any
	if err := msgpack.Unmarshal(data, &v); err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(expected, v) {
		fmt.Printf("expected: %v\n", expected)
		fmt.Printf("  actual: %v\n", v)
		test.FailNow()
	}

	var t time.Time
	if err := msgpack.Unmarshal(data, &t); err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(expected, t) {
		fmt.Printf("expected: %v\n", expected)
		fmt.Printf("  actual: %v\n", t)
		test.FailNow()
	}
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

func TestDate(t *testing.T) {
	// ~D[2024-12-01] packed by Elixir
	data := []byte("\xC7\x03\x02\x0Fс")
	expected, err := time.Parse("2006-01-02", "2024-12-01")
	if err != nil {
		panic(err)
	}

	{
		var v any
		err = msgpack.Unmarshal(data, &v)
		if err != nil {
			panic(err)
		}

		if v.(Date) != Date(expected) {
			fmt.Printf("expected: %v\n", expected)
			fmt.Printf("  actual: %v\n", time.Time(v.(Date)))
			t.FailNow()
		}
	}

	{
		var d Date
		err = msgpack.Unmarshal(data, &d)
		if err != nil {
			panic(err)
		}

		if d != Date(expected) {
			fmt.Printf("expected: %v\n", expected)
			fmt.Printf("  actual: %v\n", time.Time(d))
			t.FailNow()
		}
	}

	{
		var p *Date
		err = msgpack.Unmarshal(data, &p)
		if err != nil {
			panic(err)
		}

		if *p != Date(expected) {
			fmt.Printf("expected: %v\n", expected)
			fmt.Printf("  actual: %v\n", time.Time(*p))
			t.FailNow()
		}
	}
}

func TestComplexMap(test *testing.T) {
	data1 := []byte("\x85\x01\xA3one\xD6\x01list\x93\xA3one\x02\xCB@\t\x1E\xB8Q\xEB\x85\x1F\xC7\x06\x01nested\x82\xD6\x01date\xC7\x03\x02\x0Fс\xD6\x01time\xD7\xFF\a\xEA\x8C\xE0gCޠ\xC7\x03\x01two\x02\xA3pie\xCB@\t\x1E\xB8Q\xEB\x85\x1F")

	timeVal, err := time.Parse("2006-01-02T15:04:05.000000Z", "2024-11-25T02:19:12.033203Z")
	if err != nil {
		panic(err)
	}

	dateVal, err := time.Parse("2006-01-02", "2024-12-01")
	if err != nil {
		panic(err)
	}

	expected := map[any]any{
		1:            "one",
		Atom("two"):  2,
		"pie":        3.14,
		Atom("list"): []any{"one", 2, 3.14},
		Atom("nested"): map[any]any{
			Atom("time"): timeVal,
			Atom("date"): Date(dateVal),
		},
	}

	data2, err := msgpack.Marshal(&expected)
	if err != nil {
		panic(err)
	}

	// We can't check to see if data1 and data2 are the same because map key
	// sorting difference between Elixir and Go.

	var actual1 any
	if err := msgpack.Unmarshal(data1, &actual1); err != nil {
		panic(err)
	}

	var actual2 any
	if err := msgpack.Unmarshal(data2, &actual2); err != nil {
		panic(err)
	}

	expectedStr := fmt.Sprintf("%+v", expected)

	actualStr := fmt.Sprintf("%+v", actual1)
	if expectedStr != actualStr {
		fmt.Printf("expected: %s\n", expectedStr)
		fmt.Printf("  actual: %s\n", actualStr)
		test.FailNow()
	}

	actualStr = fmt.Sprintf("%+v", actual2)
	if expectedStr != actualStr {
		fmt.Printf("expected: %s\n", expectedStr)
		fmt.Printf("  actual: %s\n", actualStr)
		test.FailNow()
	}
}
