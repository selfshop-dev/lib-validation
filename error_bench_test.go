package validation_test

import (
	"testing"

	validation "github.com/selfshop-dev/lib-validation"
)

// BenchmarkError_Error_Single measures Error() with one FieldError —
// returns summary only, no strings.Join involved.
func BenchmarkError_Error_Single(b *testing.B) {
	e := validation.New("invalid input")
	e.Add(validation.Required("email"))
	b.ReportAllocs()
	for b.Loop() {
		_ = e.Error()
	}
}

// BenchmarkError_Error_Many measures Error() with 10 FieldErrors —
// exercises the strings.Join + per-field allocation path.
func BenchmarkError_Error_Many(b *testing.B) {
	e := validation.New("invalid input")
	e.Add(
		validation.Required("f1"),
		validation.Required("f2"),
		validation.Invalid("f3", "bad value"),
		validation.TooLong("f4", 100),
		validation.TooShort("f5", 3),
		validation.OutOfRange("f6", 0, 10),
		validation.Immutable("f7"),
		validation.Conflict("f8", "already exists"),
		validation.Unknown("f9"),
		validation.TypeMismatch("f10", "integer"),
	)
	b.ReportAllocs()
	for b.Loop() {
		_ = e.Error()
	}
}

// BenchmarkError_Codes measures the unique-code extraction loop with
// a realistic mix of repeated and distinct codes.
func BenchmarkError_Codes(b *testing.B) {
	e := validation.New("invalid input")
	for range 5 {
		e.Add(validation.Required("f"))
	}
	for range 5 {
		e.Add(validation.TooLong("f", 100))
	}
	e.Add(validation.Invalid("f", "bad"))
	b.ReportAllocs()
	for b.Loop() {
		_ = e.Codes()
	}
}

// BenchmarkError_FieldsFor measures the linear scan over Fields.
// Relevant when called repeatedly per request (e.g. rendering per-field hints).
func BenchmarkError_FieldsFor(b *testing.B) {
	e := validation.New("invalid input")
	for i := range 20 {
		e.Add(validation.Invalid("field_"+string(rune('a'+i)), "bad"))
	}
	// target field is last — worst case for linear scan
	e.Add(validation.Required("target"))
	b.ReportAllocs()
	for b.Loop() {
		_ = e.FieldsFor("target")
	}
}
