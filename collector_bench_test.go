package validation_test

import (
	"runtime"
	"testing"

	validation "github.com/selfshop-dev/lib-validation"
)

// BenchmarkCollector_Check measures the cost of a typical multi-field validation
// pass — the most common usage pattern in HTTP handlers.
func BenchmarkCollector_Check(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		c := validation.NewCollector("invalid user")
		c.Check(false, validation.Required("name"))
		c.Check(false, validation.Required("email"))
		c.Check(false, validation.OutOfRange("age", 18, 120))
		err := c.Err()
		runtime.KeepAlive(err)
	}
}

// BenchmarkCollector_Merge_Small measures Merge with a shallow nested validator
// (3 fields) — common for address/contact sub-structs.
func BenchmarkCollector_Merge_Small(b *testing.B) {
	inner := validation.New("invalid address")
	inner.Add(
		validation.Required("city"),
		validation.Required("country"),
		validation.Invalid("zip_code", "must be 5 digits"),
	)
	b.ReportAllocs()
	for b.Loop() {
		c := validation.NewCollector("invalid user")
		c.Merge("address", inner)
		err := c.Err()
		runtime.KeepAlive(err)
	}
}

// BenchmarkCollector_Merge_Large measures Merge with a wide nested validator
// (20 fields) — stress-tests the prefix string allocation loop.
func BenchmarkCollector_Merge_Large(b *testing.B) {
	inner := validation.New("invalid payload")
	for i := range 20 {
		inner.Add(validation.Invalid("field_"+string(rune('a'+i)), "invalid"))
	}
	b.ReportAllocs()
	for b.Loop() {
		c := validation.NewCollector("invalid request")
		c.Merge("payload", inner)
		err := c.Err()
		runtime.KeepAlive(err)
	}
}

// BenchmarkCollector_Merge_NonValidationError measures the fallback branch where
// src is a plain error (not a *Error) — allocates a FieldError on every call.
func BenchmarkCollector_Merge_NonValidationError(b *testing.B) {
	src := &opaqueError{}
	b.ReportAllocs()
	for b.Loop() {
		c := validation.NewCollector("summary")
		c.Merge("cfg", src)
		err := c.Err()
		runtime.KeepAlive(err)
	}
}

// opaqueError is a plain error that does not implement *validation.Error,
// used to exercise the non-validation fallback branch in Merge.
type opaqueError struct{}

func (opaqueError) Error() string { return "something went wrong" }
