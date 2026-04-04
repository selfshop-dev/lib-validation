package validation_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	validation "github.com/selfshop-dev/lib-validation"
)

func TestError_Add(t *testing.T) {
	t.Parallel()

	t.Run("appends single field error", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(validation.Required("email"))

		assert.Len(t, e.Fields, 1)
		assert.Equal(t, "email", e.Fields[0].Field)
	})

	t.Run("appends multiple field errors at once", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(validation.Required("email"), validation.Required("name"))

		assert.Len(t, e.Fields, 2)
	})

	t.Run("appends sequentially", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(validation.Required("email"))
		e.Add(validation.Required("name"))

		assert.Len(t, e.Fields, 2)
	})
}

func TestError_HasErrors(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name   string
		fields []validation.FieldError
		want   bool
	}{
		{name: "no fields", fields: nil, want: false},
		{name: "one field", fields: []validation.FieldError{validation.Required("x")}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := validation.New("summary")
			e.Add(tc.fields...)
			assert.Equal(t, tc.want, e.HasErrors())
		})
	}
}

func TestError_JSON(t *testing.T) {
	t.Parallel()

	e := validation.New("invalid user")
	e.Add(validation.TooLong("username", 50))

	b, err := json.Marshal(e)
	require.NoError(t, err)
	assert.JSONEq(t, `{
        "summary": "invalid user",
        "fields": [
            {
                "field": "username",
                "code": "too_long",
                "message": "username must not exceed 50 characters",
                "meta": {"max": 50}
            }
        ]
    }`, string(b))
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("no fields returns summary only", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		assert.Equal(t, "invalid input", e.Error())
	})

	t.Run("single field appended to summary", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "email is required"})

		assert.Equal(t, "invalid input: [required] email: email is required", e.Error())
	})

	t.Run("multiple fields joined with semicolon", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(
			validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "email is required"},
			validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "name is required"},
		)

		assert.Contains(t, e.Error(), "invalid input:")
		assert.Contains(t, e.Error(), "[required] email: email is required")
		assert.Contains(t, e.Error(), "[required] name: name is required")
	})
}

func TestError_FieldsForInto(t *testing.T) {
	t.Parallel()

	e := validation.New("invalid input")
	e.Add(
		validation.FieldError{Field: "email", Code: validation.CodeRequired},
		validation.FieldError{Field: "email", Code: validation.CodeTooLong},
	)

	buf := make([]validation.FieldError, 0, 4)
	got := e.FieldsForInto("email", buf[:0])
	assert.Len(t, got, 2)
	assert.Equal(t, validation.CodeRequired, got[0].Code)
}

func TestError_FieldsFor(t *testing.T) {
	t.Parallel()

	e := validation.New("invalid input")
	e.Add(
		validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "required"},
		validation.FieldError{Field: "email", Code: validation.CodeTooLong, Message: "too long"},
		validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "required"},
	)

	t.Run("returns all errors for field", func(t *testing.T) {
		t.Parallel()
		got := e.FieldsFor("email")
		assert.Len(t, got, 2)
	})

	t.Run("returns empty for unknown field", func(t *testing.T) {
		t.Parallel()
		got := e.FieldsFor("nonexistent")
		assert.Empty(t, got)
	})
}

func TestError_First(t *testing.T) {
	t.Parallel()

	e := validation.New("invalid input")
	e.Add(
		validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "required"},
		validation.FieldError{Field: "email", Code: validation.CodeTooLong, Message: "too long"},
	)

	t.Run("returns first match", func(t *testing.T) {
		t.Parallel()
		got, ok := e.First("email")
		require.True(t, ok)
		assert.Equal(t, validation.CodeRequired, got.Code)
	})

	t.Run("returns false for unknown field", func(t *testing.T) {
		t.Parallel()
		fe, ok := e.First("nonexistent")
		assert.False(t, ok)
		assert.Empty(t, fe.Field)
	})
}

func TestError_FirstWithCode(t *testing.T) {
	t.Parallel()

	e := validation.New("invalid input")
	e.Add(
		validation.FieldError{Field: "password", Code: validation.CodeTooShort, Message: "too short"},
		validation.FieldError{Field: "password", Code: validation.CodeTooLong, Message: "too long"},
		validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "required"},
	)

	t.Run("returns first match by field and code", func(t *testing.T) {
		t.Parallel()
		got, ok := e.FirstWithCode("password", validation.CodeTooLong)
		require.True(t, ok)
		assert.Equal(t, "too long", got.Message)
	})

	t.Run("returns false when field matches but code does not", func(t *testing.T) {
		t.Parallel()
		_, ok := e.FirstWithCode("password", validation.CodeInvalid)
		assert.False(t, ok)
	})

	t.Run("returns false when field does not exist", func(t *testing.T) {
		t.Parallel()
		_, ok := e.FirstWithCode("nonexistent", validation.CodeRequired)
		assert.False(t, ok)
	})
}

func TestError_Codes(t *testing.T) {
	t.Parallel()

	t.Run("returns unique codes only", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid input")
		e.Add(
			validation.FieldError{Field: "a", Code: validation.CodeRequired},
			validation.FieldError{Field: "b", Code: validation.CodeRequired},
			validation.FieldError{Field: "c", Code: validation.CodeTooLong},
		)

		codes := e.Codes()
		assert.Len(t, codes, 2)
		assert.Contains(t, codes, validation.CodeRequired)
		assert.Contains(t, codes, validation.CodeTooLong)
	})

	t.Run("empty error returns empty codes", func(t *testing.T) {
		t.Parallel()
		e := validation.New("summary")
		assert.Empty(t, e.Codes())
	})
}

func TestAs(t *testing.T) {
	t.Parallel()

	t.Run("unwraps *Error from error chain", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid")
		e.Add(validation.Required("x"))

		wrapped := fmt.Errorf("wrapped: %w", e)
		got, ok := validation.As(wrapped)

		require.True(t, ok)
		assert.Equal(t, "invalid", got.Summary)
	})

	t.Run("returns false for non-validation error", func(t *testing.T) {
		t.Parallel()
		ve, ok := validation.As(errors.New("plain error"))
		assert.False(t, ok)
		assert.Nil(t, ve)
	})

	t.Run("returns false for nil", func(t *testing.T) {
		t.Parallel()
		ve, ok := validation.As(nil)
		assert.False(t, ok)
		assert.Nil(t, ve)
	})
}

func TestIs(t *testing.T) {
	t.Parallel()

	t.Run("returns true for *Error", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid")
		assert.True(t, validation.Is(e))
	})

	t.Run("returns true for wrapped *Error", func(t *testing.T) {
		t.Parallel()
		e := validation.New("invalid")
		wrapped := fmt.Errorf("outer: %w", e)
		assert.True(t, validation.Is(wrapped))
	})

	t.Run("returns false for plain error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, validation.Is(errors.New("plain")))
	})

	t.Run("returns false for nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, validation.Is(nil))
	})
}
