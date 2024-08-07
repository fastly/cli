package optional_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/fastly/cli/thirdparty/optional"
)

func Example_get() {
	i := 1001
	values := []optional.Optional[int]{
		optional.Empty[int](),
		optional.Of(1000),
		optional.OfPtr[int](nil),
		optional.OfPtr(&i),
	}

	for _, v := range values {
		if i, ok := v.Get(); ok {
			fmt.Println(i)
		}
	}

	// Output:
	// 1000
	// 1001
}

func Example_if() {
	i := 1001
	values := []optional.Optional[int]{
		optional.Empty[int](),
		optional.Of(1000),
		optional.OfPtr[int](nil),
		optional.OfPtr(&i),
	}

	for _, v := range values {
		v.If(func(i int) {
			fmt.Println(i)
		})
	}

	// Output:
	// 1000
	// 1001
}

func Example_else() {
	i := 1001
	values := []optional.Optional[int]{
		optional.Empty[int](),
		optional.Of(1000),
		optional.OfPtr[int](nil),
		optional.OfPtr(&i),
	}

	for _, v := range values {
		fmt.Println(v.Else(1))
	}

	// Output:
	// 1
	// 1000
	// 1
	// 1001
}

func Example_elseZero() {
	i := 1001
	values := []optional.Optional[int]{
		optional.Empty[int](),
		optional.Of(1000),
		optional.OfPtr[int](nil),
		optional.OfPtr(&i),
	}

	for _, v := range values {
		fmt.Println(v.ElseZero())
	}

	// Output:
	// 0
	// 1000
	// 0
	// 1001
}

func Example_elseFunc() {
	i := 1001
	values := []optional.Optional[int]{
		optional.Empty[int](),
		optional.Of(1000),
		optional.OfPtr[int](nil),
		optional.OfPtr(&i),
	}

	for _, v := range values {
		fmt.Println(v.ElseFunc(func() int {
			return 2
		}))
	}

	// Output:
	// 2
	// 1000
	// 2
	// 1001
}

func Example_jsonMarshalOmitEmpty() {
	s := struct {
		Bool    optional.Optional[bool]      `json:"bool,omitempty"`
		Byte    optional.Optional[byte]      `json:"byte,omitempty"`
		Float32 optional.Optional[float32]   `json:"float32,omitempty"`
		Float64 optional.Optional[float64]   `json:"float64,omitempty"`
		Int16   optional.Optional[int16]     `json:"int16,omitempty"`
		Int32   optional.Optional[int32]     `json:"int32,omitempty"`
		Int64   optional.Optional[int64]     `json:"int64,omitempty"`
		Int     optional.Optional[int]       `json:"int,omitempty"`
		Rune    optional.Optional[rune]      `json:"rune,omitempty"`
		String  optional.Optional[string]    `json:"string,omitempty"`
		Time    optional.Optional[time.Time] `json:"time,omitempty"`
		Uint16  optional.Optional[uint16]    `json:"uint16,omitempty"`
		Uint32  optional.Optional[uint32]    `json:"uint32,omitempty"`
		Uint64  optional.Optional[uint64]    `json:"uint64,omitempty"`
		Uint    optional.Optional[uint]      `json:"uint,omitempty"`
		Uintptr optional.Optional[uintptr]   `json:"uintptr,omitempty"`
	}{
		Bool:    optional.Empty[bool](),
		Byte:    optional.Empty[byte](),
		Float32: optional.Empty[float32](),
		Float64: optional.Empty[float64](),
		Int16:   optional.Empty[int16](),
		Int32:   optional.Empty[int32](),
		Int64:   optional.Empty[int64](),
		Int:     optional.Empty[int](),
		Rune:    optional.Empty[rune](),
		String:  optional.Empty[string](),
		Time:    optional.Empty[time.Time](),
		Uint16:  optional.Empty[uint16](),
		Uint32:  optional.Empty[uint32](),
		Uint64:  optional.Empty[uint64](),
		Uint:    optional.Empty[uint](),
		Uintptr: optional.Empty[uintptr](),
	}

	output, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// {}
}

