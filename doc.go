// Package validation provides primitives for structured, field-level validation errors.
//
// # Motivation
//
// Most validation libraries either return a single error string or require heavy
// framework coupling. This package stays framework-agnostic: it gives you a
// machine-readable [Error] that carries every field failure at once, so API
// handlers can serialise all problems in one response and clients can
// switch on stable [Code] values without parsing message strings.
//
// # Usage
//
// Use [NewCollector] when validating multiple fields in one pass:
//
//	func ValidateUser(req CreateUserRequest) error {
//	    c := validation.NewCollector("invalid user")
//	    c.Check(req.Name != "", validation.Required("name"))
//	    c.Check(len(req.Name) <= 100, validation.TooLong("name", 100))
//	    c.Check(isEmail(req.Email), validation.Invalid("email", "must be a valid address"))
//	    c.Merge("address", validateAddress(req.Address))
//	    return c.Err()
//	}
//
// Use [As] to unwrap and inspect the structured error on the receiving side:
//
//	if ve, ok := validation.As(err); ok {
//	    // ve.Fields contains all FieldErrors
//	    // ve.Summary is the top-level human-readable message
//	}
//
// When you need the concrete [*Error] for field inspection without going
// through the error interface, use [Collector.Validation]:
//
//	if ve := c.Validation(); ve != nil {
//	    fe, ok := ve.First("email")
//	}
//
// # Concurrency
//
// [Error] and [Collector] are not safe for concurrent use. Build them within a
// single goroutine and pass the finished error across goroutine boundaries.
//
// # Caveats
//
// Never set [FieldError.Value] for sensitive fields such as passwords or tokens.
// Use [FieldError.WithValue] explicitly only when the value is safe to expose.
//
// # Nil safety
//
// All functions that accept error arguments ([As], [Is], [Collector.Merge]) treat
// nil as "no error" and return their zero/false values without panicking.
// This matches standard library conventions and allows patterns like:
//
//	c.Merge("address", mayReturnNil())
//
// # Zero values
//
// A [FieldError] with an empty Field represents an entity-level error — a
// validation failure not attributable to a single field (e.g. "user already
// exists"). Use the [Entity] builder for these cases rather than leaving
// Field empty by accident.
package validation
