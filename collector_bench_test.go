package validation_test

import (
	"testing"

	validation "github.com/selfshop-dev/lib-validation"
)

func BenchmarkError_Error(b *testing.B) {
	e := validation.New("invalid user")
	for range 10 {
		e.Add(validation.Required("field"))
	}
	for b.Loop() {
		_ = e.Error()
	}
}

func BenchmarkCollector_Merge(b *testing.B) {
	in := validation.New("invalid address")
	in.Add(
		validation.Required("city"),
		validation.Required("zip_code"),
		validation.Invalid("country", "unrecognised"),
	)
	for b.Loop() {
		c := validation.NewCollector("invalid user")
		c.Merge("shipping_address", in)
	}
}
