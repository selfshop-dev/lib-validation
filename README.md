# lib-validation

[![CI](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-validation/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-validation)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-validation)](https://goreportcard.com/report/github.com/selfshop-dev/lib-validation)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-validation)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-validation)](LICENSE)

Structured validation errors with machine-readable codes for Go. No external dependencies. A project by [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-validation
```

## Overview

`lib-validation` provides a single `Error` type that carries all field errors at once — an API handler can serialize every problem in a single response, and clients switch on stable `Code` values instead of parsing message strings.

```go
func ValidateUser(req CreateUserRequest) error {
    c := validation.NewCollector("invalid user")
    c.Check(req.Name != "", validation.Required("name"))
    c.Check(len(req.Name) <= 100, validation.TooLong("name", 100))
    c.Check(isEmail(req.Email), validation.Invalid("email", "must be a valid address"))
    c.Merge("address", validateAddress(req.Address))
    return c.Err()
}

// On the receiving side:
if ve, ok := validation.As(err); ok {
    for _, fe := range ve.Fields {
        fmt.Printf("%s: [%s] %s\n", fe.Field, fe.Code, fe.Message)
    }
}
```

Key properties:

- **No external dependencies** — standard library only.
- **Stable codes** — `Code` is part of the public API contract; clients may depend on their string values.
- **Nested validation** — `Merge` automatically prefixes field names using dot-notation.
- **Safe by default** — field values are never included in errors automatically; an explicit `WithValue` call is required.

### Quick Start

```go
import "github.com/selfshop-dev/lib-validation"

c := validation.NewCollector("invalid user")
c.Check(req.Name != "", validation.Required("name"))
c.Check(len(req.Name) <= 50, validation.TooLong("name", 50))

if err := c.Err(); err != nil {
    return err
}
```

## Error Codes

Codes are a stable part of the public API. Renaming or removing a code is a breaking change requiring a major release. Adding new codes is safe.

| Code | Constant | Description |
|---|---|---|
| `required` | `CodeRequired` | Field is missing or empty |
| `invalid` | `CodeInvalid` | Value is present but invalid |
| `too_long` | `CodeTooLong` | String or slice exceeds maximum length |
| `too_short` | `CodeTooShort` | String or slice is shorter than minimum length |
| `out_of_range` | `CodeOutOfRange` | Numeric value is outside the allowed range |
| `conflict` | `CodeConflict` | Value conflicts with existing state |
| `immutable` | `CodeImmutable` | Field cannot be changed after creation |
| `type_mismatch` | `CodeTypeMismatch` | Value has the wrong type |
| `unknown` | `CodeUnknown` | Unrecognized key (used in lib-config) |

## Builders

Ready-made constructors cover the most common scenarios and automatically populate `Code`, `Message`, and `Meta`.

```go
validation.Required("email")                       // field is required
validation.Invalid("email", "not a valid address") // invalid value
validation.TooLong("username", 50)                 // Meta: {"max": 50}
validation.TooShort("password", 8)                 // Meta: {"min": 8}
validation.OutOfRange("age", 18, 120)              // Meta: {"min": 18, "max": 120}
validation.Conflict("email", "already taken")
validation.Immutable("user_id")
validation.TypeMismatch("count", "integer")        // Meta: {"expected_type": "integer"}
validation.Unknown("extra_field")

validation.Entity(validation.CodeConflict, "duplicate entry") // entity-level error, no field
```

## Collector

`Collector` accumulates errors during a field-by-field pass and returns `*Error` (or `nil`) at the end. All methods return `*Collector` and support method chaining.

```go
c := validation.NewCollector("invalid user")

// Add an error if the condition is false
c.Check(req.Name != "", validation.Required("name"))

// Add an error if the condition is true — useful when the condition describes a violation
c.Fail(len(req.Name) < minLen, validation.TooShort("name", minLen))
c.Fail(len(req.Name) > maxLen, validation.TooLong("name", maxLen))

// Add unconditionally
c.Add(validation.Required("email"))

// Nested validation with automatic prefix
c.Merge("shipping_address", validateAddress(req.Address))
// field "city" inside → "shipping_address.city"

// Get the result
err := c.Err()        // error or nil
ve  := c.Validation() // *Error or nil — for field inspection
```

## Error Inspection

Once you have a `*Error`, several methods are available for field lookup.

```go
ve, ok := validation.As(err) // extract *Error from the error chain
ve.Fields                    // all FieldErrors

ve.First("email")                                     // (FieldError, bool) — first error for the field
ve.FirstWithCode("password", validation.CodeTooShort) // by field and code

ve.FieldsFor("email") // all errors for a field
ve.Codes()            // unique codes across all fields

validation.Is(err) // whether *Error is present in the chain (without field inspection)
```

## FieldError

`FieldError` describes a single validation error. The `Field` uses dot-notation and is compatible with both config keys (`database.host`) and JSON body paths (`user.address.zip_code`). An empty `Field` denotes an entity-level error.

```go
fe := validation.Invalid("status", "unrecognised value")

// Attach a safe value for debugging — only for non-sensitive fields
fe = fe.WithValue("PENDING_APPROVAL")

// Add metadata
fe = fe.WithMetaPair("allowed", []string{"active", "inactive"})
```

`WithValue` and `WithMetaPair` return a copy — the original `FieldError` is not mutated.

## Nested Validation

`Merge` lets you call separate validation functions for nested structs and collect their errors into a single result with correct field paths.

```go
c := validation.NewCollector("invalid order")
c.Merge("shipping_address", validateAddress(req.ShippingAddress))
c.Merge("billing_address", validateAddress(req.BillingAddress))

// Resulting paths: "shipping_address.city", "billing_address.zip_code", etc.
```

Nesting depth is unlimited — each level simply prepends its own prefix.

## License

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)