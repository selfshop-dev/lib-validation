package validation_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	validation "github.com/selfshop-dev/lib-validation"
)

func TestCollector_Err_NilWhenNoErrors(t *testing.T) {
	t.Parallel()

	c := validation.NewCollector("invalid user")
	require.NoError(t, c.Err())
	assert.Nil(t, c.Validation())
}

func TestCollector_Check(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name      string
		ok        bool
		wantCount int
	}{
		{name: "ok=true does not add error", ok: true, wantCount: 0},
		{name: "ok=false adds error", ok: false, wantCount: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validation.NewCollector("summary")
			c.Check(tc.ok, validation.Required("field"))

			if tc.wantCount == 0 {
				assert.NoError(t, c.Err())
			} else {
				require.Error(t, c.Err())
				assert.Len(t, c.Validation().Fields, tc.wantCount)
			}
		})
	}
}

func TestCollector_Check_Chainable(t *testing.T) {
	t.Parallel()

	c := validation.NewCollector("invalid user")
	c.Check(false, validation.Required("email")).
		Check(false, validation.Required("name")).
		Check(true, validation.Required("age"))

	require.NotNil(t, c.Validation())
	assert.Len(t, c.Validation().Fields, 2)
}

func TestCollector_Fail(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name      string
		bad       bool
		wantCount int
	}{
		{name: "bad=true adds error", bad: true, wantCount: 1},
		{name: "bad=false does not add error", bad: false, wantCount: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validation.NewCollector("summary")
			c.Fail(tc.bad, validation.Required("field"))

			if tc.wantCount == 0 {
				assert.NoError(t, c.Err())
			} else {
				require.Error(t, c.Err())
				assert.Len(t, c.Validation().Fields, tc.wantCount)
			}
		})
	}
}

func TestCollector_Fail_Chainable(t *testing.T) {
	t.Parallel()

	c := validation.NewCollector("invalid user")
	c.Fail(true, validation.Required("email")).
		Fail(true, validation.Required("name")).
		Fail(false, validation.Required("age"))

	require.NotNil(t, c.Validation())
	assert.Len(t, c.Validation().Fields, 2)
}

func TestCollector_Add(t *testing.T) {
	t.Parallel()

	t.Run("adds single error", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("summary")
		c.Add(validation.Required("email"))

		require.NotNil(t, c.Validation())
		assert.Len(t, c.Validation().Fields, 1)
	})

	t.Run("adds multiple errors at once", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("summary")
		c.Add(validation.Required("email"), validation.Required("name"))

		require.NotNil(t, c.Validation())
		assert.Len(t, c.Validation().Fields, 2)
	})
}

func TestCollector_Merge(t *testing.T) {
	t.Parallel()

	t.Run("nil src is a no-op", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("summary")
		c.Merge("address", nil)
		assert.NoError(t, c.Err())
	})

	t.Run("prefixes field with namespace", func(t *testing.T) {
		t.Parallel()
		inner := validation.New("invalid address")
		inner.Add(validation.Required("zip_code"))

		c := validation.NewCollector("invalid user")
		c.Merge("address", inner)

		require.NotNil(t, c.Validation())
		assert.Equal(t, "address.zip_code", c.Validation().Fields[0].Field)
	})

	t.Run("entity-level inner error gets namespace as field", func(t *testing.T) {
		t.Parallel()
		inner := validation.New("invalid address")
		inner.Add(validation.Entity(validation.CodeInvalid, "missing required fields"))

		c := validation.NewCollector("invalid user")
		c.Merge("address", inner)

		require.NotNil(t, c.Validation())
		assert.Equal(t, "address", c.Validation().Fields[0].Field)
	})

	t.Run("empty namespace preserves original field", func(t *testing.T) {
		t.Parallel()
		inner := validation.New("invalid")
		inner.Add(validation.Required("zip_code"))

		c := validation.NewCollector("summary")
		c.Merge("", inner)

		require.NotNil(t, c.Validation())
		assert.Equal(t, "zip_code", c.Validation().Fields[0].Field)
	})

	t.Run("non-validation error wrapped as CodeInvalid with namespace field", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("summary")
		c.Merge("config", errors.New("file not found"))

		require.NotNil(t, c.Validation())
		fe := c.Validation().Fields[0]
		assert.Equal(t, "config", fe.Field)
		assert.Equal(t, validation.CodeInvalid, fe.Code)
		assert.Equal(t, "file not found", fe.Message)
	})

	t.Run("merges multiple fields from nested validator", func(t *testing.T) {
		t.Parallel()
		inner := validation.New("invalid address")
		inner.Add(validation.Required("city"), validation.Required("country"))

		c := validation.NewCollector("invalid user")
		c.Merge("address", inner)

		require.NotNil(t, c.Validation())
		fields := c.Validation().Fields
		assert.Len(t, fields, 2)
		assert.Equal(t, "address.city", fields[0].Field)
		assert.Equal(t, "address.country", fields[1].Field)
	})
}

func TestCollector_Validation(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no errors", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("summary")
		assert.Nil(t, c.Validation())
	})

	t.Run("returns *Error with correct summary", func(t *testing.T) {
		t.Parallel()
		c := validation.NewCollector("invalid order")
		c.Add(validation.Required("item_id"))

		e := c.Validation()
		require.NotNil(t, e)
		assert.Equal(t, "invalid order", e.Summary)
	})
}

func TestCollector_Err_ImplementsError(t *testing.T) {
	t.Parallel()

	c := validation.NewCollector("invalid request")
	c.Add(validation.Required("email"))

	err := c.Err()
	require.Error(t, err)

	ve, ok := validation.As(err)
	require.True(t, ok)
	assert.Equal(t, "invalid request", ve.Summary)
}
