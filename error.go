package validation

import (
	"errors"
	"strings"
)

const (
	sepLen           = 2  // length of separator between field errors: "; " and ": "
	avgFieldErrorLen = 32 // reasonable average serialised length per FieldError
)

// Error has two distinct lifecycle phases that the type system cannot enforce:
//
//	Building phase: *Error is mutated via Add during a validation pass.
//	Do not share a *Error across goroutines during this phase.
//
//	Immutable phase: once returned from Collector.Err or Collector.Validation,
//	the *Error must be treated as read-only. Callers must not call Add after
//	this point — doing so will corrupt shared state if the error is passed
//	across goroutine boundaries (e.g. stored in a context, sent over a channel).
//
// The idiomatic pattern is to build via NewCollector and never retain a
// *Error pointer after returning it as error from a validation function.
type Error struct {
	Summary string       `json:"summary"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// New creates an empty Error with a summary message.
func New(summary string) *Error {
	return &Error{
		Summary: summary,
		Fields:  make([]FieldError, 0, 4),
	}
}

// Add appends one or more FieldErrors to the Error.
func (e *Error) Add(fes ...FieldError) {
	e.Fields = append(e.Fields, fes...)
}

// HasErrors reports whether any FieldErrors have been accumulated.
func (e *Error) HasErrors() bool { return len(e.Fields) > 0 }

func (e *Error) Error() string {
	n := len(e.Fields)
	if n == 0 {
		return e.Summary
	}
	var b strings.Builder
	b.Grow(len(e.Summary) + sepLen + n*avgFieldErrorLen)
	b.WriteString(e.Summary)
	b.WriteString(": ")
	for i, fe := range e.Fields {
		if i > 0 {
			b.WriteString("; ")
		}
		// Inline fe.Error() to avoid per-field string allocation.
		b.WriteByte('[')
		b.WriteString(string(fe.Code))
		b.WriteByte(']')
		b.WriteByte(' ')
		if fe.Field != "" {
			b.WriteString(fe.Field)
			b.WriteByte(':')
			b.WriteByte(' ')
		}
		b.WriteString(fe.Message)
	}
	return b.String()
}

// FieldsFor returns all FieldErrors for the given dot-notation path.
func (e *Error) FieldsFor(field string) []FieldError {
	return e.FieldsForInto(field, nil)
}

// FieldsForInto appends all FieldErrors for the given dot-notation path into dst
// and returns the extended slice. Pass a pre-allocated slice to avoid allocation:
//
//	buf := make([]FieldError, 0, 4)
//	errs := ve.FieldsForInto("email", buf[:0])
func (e *Error) FieldsForInto(field string, dst []FieldError) []FieldError {
	for _, fe := range e.Fields {
		if fe.Field == field {
			dst = append(dst, fe)
		}
	}
	return dst
}

// First returns the first FieldError for the given path, or (zero, false).
func (e *Error) First(field string) (FieldError, bool) {
	for _, fe := range e.Fields {
		if fe.Field == field {
			return fe, true
		}
	}
	return FieldError{}, false
}

// FirstWithCode returns the first FieldError for the given field path and code,
// or (zero, false) if no match is found.
//
// Use when a single field may carry multiple errors of different codes and you
// need to distinguish between them:
//
//	fe, ok := ve.FirstWithCode("password", validation.CodeTooShort)
//	if ok {
//	    fmt.Println("minimum length:", fe.Meta["min"])
//	}
//
// For retrieving all errors on a field regardless of code, use [Error.FieldsFor].
// For retrieving the first error on a field regardless of code, use [Error.First].
func (e *Error) FirstWithCode(field string, code Code) (FieldError, bool) {
	for _, fe := range e.Fields {
		if fe.Field == field && fe.Code == code {
			return fe, true
		}
	}
	return FieldError{}, false
}

// Codes returns the unique set of Codes present across all FieldErrors.
func (e *Error) Codes() []Code {
	n := len(e.Fields)
	seen := make(map[Code]struct{}, n)
	out := make([]Code, 0, n)
	for _, fe := range e.Fields {
		if _, ok := seen[fe.Code]; !ok {
			seen[fe.Code] = struct{}{}
			out = append(out, fe.Code)
		}
	}
	return out
}

// As is a typed convenience wrapper around errors.As.
// Prefer this over the standard errors.As when you need the concrete *Error
// for field inspection (First, FieldsFor, Codes). Use the standard errors.Is
// pattern only when you need to check presence without inspecting fields.
//
// Note: As returns (nil, false) for a nil error argument — it does not panic.
func As(err error) (*Error, bool) {
	return errors.AsType[*Error](err)
}

// Is reports whether err chain contains a *Error.
func Is(err error) bool { _, ok := As(err); return ok } //nolint:errcheck // *Error is not an error return value; blank is intentional
