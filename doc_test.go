package validation_test

import (
	"errors"
	"fmt"

	validation "github.com/selfshop-dev/lib-validation"
)

// ExampleNewCollector demonstrates a typical multi-field validation pass.
func ExampleNewCollector() {
	type CreateUserRequest struct {
		Name  string
		Email string
		Age   int
	}

	// isEmail is a minimal check: must contain exactly one '@' not at the edges.
	isEmail := func(s string) bool {
		for i, ch := range s {
			if ch == '@' && i > 0 && i < len(s)-1 {
				return true
			}
		}
		return false
	}

	validate := func(req CreateUserRequest) error {
		c := validation.NewCollector("invalid user")
		c.Check(req.Name != "", validation.Required("name"))
		c.Check(len(req.Name) <= 50, validation.TooLong("name", 50))
		c.Check(req.Email != "", validation.Required("email"))
		c.Check(isEmail(req.Email), validation.Invalid("email", "must be a valid address"))
		c.Check(req.Age >= 18, validation.OutOfRange("age", 18, 120))
		return c.Err()
	}

	err := validate(CreateUserRequest{Name: "", Email: "bad", Age: 10})
	if ve, ok := validation.As(err); ok {
		fmt.Println(ve.Summary)
		for _, fe := range ve.Fields {
			fmt.Printf("  %s: [%s] %s\n", fe.Field, fe.Code, fe.Message)
		}
	}

	// Output:
	// invalid user
	//   name: [required] name is required
	//   email: [invalid] must be a valid address
	//   age: [out_of_range] age must be between 18 and 120
}

// ExampleCollector_Merge shows nested validation with dot-notation field prefixing.
func ExampleCollector_Merge() {
	validateAddress := func(city, zip string) error {
		c := validation.NewCollector("invalid address")
		c.Check(city != "", validation.Required("city"))
		c.Check(len(zip) == 5, validation.Invalid("zip_code", "must be 5 digits"))
		return c.Err()
	}

	c := validation.NewCollector("invalid order")
	c.Merge("shipping_address", validateAddress("", "123"))

	if ve := c.Validation(); ve != nil {
		for _, fe := range ve.Fields {
			fmt.Printf("%s: %s\n", fe.Field, fe.Message)
		}
	}

	// Output:
	// shipping_address.city: city is required
	// shipping_address.zip_code: must be 5 digits
}

func ExampleCollector_Merge_nested() {
	validateAddress := func(city string) error {
		c := validation.NewCollector("invalid address")
		c.Check(city != "", validation.Required("city"))
		return c.Err()
	}

	c := validation.NewCollector("invalid order")
	c.Merge("order.shipping", validateAddress(""))

	if ve := c.Validation(); ve != nil {
		fmt.Println(ve.Fields[0].Field)
	}

	// Output:
	// order.shipping.city
}

// ExampleAs demonstrates unwrapping a *Error from an error chain.
func ExampleAs() {
	c := validation.NewCollector("invalid product")
	c.Add(validation.Required("sku"), validation.TooLong("description", 500))

	err := fmt.Errorf("create product: %w", c.Err())

	if ve, ok := validation.As(err); ok {
		fmt.Println("summary:", ve.Summary)
		fmt.Println("field count:", len(ve.Fields))
	}

	// Output:
	// summary: invalid product
	// field count: 2
}

// ExampleError_First shows retrieving the first error for a specific field.
func ExampleError_First() {
	c := validation.NewCollector("invalid input")
	c.Add(
		validation.TooShort("password", 8),
		validation.TooLong("password", 64),
	)

	if ve := c.Validation(); ve != nil {
		if fe, ok := ve.First("password"); ok {
			fmt.Printf("[%s] %s\n", fe.Code, fe.Message)
		}
	}

	// Output:
	// [too_short] password must be at least 8 characters
}

func ExampleError_FirstWithCode() {
	c := validation.NewCollector("invalid input")
	c.Add(
		validation.TooShort("password", 8),
		validation.TooLong("password", 64),
	)

	if ve := c.Validation(); ve != nil {
		if fe, ok := ve.FirstWithCode("password", validation.CodeTooShort); ok {
			fmt.Println(fe.Meta["min"])
		}
	}

	// Output:
	// 8
}

// ExampleError_Codes shows retrieving the unique set of error codes.
func ExampleError_Codes() {
	c := validation.NewCollector("invalid input")
	c.Add(
		validation.Required("email"),
		validation.Required("name"),
		validation.TooLong("bio", 200),
	)

	if ve := c.Validation(); ve != nil {
		for _, code := range ve.Codes() {
			fmt.Println(code)
		}
	}

	// Output:
	// required
	// too_long
}

// ExampleFieldError_WithValue demonstrates attaching a safe field value for debugging.
func ExampleFieldError_WithValue() {
	fe := validation.Invalid("status", "unrecognised status value").WithValue("PENDING_APPROVAL")
	fmt.Println(fe.Value)

	// Output:
	// PENDING_APPROVAL
}

// ExampleIs demonstrates checking whether an error chain contains a *Error.
func ExampleIs() {
	c := validation.NewCollector("invalid request")
	c.Add(validation.Required("id"))

	wrapped := fmt.Errorf("handler: %w", c.Err())

	fmt.Println(validation.Is(wrapped))
	fmt.Println(validation.Is(errors.New("plain error")))

	// Output:
	// true
	// false
}
