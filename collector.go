package validation

import (
	"slices"
	"strings"
)

const (
	avgFieldNameLen = 32 // reasonable average length of a dot-notation field name
)

// Collector accumulates FieldErrors during a multi-field validation pass
// and returns a *Error (or nil) at the end.
type Collector struct {
	err *Error
}

// NewCollector creates a Collector with the given summary message.
// The summary becomes [Error.Summary] if any field errors are collected.
func NewCollector(summary string) *Collector {
	return &Collector{err: New(summary)}
}

// Check adds fe when ok is false. Chainable.
func (c *Collector) Check(ok bool, fe FieldError) *Collector {
	if !ok {
		c.err.Add(fe)
	}
	return c
}

// Fail adds fe when bad is true. Chainable.
func (c *Collector) Fail(bad bool, fe FieldError) *Collector { return c.Check(!bad, fe) }

// Add unconditionally appends one or more FieldErrors.
func (c *Collector) Add(fes ...FieldError) *Collector {
	c.err.Add(fes...)
	return c
}

// Merge preserves the original order of FieldErrors from src.
// Fields are appended in the order they were added to the source validator,
// which means the final error retains a deterministic, top-to-bottom field order
// matching the validation pass — useful for rendering form errors in document order.
//
// Namespaces are joined with a dot, so nested calls produce deep paths:
//
//	c.Merge("order.shipping", validateAddress(req.Shipping))
//	// field "city" in validateAddress → "order.shipping.city"
func (c *Collector) Merge(namespace string, incoming error) *Collector {
	if incoming == nil {
		return c
	}
	ve, ok := As(incoming)
	if !ok {
		c.err.Add(FieldError{Field: namespace, Code: CodeInvalid, Message: incoming.Error()})
		return c
	}
	c.err.Fields = slices.Grow(c.err.Fields, len(ve.Fields))
	if namespace == "" {
		c.err.Fields = append(c.err.Fields, ve.Fields...)
		return c
	}
	// Build prefixed field names via a single Builder, resetting between iterations
	// to avoid a separate string allocation per field.
	var b strings.Builder
	b.Grow(len(namespace) + 1 + avgFieldNameLen)
	for _, fe := range ve.Fields {
		if fe.Field == "" {
			fe.Field = namespace
		} else {
			b.WriteString(namespace)
			b.WriteByte('.')
			b.WriteString(fe.Field)
			fe.Field = b.String()
			b.Reset()
		}
		c.err.Fields = append(c.err.Fields, fe)
	}
	return c
}

// Err returns *[Error] if any fields were collected, nil otherwise.
func (c *Collector) Err() error {
	if c.err.HasErrors() {
		return c.err
	}
	return nil
}

// Validation returns *[Error] directly (not error interface) for field inspection.
func (c *Collector) Validation() *Error {
	if c.err.HasErrors() {
		return c.err
	}
	return nil
}
