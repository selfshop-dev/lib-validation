package validation

import (
	"maps"
	"strings"
)

// FieldError describes a validation failure on a single field.
// Field uses dot-notation to match both config keys and JSON body paths:
//
//	"database.host"         ← lib-config key
//	"user.address.zip_code" ← nested JSON field
//	""                      ← entity-level error (no specific field)
//
// Field ordering is chosen for optimal struct alignment (map pointer first).
type FieldError struct {
	Meta    map[string]any `json:"meta,omitempty"`
	Field   string         `json:"field,omitempty"`
	Code    Code           `json:"code"`
	Message string         `json:"message"`

	// Value is intentionally never set by the builder functions (Required, TooLong,
	// etc.) because validation errors often involve sensitive input — passwords,
	// tokens, PII. Callers must opt in via WithValue for fields where exposure
	// is safe and useful (e.g. an unrecognised enum value, a malformed date string).
	//
	// Never call WithValue on fields that may contain credentials or personal data.
	Value string `json:"value,omitempty"` // omit for sensitive fields
}

func (fe FieldError) Error() string {
	// Pre-calculate capacity: "[" + code + "] " + field? + message
	var b strings.Builder
	b.Grow(2 + len(fe.Code) +
		2 + len(fe.Field) +
		2 + len(fe.Message))
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
	return b.String()
}

// WithValue returns a copy with Value set. Opt-in — never set on passwords/tokens.
func (fe FieldError) WithValue(v string) FieldError {
	fe.Value = v
	return fe
}

// WithMetaPair returns a copy with an additional metadata key set.
// Always allocates a new map so the original FieldError is never mutated.
func (fe FieldError) WithMetaPair(key string, val any) FieldError {
	dst := make(map[string]any, len(fe.Meta)+1)
	maps.Copy(dst, fe.Meta)
	dst[key] = val
	fe.Meta = dst
	return fe
}
