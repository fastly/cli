/*
Package optional exports an Optional[T] type that can wrap any type to represent the lack of value. The types guarantee safety by requiring the developer to unwrap them to get to the inner value. This prevents a nil value being operated on. Optionals marshal to XML and JSON like their underlying type, and omitempty works just like their wrapped type would with a pointer, but without the use of pointers.

These types are an alternative to using pointers, zero values, or similar null wrapper packages. Unlike similar solutions these will omit correctly from XML and JSON without the use of pointers and the compiler will ensure their value is not used when empty.

# Examples

Wrap a value in an optional:

	var i int = ...
	o := optional.Of(i)

Or, create an empty optional:

	o := optional.Empty[int]()

Or, wrap a pointer in an optional:

	var i *int = ...
	o := optional.OfPtr(i)

Unwrap it safely:

	o.If(func(i int) {
		// called if o is not empty
	})

	if _, ok := o.Get(); ok {
		// ok is true if o is not empty
	}

Or get it's value with a fallback to a default:

	_ := o.ElseZero() // returns the zero value if empty

	_ := o.Else(100) // returns 100 if o is empty

	_ := o.ElseFunc(func() {
		// called if o is empty
		return 100
	})

XML and JSON are supported out of the box. Use `omitempty` to omit the field when the optional is empty:

	s := struct {
		Int1 optional.Optional[int] `json:"int1,omitempty"`
		Int2 optional.Optional[int] `json:"int2,omitempty"`
		Int3 optional.Optional[int] `json:"int3,omitempty"`
	}{
		Int1: optional.Empty[int](),
		Int2: optional.Of(1000),
		Int3: optional.OfPtr(nil),
	}

	output, _ := json.Marshal(s)

	// output = {"int2":1000}
*/
package optional