func Example_jsonMarshalEmpty() {
	s := struct {
		Bool    optional.Optional[bool]      `json:"bool"`
		Byte    optional.Optional[byte]      `json:"byte"`
		Float32 optional.Optional[float32]   `json:"float32"`
		Float64 optional.Optional[float64]   `json:"float64"`
		Int16   optional.Optional[int16]     `json:"int16"`
		Int32   optional.Optional[int32]     `json:"int32"`
		Int64   optional.Optional[int64]     `json:"int64"`
		Int     optional.Optional[int]       `json:"int"`
		Rune    optional.Optional[rune]      `json:"rune"`
		String  optional.Optional[string]    `json:"string"`
		Time    optional.Optional[time.Time] `json:"time"`
		Uint16  optional.Optional[uint16]    `json:"uint16"`
		Uint32  optional.Optional[uint32]    `json:"uint32"`
		Uint64  optional.Optional[uint64]    `json:"uint64"`
		Uint    optional.Optional[uint]      `json:"uint"`
		Uintptr optional.Optional[uintptr]   `json:"uintptr"`
	}{
		Bool:    optional.Empty[bool](),
		Byte:    optional.Empty[byte](),
		Float32: optional.Empty[float32](),
		Float64: optional.Empty[float64](),
		Int16:   optional.Empty[int16](),
		Int32:   optional.Empty[int32](),
		Int64:   optional.Empty[int64](),
		Int:     optional.Empty[int](),
		Rune:    optional.Empty[rune](),
		String:  optional.Empty[string](),
		Time:    optional.Empty[time.Time](),
		Uint16:  optional.Empty[uint16](),
		Uint32:  optional.Empty[uint32](),
		Uint64:  optional.Empty[uint64](),
		Uint:    optional.Empty[uint](),
		Uintptr: optional.Empty[uintptr](),
	}

	output, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// {
	//   "bool": false,
	//   "byte": 0,
	//   "float32": 0,
	//   "float64": 0,
	//   "int16": 0,
	//   "int32": 0,
	//   "int64": 0,
	//   "int": 0,
	//   "rune": 0,
	//   "string": "",
	//   "time": "0001-01-01T00:00:00Z",
	//   "uint16": 0,
	//   "uint32": 0,
	//   "uint64": 0,
	//   "uint": 0,
	//   "uintptr": 0
	// }
}

func Example_jsonMarshalPresent() {
	s := struct {
		Bool    optional.Optional[bool]      `json:"bool"`
		Byte    optional.Optional[byte]      `json:"byte"`
		Float32 optional.Optional[float32]   `json:"float32"`
		Float64 optional.Optional[float64]   `json:"float64"`
		Int16   optional.Optional[int16]     `json:"int16"`
		Int32   optional.Optional[int32]     `json:"int32"`
		Int64   optional.Optional[int64]     `json:"int64"`
		Int     optional.Optional[int]       `json:"int"`
		Rune    optional.Optional[rune]      `json:"rune"`
		String  optional.Optional[string]    `json:"string"`
		Time    optional.Optional[time.Time] `json:"time"`
		Uint16  optional.Optional[uint16]    `json:"uint16"`
		Uint32  optional.Optional[uint32]    `json:"uint32"`
		Uint64  optional.Optional[uint64]    `json:"uint64"`
		Uint    optional.Optional[uint]      `json:"uint"`
		Uintptr optional.Optional[uintptr]   `json:"uintptr"`
	}{
		Bool:    optional.Of(true),
		Byte:    optional.Of[byte](1),
		Float32: optional.Of[float32](2.1),
		Float64: optional.Of(2.2),
		Int16:   optional.Of[int16](3),
		Int32:   optional.Of[int32](4),
		Int64:   optional.Of[int64](5),
		Int:     optional.Of(6),
		Rune:    optional.Of[rune](7),
		String:  optional.Of("string"),
		Time:    optional.Of(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)),
		Uint16:  optional.Of[uint16](8),
		Uint32:  optional.Of[uint32](9),
		Uint64:  optional.Of[uint64](10),
		Uint:    optional.Of[uint](11),
		Uintptr: optional.Of[uintptr](12),
	}

	output, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// {
	//   "bool": true,
	//   "byte": 1,
	//   "float32": 2.1,
	//   "float64": 2.2,
	//   "int16": 3,
	//   "int32": 4,
	//   "int64": 5,
	//   "int": 6,
	//   "rune": 7,
	//   "string": "string",
	//   "time": "2006-01-02T15:04:05Z",
	//   "uint16": 8,
	//   "uint32": 9,
	//   "uint64": 10,
	//   "uint": 11,
	//   "uintptr": 12
	// }
}

