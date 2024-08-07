package optional

import "testing"

func TestIsPresent(t *testing.T) {
	s := "ptr to string"
	tests := []struct {
		Optional          Optional[string]
		ExpectedIsPresent bool
	}{
		{Empty[string](), false},
		{Of(""), true},
		{Of("string"), true},
		{OfPtr((*string)(nil)), false},
		{OfPtr((*string)(&s)), true},
	}

	for _, test := range tests {
		isPresent := test.Optional.IsPresent()

		if isPresent != test.ExpectedIsPresent {
			t.Errorf("%#v IsPresent got %#v, want %#v", test.Optional, isPresent, test.ExpectedIsPresent)
		}
	}
}

func TestGet(t *testing.T) {
	s := "ptr to string"
	tests := []struct {
		Optional      Optional[string]
		ExpectedValue string
		ExpectedOk    bool
	}{
		{Empty[string](), "", false},
		{Of(""), "", true},
		{Of("string"), "string", true},
		{OfPtr((*string)(nil)), "", false},
		{OfPtr((*string)(&s)), "ptr to string", true},
	}

	for _, test := range tests {
		value, ok := test.Optional.Get()

		if value != test.ExpectedValue || ok != test.ExpectedOk {
			t.Errorf("%#v Get got %#v, %#v, want %#v, %#v", test.Optional, ok, test.ExpectedOk, value, test.ExpectedValue)
		}
	}
}

func TestIfPresent(t *testing.T) {
	s := "ptr to string"
	tests := []struct {
		Optional       Optional[string]
		ExpectedCalled bool
		IfCalledValue  string
	}{
		{Empty[string](), false, ""},
		{Of(""), true, ""},
		{Of("string"), true, "string"},
		{OfPtr((*string)(nil)), false, ""},
		{OfPtr((*string)(&s)), true, "ptr to string"},
	}

	for _, test := range tests {
		called := false
		test.Optional.If(func(v string) {
			called = true
			if v != test.IfCalledValue {
				t.Errorf("%#v IfPresent got %#v, want #%v", test.Optional, v, test.IfCalledValue)
			}
		})

		if called != test.ExpectedCalled {
			t.Errorf("%#v IfPresent called %#v, want %#v", test.Optional, called, test.ExpectedCalled)
		}
	}
}

func TestElse(t *testing.T) {
	s := "ptr to string"
	const orElse = "orelse"
	tests := []struct {
		Optional       Optional[string]
		ExpectedResult string
	}{
		{Empty[string](), orElse},
		{Of(""), ""},
		{Of("string"), "string"},
		{OfPtr((*string)(nil)), orElse},
		{OfPtr((*string)(&s)), "ptr to string"},
	}

	for _, test := range tests {
		result := test.Optional.Else(orElse)

		if result != test.ExpectedResult {
			t.Errorf("%#v OrElse(%#v) got %#v, want %#v", test.Optional, orElse, result, test.ExpectedResult)
		}
	}
}

func TestElseFunc(t *testing.T) {
	s := "ptr to string"
	const orElse = "orelse"
	tests := []struct {
		Optional       Optional[string]
		ExpectedResult string
	}{
		{Empty[string](), orElse},
		{Of(""), ""},
		{Of("string"), "string"},
		{OfPtr((*string)(nil)), orElse},
		{OfPtr((*string)(&s)), "ptr to string"},
	}

	for _, test := range tests {
		result := test.Optional.ElseFunc(func() string { return orElse })

		if result != test.ExpectedResult {
			t.Errorf("%#v OrElse(%#v) got %#v, want %#v", test.Optional, orElse, result, test.ExpectedResult)
		}
	}
}

func TestElseZero(t *testing.T) {
	s := "ptr to string"
	tests := []struct {
		Optional       Optional[string]
		ExpectedResult string
	}{
		{Empty[string](), ""},
		{Of(""), ""},
		{Of("string"), "string"},
		{OfPtr((*string)(nil)), ""},
		{OfPtr((*string)(&s)), "ptr to string"},
	}

	for _, test := range tests {
		result := test.Optional.ElseZero()

		if result != test.ExpectedResult {
			t.Errorf("%#v ElseZero() got %#v, want %#v", test.Optional, result, test.ExpectedResult)
		}
	}
}
