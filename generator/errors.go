package generator

import (
	"errors"
	"strconv"
	"strings"
)

/////////////////////////////////////////////////////

// ErrInvalidID represents an error when an identifier is invalid.
type ErrInvalidID struct {
	// ID is the invalid identifier.
	ID string

	// Reason is the reason why the identifier is invalid.
	Reason error
}

// Error implements the error interface.
//
// Message: "identifier <id> is invalid: <reason>"
func (e *ErrInvalidID) Error() string {
	q_id := strconv.Quote(e.ID)

	var reason string
	var builder strings.Builder

	if e.Reason != nil {
		re := e.Reason.Error()

		builder.WriteString(": ")
		builder.WriteString(re)

		reason = builder.String()
		builder.Reset()
	}

	builder.WriteString("identifier ")
	builder.WriteString(q_id)
	builder.WriteString(" is invalid")
	builder.WriteString(reason)

	str := builder.String()
	return str
}

// NewErrInvalidID creates a new ErrInvalidID error.
//
// Parameters:
//   - id: The invalid identifier.
//   - reason: The reason for the error.
//
// Returns:
//   - *ErrInvalidID: The new error.
func NewErrInvalidID(id string, reason error) *ErrInvalidID {
	e := &ErrInvalidID{
		ID:     id,
		Reason: reason,
	}

	return e
}

// ErrNotGeneric is an error type for when a type is not a generic.
type ErrNotGeneric struct {
	// Reason is the reason for the error.
	Reason error
}

// Error implements the error interface.
//
// Message: "not a generic type"
func (e *ErrNotGeneric) Error() string {
	if e.Reason == nil {
		return "not a generic type"
	}

	var builder strings.Builder

	builder.WriteString("not a generic type: ")
	builder.WriteString(e.Reason.Error())

	str := builder.String()

	return str
}

// NewErrNotGeneric creates a new ErrNotGeneric error.
//
// Parameters:
//   - reason: The reason for the error.
//
// Returns:
//   - *ErrNotGeneric: The new error.
func NewErrNotGeneric(reason error) *ErrNotGeneric {
	e := &ErrNotGeneric{
		Reason: reason,
	}

	return e
}

// IsErrNotGeneric checks if an error is of type ErrNotGeneric.
//
// Parameters:
//   - target: The error to check.
//
// Returns:
//   - bool: True if the error is of type ErrNotGeneric, false otherwise.
func IsErrNotGeneric(target error) bool {
	if target == nil {
		return false
	}

	var targetErr *ErrNotGeneric

	ok := errors.As(target, &targetErr)
	return ok
}