func Example_jsonUnmarshalEmpty() {
	s := struct {
		Bool    optional.Optional[bool]      `json:"bool"`
		Byte    optional.Optional[byte]      `json:"byte"`
		Float32 optional.Optional[float32]   `json:"float32"`
		Float64 optional.Optional[float64]   `json:"float64"`
		Int16   optional.Optional[int16]     `json:"int16"`
		Int32   optional.Optional[int32]     `json:"int32"`
		Int64   optional.Optional[int64]     `json:"int64"`
		Int     optional.Optional[int]       `json:"int"`
		Rune    optional.Optional[rune]      `json:"rune"`
		String  optional.Optional[string]    `json:"string"`
		Time    optional.Optional[time.Time] `json:"time"`
		Uint16  optional.Optional[uint16]    `json:"uint16"`
		Uint32  optional.Optional[uint32]    `json:"uint32"`
		Uint64  optional.Optional[uint64]    `json:"uint64"`
		Uint    optional.Optional[uint]      `json:"uint"`
		Uintptr optional.Optional[uintptr]   `json:"uintptr"`
	}{}

	x := `{}`
	json.Unmarshal([]byte(x), &s)
	fmt.Println("Bool:", s.Bool.IsPresent())
	fmt.Println("Byte:", s.Byte.IsPresent())
	fmt.Println("Float32:", s.Float32.IsPresent())
	fmt.Println("Float64:", s.Float64.IsPresent())
	fmt.Println("Int16:", s.Int16.IsPresent())
	fmt.Println("Int32:", s.Int32.IsPresent())
	fmt.Println("Int64:", s.Int64.IsPresent())
	fmt.Println("Int:", s.Int.IsPresent())
	fmt.Println("Rune:", s.Rune.IsPresent())
	fmt.Println("String:", s.String.IsPresent())
	fmt.Println("Time:", s.Time.IsPresent())
	fmt.Println("Uint16:", s.Uint16.IsPresent())
	fmt.Println("Uint32:", s.Uint32.IsPresent())
	fmt.Println("Uint64:", s.Uint64.IsPresent())
	fmt.Println("Uint64:", s.Uint64.IsPresent())
	fmt.Println("Uint:", s.Uint.IsPresent())
	fmt.Println("Uintptr:", s.Uint.IsPresent())

	// Output:
	// Bool: false
	// Byte: false
	// Float32: false
	// Float64: false
	// Int16: false
	// Int32: false
	// Int64: false
	// Int: false
	// Rune: false
	// String: false
	// Time: false
	// Uint16: false
	// Uint32: false
	// Uint64: false
	// Uint64: false
	// Uint: false
	// Uintptr: false
}

