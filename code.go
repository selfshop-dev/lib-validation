package validation // code.go

// Code is a stable, machine-readable string that API clients can switch on.
//
// Treat each Code as a versioned API contract: clients in production may have
// hardcoded switch statements, error mapping tables, or localisation keys
// that depend on these exact strings. Renaming a Code is a breaking change
// equivalent to removing a public function — it requires a major version bump.
//
// Adding new Codes is safe. Renaming or removing existing Codes is not.
type Code string

// Validation error codes. These values are part of the public API contract.
// Never rename or remove a code once it has been published — consumers depend
// on them for programmatic error handling and localisation.
const (
	CodeRequired     Code = "required"
	CodeConflict     Code = "conflict"
	CodeTooLong      Code = "too_long"
	CodeUnknown      Code = "unknown" // unrecognised config key (lib-config strict mode)
	CodeInvalid      Code = "invalid"
	CodeOutOfRange   Code = "out_of_range"
	CodeTooShort     Code = "too_short"
	CodeImmutable    Code = "immutable"
	CodeTypeMismatch Code = "type_mismatch"
)
