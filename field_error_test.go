package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	validation "github.com/selfshop-dev/lib-validation"
)

func TestFieldError_Error(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name string
		fe   validation.FieldError
		want string
	}{
		{
			name: "with field",
			fe:   validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "email is required"},
			want: "[required] email: email is required",
		},
		{
			name: "without field (entity-level)",
			fe:   validation.FieldError{Code: validation.CodeInvalid, Message: "entity is invalid"},
			want: "[invalid] entity is invalid",
		},
		{
			name: "nested dot-notation field",
			fe:   validation.FieldError{Field: "address.zip_code", Code: validation.CodeTooLong, Message: "address.zip_code must not exceed 10 characters"},
			want: "[too_long] address.zip_code: address.zip_code must not exceed 10 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.fe.Error())
		})
	}
}

func TestFieldError_WithValue(t *testing.T) {
	t.Parallel()

	fe := validation.FieldError{Field: "username", Code: validation.CodeInvalid, Message: "invalid"}
	got := fe.WithValue("bad-value")

	assert.Equal(t, "bad-value", got.Value)
	assert.Empty(t, fe.Value, "original must not be mutated")
}

func TestFieldError_WithMetaPair(t *testing.T) {
	t.Parallel()

	t.Run("adds key to nil meta", func(t *testing.T) {
		t.Parallel()
		fe := validation.FieldError{Field: "age", Code: validation.CodeOutOfRange}
		got := fe.WithMetaPair("max", 120)

		assert.Equal(t, map[string]any{"max": 120}, got.Meta)
		assert.Nil(t, fe.Meta, "original must not be mutated")
	})

	t.Run("adds key to existing meta", func(t *testing.T) {
		t.Parallel()
		fe := validation.FieldError{
			Field: "age",
			Code:  validation.CodeOutOfRange,
			Meta:  map[string]any{"min": 18},
		}
		got := fe.WithMetaPair("max", 120)

		assert.Equal(t, 120, got.Meta["max"])
		assert.Equal(t, 18, got.Meta["min"])
	})

	t.Run("does not mutate original meta map", func(t *testing.T) {
		t.Parallel()
		original := map[string]any{"min": 18}
		fe := validation.FieldError{Meta: original}
		got := fe.WithMetaPair("max", 120)

		assert.NotContains(t, original, "max", "original map must not be mutated")
		assert.Equal(t, 120, got.Meta["max"], "returned copy must contain new key")
	})
}