func Example_jsonUnmarshalPresent() {
	s := struct {
		Bool    optional.Optional[bool]      `json:"bool"`
		Byte    optional.Optional[byte]      `json:"byte"`
		Float32 optional.Optional[float32]   `json:"float32"`
		Float64 optional.Optional[float64]   `json:"float64"`
		Int16   optional.Optional[int16]     `json:"int16"`
		Int32   optional.Optional[int32]     `json:"int32"`
		Int64   optional.Optional[int64]     `json:"int64"`
		Int     optional.Optional[int]       `json:"int"`
		Rune    optional.Optional[rune]      `json:"rune"`
		String  optional.Optional[string]    `json:"string"`
		Time    optional.Optional[time.Time] `json:"time"`
		Uint16  optional.Optional[uint16]    `json:"uint16"`
		Uint32  optional.Optional[uint32]    `json:"uint32"`
		Uint64  optional.Optional[uint64]    `json:"uint64"`
		Uint    optional.Optional[uint]      `json:"uint"`
		Uintptr optional.Optional[uintptr]   `json:"uintptr"`
	}{}

	x := `{
   "bool": false,
   "byte": 0,
   "float32": 0,
   "float64": 0,
   "int16": 0,
   "int32": 0,
   "int64": 0,
   "int": 0,
   "rune": 0,
   "string": "string",
   "time": "0001-01-01T00:00:00Z",
   "uint16": 0,
   "uint32": 0,
   "uint64": 0,
   "uint": 0,
   "uintptr": 0
 }`
	json.Unmarshal([]byte(x), &s)
	fmt.Println("Bool:", s.Bool.IsPresent(), s.Bool)
	fmt.Println("Byte:", s.Byte.IsPresent(), s.Byte)
	fmt.Println("Float32:", s.Float32.IsPresent(), s.Float32)
	fmt.Println("Float64:", s.Float64.IsPresent(), s.Float64)
	fmt.Println("Int16:", s.Int16.IsPresent(), s.Int16)
	fmt.Println("Int32:", s.Int32.IsPresent(), s.Int32)
	fmt.Println("Int64:", s.Int64.IsPresent(), s.Int64)
	fmt.Println("Int:", s.Int.IsPresent(), s.Int)
	fmt.Println("Rune:", s.Rune.IsPresent(), s.Rune)
	fmt.Println("String:", s.String.IsPresent(), s.String)
	fmt.Println("Time:", s.Time.IsPresent(), s.Time)
	fmt.Println("Uint16:", s.Uint16.IsPresent(), s.Uint16)
	fmt.Println("Uint32:", s.Uint32.IsPresent(), s.Uint32)
	fmt.Println("Uint64:", s.Uint64.IsPresent(), s.Uint64)
	fmt.Println("Uint64:", s.Uint64.IsPresent(), s.Uint64)
	fmt.Println("Uint:", s.Uint.IsPresent(), s.Uint)
	fmt.Println("Uintptr:", s.Uint.IsPresent(), s.Uint)

	// Output:
	// Bool: true false
	// Byte: true 0
	// Float32: true 0
	// Float64: true 0
	// Int16: true 0
	// Int32: true 0
	// Int64: true 0
	// Int: true 0
	// Rune: true 0
	// String: true string
	// Time: true 0001-01-01 00:00:00 +0000 UTC
	// Uint16: true 0
	// Uint32: true 0
	// Uint64: true 0
	// Uint64: true 0
	// Uint: true 0
	// Uintptr: true 0
}

func Example_xmlMarshalOmitEmpty() {
	s := struct {
		XMLName xml.Name                     `xml:"s"`
		Bool    optional.Optional[bool]      `xml:"bool,omitempty"`
		Byte    optional.Optional[byte]      `xml:"byte,omitempty"`
		Float32 optional.Optional[float32]   `xml:"float32,omitempty"`
		Float64 optional.Optional[float64]   `xml:"float64,omitempty"`
		Int16   optional.Optional[int16]     `xml:"int16,omitempty"`
		Int32   optional.Optional[int32]     `xml:"int32,omitempty"`
		Int64   optional.Optional[int64]     `xml:"int64,omitempty"`
		Int     optional.Optional[int]       `xml:"int,omitempty"`
		Rune    optional.Optional[rune]      `xml:"rune,omitempty"`
		String  optional.Optional[string]    `xml:"string,omitempty"`
		Time    optional.Optional[time.Time] `xml:"time,omitempty"`
		Uint16  optional.Optional[uint16]    `xml:"uint16,omitempty"`
		Uint32  optional.Optional[uint32]    `xml:"uint32,omitempty"`
		Uint64  optional.Optional[uint64]    `xml:"uint64,omitempty"`
		Uint    optional.Optional[uint]      `xml:"uint,omitempty"`
		Uintptr optional.Optional[uintptr]   `xml:"uintptr,omitempty"`
	}{
		Bool:    optional.Empty[bool](),
		Byte:    optional.Empty[byte](),
		Float32: optional.Empty[float32](),
		Float64: optional.Empty[float64](),
		Int16:   optional.Empty[int16](),
		Int32:   optional.Empty[int32](),
		Int64:   optional.Empty[int64](),
		Int:     optional.Empty[int](),
		Rune:    optional.Empty[rune](),
		String:  optional.Empty[string](),
		Time:    optional.Empty[time.Time](),
		Uint16:  optional.Empty[uint16](),
		Uint32:  optional.Empty[uint32](),
		Uint64:  optional.Empty[uint64](),
		Uint:    optional.Empty[uint](),
		Uintptr: optional.Empty[uintptr](),
	}

	output, _ := xml.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// <s></s>
}

