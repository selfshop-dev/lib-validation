package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	validation "github.com/selfshop-dev/lib-validation"
)

func TestRequired(t *testing.T) {
	t.Parallel()

	got := validation.Required("email")

	assert.Equal(t, "email", got.Field)
	assert.Equal(t, validation.CodeRequired, got.Code)
	assert.Equal(t, "email is required", got.Message)
}

func TestInvalid(t *testing.T) {
	t.Parallel()

	got := validation.Invalid("email", "must be a valid address")

	assert.Equal(t, "email", got.Field)
	assert.Equal(t, validation.CodeInvalid, got.Code)
	assert.Equal(t, "must be a valid address", got.Message)
}

func TestUnknown(t *testing.T) {
	t.Parallel()

	got := validation.Unknown("extra_field")

	assert.Equal(t, "extra_field", got.Field)
	assert.Equal(t, validation.CodeUnknown, got.Code)
	assert.Equal(t, "extra_field is not a recognised key", got.Message)
}

func TestConflict(t *testing.T) {
	t.Parallel()

	got := validation.Conflict("email", "already taken")

	assert.Equal(t, "email", got.Field)
	assert.Equal(t, validation.CodeConflict, got.Code)
	assert.Equal(t, "already taken", got.Message)
}

func TestTooLong(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		field   string
		wantMsg string
		maxLen  int
		wantMax int
	}{
		{name: "username 50", field: "username", maxLen: 50, wantMsg: "username must not exceed 50 characters", wantMax: 50},
		{name: "bio 500", field: "bio", maxLen: 500, wantMsg: "bio must not exceed 500 characters", wantMax: 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := validation.TooLong(tc.field, tc.maxLen)

			assert.Equal(t, tc.field, got.Field)
			assert.Equal(t, validation.CodeTooLong, got.Code)
			assert.Equal(t, tc.wantMsg, got.Message)
			assert.Equal(t, tc.wantMax, got.Meta["max"])
		})
	}
}

func TestTooShort(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		field   string
		wantMsg string
		minLen  int
		wantMin int
	}{
		{name: "password 8", field: "password", minLen: 8, wantMsg: "password must be at least 8 characters", wantMin: 8},
		{name: "name 2", field: "name", minLen: 2, wantMsg: "name must be at least 2 characters", wantMin: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := validation.TooShort(tc.field, tc.minLen)

			assert.Equal(t, tc.field, got.Field)
			assert.Equal(t, validation.CodeTooShort, got.Code)
			assert.Equal(t, tc.wantMsg, got.Message)
			assert.Equal(t, tc.wantMin, got.Meta["min"])
		})
	}
}

func TestOutOfRange(t *testing.T) {
	t.Parallel()

	got := validation.OutOfRange("age", 18, 120)

	assert.Equal(t, "age", got.Field)
	assert.Equal(t, validation.CodeOutOfRange, got.Code)
	assert.Equal(t, "age must be between 18 and 120", got.Message)
	assert.Equal(t, 18, got.Meta["min"])
	assert.Equal(t, 120, got.Meta["max"])
}

func TestImmutable(t *testing.T) {
	t.Parallel()

	got := validation.Immutable("user_id")

	assert.Equal(t, "user_id", got.Field)
	assert.Equal(t, validation.CodeImmutable, got.Code)
	assert.Equal(t, "user_id cannot be changed after creation", got.Message)
}

func TestTypeMismatch(t *testing.T) {
	t.Parallel()

	got := validation.TypeMismatch("count", "integer")

	assert.Equal(t, "count", got.Field)
	assert.Equal(t, validation.CodeTypeMismatch, got.Code)
	assert.Equal(t, "count must be of type integer", got.Message)
	assert.Equal(t, "integer", got.Meta["expected_type"])
}

func TestEntity(t *testing.T) {
	t.Parallel()

	got := validation.Entity(validation.CodeConflict, "duplicate entry")

	assert.Empty(t, got.Field)
	assert.Equal(t, validation.CodeConflict, got.Code)
	assert.Equal(t, "duplicate entry", got.Message)
}
