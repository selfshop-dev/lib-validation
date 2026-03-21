package validation_test

import (
	"runtime"
	"testing"

	validation "github.com/selfshop-dev/lib-validation"
)

// BenchmarkFieldError_WithMetaPair_NilMeta measures the fast path:
// allocating a fresh map when Meta is nil.
func BenchmarkFieldError_WithMetaPair_NilMeta(b *testing.B) {
	fe := validation.FieldError{Field: "age", Code: validation.CodeOutOfRange}
	b.ReportAllocs()
	for b.Loop() {
		got := fe.WithMetaPair("max", 120)
		runtime.KeepAlive(got)
	}
}

// BenchmarkFieldError_WithMetaPair_ExistingMeta measures the copy path:
// copying an existing map before adding the new key.
// This is the common case — TooLong/TooShort/OutOfRange all pre-populate Meta.
func BenchmarkFieldError_WithMetaPair_ExistingMeta(b *testing.B) {
	fe := validation.TooLong("username", 50) // Meta already has "max"
	b.ReportAllocs()
	for b.Loop() {
		got := fe.WithMetaPair("actual_len", 75)
		runtime.KeepAlive(got)
	}
}