func Example_xmlMarshalEmpty() {
	s := struct {
		XMLName xml.Name                     `xml:"s"`
		Bool    optional.Optional[bool]      `xml:"bool"`
		Byte    optional.Optional[byte]      `xml:"byte"`
		Float32 optional.Optional[float32]   `xml:"float32"`
		Float64 optional.Optional[float64]   `xml:"float64"`
		Int16   optional.Optional[int16]     `xml:"int16"`
		Int32   optional.Optional[int32]     `xml:"int32"`
		Int64   optional.Optional[int64]     `xml:"int64"`
		Int     optional.Optional[int]       `xml:"int"`
		Rune    optional.Optional[rune]      `xml:"rune"`
		String  optional.Optional[string]    `xml:"string"`
		Time    optional.Optional[time.Time] `xml:"time"`
		Uint16  optional.Optional[uint16]    `xml:"uint16"`
		Uint32  optional.Optional[uint32]    `xml:"uint32"`
		Uint64  optional.Optional[uint64]    `xml:"uint64"`
		Uint    optional.Optional[uint]      `xml:"uint"`
		Uintptr optional.Optional[uintptr]   `xml:"uintptr"`
	}{
		Bool:    optional.Empty[bool](),
		Byte:    optional.Empty[byte](),
		Float32: optional.Empty[float32](),
		Float64: optional.Empty[float64](),
		Int16:   optional.Empty[int16](),
		Int32:   optional.Empty[int32](),
		Int64:   optional.Empty[int64](),
		Int:     optional.Empty[int](),
		Rune:    optional.Empty[rune](),
		String:  optional.Empty[string](),
		Time:    optional.Empty[time.Time](),
		Uint16:  optional.Empty[uint16](),
		Uint32:  optional.Empty[uint32](),
		Uint64:  optional.Empty[uint64](),
		Uint:    optional.Empty[uint](),
		Uintptr: optional.Empty[uintptr](),
	}

	output, _ := xml.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// <s>
	//   <bool>false</bool>
	//   <byte>0</byte>
	//   <float32>0</float32>
	//   <float64>0</float64>
	//   <int16>0</int16>
	//   <int32>0</int32>
	//   <int64>0</int64>
	//   <int>0</int>
	//   <rune>0</rune>
	//   <string></string>
	//   <time>0001-01-01T00:00:00Z</time>
	//   <uint16>0</uint16>
	//   <uint32>0</uint32>
	//   <uint64>0</uint64>
	//   <uint>0</uint>
	//   <uintptr>0</uintptr>
	// </s>
}

