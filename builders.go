package validation

import "fmt"

// Required returns a [FieldError] for a missing/zero-value field.
func Required(field string) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeRequired,
		Message: field + " is required",
	}
}

// Invalid returns a [FieldError] for a present but invalid value.
func Invalid(field, reason string) FieldError {
	return FieldError{Field: field, Code: CodeInvalid, Message: reason}
}

// Unknown returns a [FieldError] for an unrecognised field/key.
// Primary consumer: lib-config strict-mode unknown key detection.
func Unknown(field string) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeUnknown,
		Message: field + " is not a recognised key",
	}
}

// Conflict returns a [FieldError] for a value conflicting with existing state.
func Conflict(field, reason string) FieldError {
	return FieldError{Field: field, Code: CodeConflict, Message: reason}
}

// TooLong returns a [FieldError] for a string/slice exceeding maxLen length.
func TooLong(field string, maxLen int) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeTooLong,
		Message: fmt.Sprintf("%s must not exceed %d characters", field, maxLen),
		Meta:    map[string]any{"max": maxLen},
	}
}

// TooShort returns a [FieldError] for a string/slice below minLen length.
func TooShort(field string, minLen int) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeTooShort,
		Message: fmt.Sprintf("%s must be at least %d characters", field, minLen),
		Meta:    map[string]any{"min": minLen},
	}
}

// OutOfRange returns a [FieldError] for a numeric value outside [minVal, maxVal].
func OutOfRange(field string, minVal, maxVal any) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeOutOfRange,
		Message: fmt.Sprintf("%s must be between %v and %v", field, minVal, maxVal),
		Meta:    map[string]any{"min": minVal, "max": maxVal},
	}
}

// Immutable returns a [FieldError] for a field that cannot change after creation.
func Immutable(field string) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeImmutable,
		Message: field + " cannot be changed after creation",
	}
}

// TypeMismatch returns a [FieldError] for a wrong-type value.
func TypeMismatch(field, expected string) FieldError {
	return FieldError{
		Field:   field,
		Code:    CodeTypeMismatch,
		Message: fmt.Sprintf("%s must be of type %s", field, expected),
		Meta:    map[string]any{"expected_type": expected},
	}
}

// Entity returns a top-level [FieldError] not attributable to a specific field.
func Entity(code Code, message string) FieldError {
	return FieldError{Code: code, Message: message}
}
