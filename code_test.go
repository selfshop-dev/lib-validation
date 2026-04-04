package validation_test // code_test.go

import (
	"testing"

	"github.com/stretchr/testify/assert"

	validation "github.com/selfshop-dev/lib-validation"
)

func TestCode_Values(t *testing.T) {
	t.Parallel()

	// Codes are a public API contract — their string values must never change.
	testCases := [...]struct {
		code validation.Code
		want string
	}{
		{validation.CodeRequired, "required"},
		{validation.CodeConflict, "conflict"},
		{validation.CodeTooLong, "too_long"},
		{validation.CodeUnknown, "unknown"},
		{validation.CodeInvalid, "invalid"},
		{validation.CodeOutOfRange, "out_of_range"},
		{validation.CodeTooShort, "too_short"},
		{validation.CodeImmutable, "immutable"},
		{validation.CodeTypeMismatch, "type_mismatch"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.code), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, validation.Code(tc.want), tc.code)
		})
	}
}

func TestCode_Uniqueness(t *testing.T) {
	t.Parallel()

	all := []validation.Code{
		validation.CodeRequired,
		validation.CodeConflict,
		validation.CodeTooLong,
		validation.CodeUnknown,
		validation.CodeInvalid,
		validation.CodeOutOfRange,
		validation.CodeTooShort,
		validation.CodeImmutable,
		validation.CodeTypeMismatch,
	}

	seen := make(map[validation.Code]struct{}, len(all))
	for _, c := range all {
		_, duplicate := seen[c]
		assert.False(t, duplicate, "duplicate code value: %q", c)
		seen[c] = struct{}{}
	}
}