func Example_xmlMarshalPresent() {
	s := struct {
		XMLName xml.Name                     `xml:"s"`
		Bool    optional.Optional[bool]      `xml:"bool"`
		Byte    optional.Optional[byte]      `xml:"byte"`
		Float32 optional.Optional[float32]   `xml:"float32"`
		Float64 optional.Optional[float64]   `xml:"float64"`
		Int16   optional.Optional[int16]     `xml:"int16"`
		Int32   optional.Optional[int32]     `xml:"int32"`
		Int64   optional.Optional[int64]     `xml:"int64"`
		Int     optional.Optional[int]       `xml:"int"`
		Rune    optional.Optional[rune]      `xml:"rune"`
		String  optional.Optional[string]    `xml:"string"`
		Time    optional.Optional[time.Time] `xml:"time"`
		Uint16  optional.Optional[uint16]    `xml:"uint16"`
		Uint32  optional.Optional[uint32]    `xml:"uint32"`
		Uint64  optional.Optional[uint64]    `xml:"uint64"`
		Uint    optional.Optional[uint]      `xml:"uint"`
		Uintptr optional.Optional[uintptr]   `xml:"uintptr"`
	}{
		Bool:    optional.Of(true),
		Byte:    optional.Of[byte](1),
		Float32: optional.Of[float32](2.1),
		Float64: optional.Of(2.2),
		Int16:   optional.Of[int16](3),
		Int32:   optional.Of[int32](4),
		Int64:   optional.Of[int64](5),
		Int:     optional.Of(6),
		Rune:    optional.Of[rune](7),
		String:  optional.Of("string"),
		Time:    optional.Of(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)),
		Uint16:  optional.Of[uint16](8),
		Uint32:  optional.Of[uint32](9),
		Uint64:  optional.Of[uint64](10),
		Uint:    optional.Of[uint](11),
		Uintptr: optional.Of[uintptr](12),
	}

	output, _ := xml.MarshalIndent(s, "", "  ")
	fmt.Println(string(output))

	// Output:
	// <s>
	//   <bool>true</bool>
	//   <byte>1</byte>
	//   <float32>2.1</float32>
	//   <float64>2.2</float64>
	//   <int16>3</int16>
	//   <int32>4</int32>
	//   <int64>5</int64>
	//   <int>6</int>
	//   <rune>7</rune>
	//   <string>string</string>
	//   <time>2006-01-02T15:04:05Z</time>
	//   <uint16>8</uint16>
	//   <uint32>9</uint32>
	//   <uint64>10</uint64>
	//   <uint>11</uint>
	//   <uintptr>12</uintptr>
	// </s>
}

func Example_xmlUnmarshalEmpty() {
	s := struct {
		XMLName xml.Name                     `xml:"s"`
		Bool    optional.Optional[bool]      `xml:"bool"`
		Byte    optional.Optional[byte]      `xml:"byte"`
		Float32 optional.Optional[float32]   `xml:"float32"`
		Float64 optional.Optional[float64]   `xml:"float64"`
		Int16   optional.Optional[int16]     `xml:"int16"`
		Int32   optional.Optional[int32]     `xml:"int32"`
		Int64   optional.Optional[int64]     `xml:"int64"`
		Int     optional.Optional[int]       `xml:"int"`
		Rune    optional.Optional[rune]      `xml:"rune"`
		String  optional.Optional[string]    `xml:"string"`
		Time    optional.Optional[time.Time] `xml:"time"`
		Uint16  optional.Optional[uint16]    `xml:"uint16"`
		Uint32  optional.Optional[uint32]    `xml:"uint32"`
		Uint64  optional.Optional[uint64]    `xml:"uint64"`
		Uint    optional.Optional[uint]      `xml:"uint"`
		Uintptr optional.Optional[uintptr]   `xml:"uintptr"`
	}{}

	x := `<s></s>`
	xml.Unmarshal([]byte(x), &s)
	fmt.Println("Bool:", s.Bool.IsPresent())
	fmt.Println("Byte:", s.Byte.IsPresent())
	fmt.Println("Float32:", s.Float32.IsPresent())
	fmt.Println("Float64:", s.Float64.IsPresent())
	fmt.Println("Int16:", s.Int16.IsPresent())
	fmt.Println("Int32:", s.Int32.IsPresent())
	fmt.Println("Int64:", s.Int64.IsPresent())
	fmt.Println("Int:", s.Int.IsPresent())
	fmt.Println("Rune:", s.Rune.IsPresent())
	fmt.Println("String:", s.String.IsPresent())
	fmt.Println("Time:", s.Time.IsPresent())
	fmt.Println("Uint16:", s.Uint16.IsPresent())
	fmt.Println("Uint32:", s.Uint32.IsPresent())
	fmt.Println("Uint64:", s.Uint64.IsPresent())
	fmt.Println("Uint64:", s.Uint64.IsPresent())
	fmt.Println("Uint:", s.Uint.IsPresent())
	fmt.Println("Uintptr:", s.Uint.IsPresent())

	// Output:
	// Bool: false
	// Byte: false
	// Float32: false
	// Float64: false
	// Int16: false
	// Int32: false
	// Int64: false
	// Int: false
	// Rune: false
	// String: false
	// Time: false
	// Uint16: false
	// Uint32: false
	// Uint64: false
	// Uint64: false
	// Uint: false
	// Uintptr: false
}

func Example_xmlUnmarshalPresent() {
	s := struct {
		XMLName xml.Name                     `xml:"s"`
		Bool    optional.Optional[bool]      `xml:"bool"`
		Byte    optional.Optional[byte]      `xml:"byte"`
		Float32 optional.Optional[float32]   `xml:"float32"`
		Float64 optional.Optional[float64]   `xml:"float64"`
		Int16   optional.Optional[int16]     `xml:"int16"`
		Int32   optional.Optional[int32]     `xml:"int32"`
		Int64   optional.Optional[int64]     `xml:"int64"`
		Int     optional.Optional[int]       `xml:"int"`
		Rune    optional.Optional[rune]      `xml:"rune"`
		String  optional.Optional[string]    `xml:"string"`
		Time    optional.Optional[time.Time] `xml:"time"`
		Uint16  optional.Optional[uint16]    `xml:"uint16"`
		Uint32  optional.Optional[uint32]    `xml:"uint32"`
		Uint64  optional.Optional[uint64]    `xml:"uint64"`
		Uint    optional.Optional[uint]      `xml:"uint"`
		Uintptr optional.Optional[uintptr]   `xml:"uintptr"`
	}{}

	x := `<s>
   <bool>false</bool>
   <byte>0</byte>
   <float32>0</float32>
   <float64>0</float64>
   <int16>0</int16>
   <int32>0</int32>
   <int64>0</int64>
   <int>0</int>
   <rune>0</rune>
   <string>string</string>
   <time>0001-01-01T00:00:00Z</time>
   <uint16>0</uint16>
   <uint32>0</uint32>
   <uint64>0</uint64>
   <uint>0</uint>
   <uintptr>0</uintptr>
 </s>`
	xml.Unmarshal([]byte(x), &s)
	fmt.Println("Bool:", s.Bool.IsPresent(), s.Bool)
	fmt.Println("Byte:", s.Byte.IsPresent(), s.Byte)
	fmt.Println("Float32:", s.Float32.IsPresent(), s.Float32)
	fmt.Println("Float64:", s.Float64.IsPresent(), s.Float64)
	fmt.Println("Int16:", s.Int16.IsPresent(), s.Int16)
	fmt.Println("Int32:", s.Int32.IsPresent(), s.Int32)
	fmt.Println("Int64:", s.Int64.IsPresent(), s.Int64)
	fmt.Println("Int:", s.Int.IsPresent(), s.Int)
	fmt.Println("Rune:", s.Rune.IsPresent(), s.Rune)
	fmt.Println("String:", s.String.IsPresent(), s.String)
	fmt.Println("Time:", s.Time.IsPresent(), s.Time)
	fmt.Println("Uint16:", s.Uint16.IsPresent(), s.Uint16)
	fmt.Println("Uint32:", s.Uint32.IsPresent(), s.Uint32)
	fmt.Println("Uint64:", s.Uint64.IsPresent(), s.Uint64)
	fmt.Println("Uint64:", s.Uint64.IsPresent(), s.Uint64)
	fmt.Println("Uint:", s.Uint.IsPresent(), s.Uint)
	fmt.Println("Uintptr:", s.Uint.IsPresent(), s.Uint)

	// Output:
	// Bool: true false
	// Byte: true 0
	// Float32: true 0
	// Float64: true 0
	// Int16: true 0
	// Int32: true 0
	// Int64: true 0
	// Int: true 0
	// Rune: true 0
	// String: true string
	// Time: true 0001-01-01 00:00:00 +0000 UTC
	// Uint16: true 0
	// Uint32: true 0
	// Uint64: true 0
	// Uint64: true 0
	// Uint: true 0
	// Uintptr: true 0
}
