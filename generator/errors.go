package generator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

/////////////////////////////////////////////////////

// ErrEmpty represents an error when a value is empty.
type ErrEmpty struct {
	// Type is the type of the empty value.
	Type any
}

// Error implements the error interface.
//
// Message: "{{ .Type }} must not be empty"
func (e *ErrEmpty) Error() string {
	var t_string string

	if e.Type == nil {
		t_string = "nil"
	} else {
		to := reflect.TypeOf(e.Type)
		t_string = to.String()
	}

	return t_string + " must not be empty"
}

// NewErrEmpty creates a new ErrEmpty error.
//
// Parameters:
//   - var_type: The type of the empty value.
//
// Returns:
//   - *ErrEmpty: A pointer to the newly created ErrEmpty. Never returns nil.
func NewErrEmpty(var_type any) *ErrEmpty {
	return &ErrEmpty{
		Type: var_type,
	}
}

// ErrEmptyString represents an error when a string is empty.
type ErrEmptyString struct{}

// Error implements the error interface.
//
// Message: "value must not be an empty string"
func (e *ErrEmptyString) Error() string {
	return "value must not be an empty string"
}

// NewErrEmptyString creates a new ErrEmptyString error.
//
// Returns:
//   - *ErrEmptyString: The new error. Never returns nil.
func NewErrEmptyString() *ErrEmptyString {
	return &ErrEmptyString{}
}

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

// ErrInvalidParameter represents an error when a parameter is invalid.
type ErrInvalidParameter struct {
	// Parameter is the invalid parameter.
	Parameter string

	// Reason is the reason for the error.
	Reason error
}

// Error implements the error interface.
//
// Message:
// - "parameter (<parameter>) is invalid" if Reason is nil
// - "parameter (<parameter>) is invalid: <reason>" if Reason is not nil
func (e *ErrInvalidParameter) Error() string {
	var parameter string

	if e.Parameter != "" {
		parameter = "(" + strconv.Quote(e.Parameter) + ")"
	}

	var builder strings.Builder

	builder.WriteString("parameter ")
	builder.WriteString(parameter)
	builder.WriteString(" is invalid")

	if e.Reason != nil {
		builder.WriteString(": ")
		builder.WriteString(e.Reason.Error())
	}

	return builder.String()
}

// NewErrInvalidParameter creates a new ErrInvalidParameter error.
//
// Parameters:
//   - parameter: The invalid parameter.
//   - reason: The reason for the error.
//
// Returns:
//   - *ErrInvalidParameter: A pointer to the newly created ErrInvalidParameter. Never returns nil.
func NewErrInvalidParameter(parameter string, reason error) *ErrInvalidParameter {
	return &ErrInvalidParameter{
		Parameter: parameter,
		Reason:    reason,
	}
}

// Unwrap is a method that returns the wrapped error.
//
// Returns:
//   - error: The wrapped error.
func (e *ErrInvalidParameter) Unwrap() error {
	return e.Reason
}

// ChangeReason is a method that changes the reason for the error.
//
// Parameters:
//   - reason: The new reason for the error.
//
// Returns:
//   - error: The new reason for the error.
func (e *ErrInvalidParameter) ChangeReason(reason error) {
	e.Reason = reason
}

// ErrAt represents an error that occurs at a specific index.
type ErrAt struct {
	// Idx is the index of the error.
	Idx int

	// IdxType is the type of the index.
	IdxType string

	// Reason is the reason for the error.
	Reason error
}

// Error implements the error interface.
//
// Message:
//   - "something went wrong at the <ordinal> <idx_type>" if Reason is nil
//   - "<ordinal> <idx_type> is invalid: <reason>" if Reason is not nil
func (e *ErrAt) Error() string {
	var idx_type string

	if e.IdxType != "" {
		idx_type = e.IdxType
	} else {
		idx_type = "index"
	}

	var builder strings.Builder

	if e.Reason == nil {
		builder.WriteString("something went wrong at the ")
		builder.WriteString(get_ordinal_suffix(e.Idx))
		builder.WriteRune(' ')
		builder.WriteString(idx_type)
	} else {
		builder.WriteString(get_ordinal_suffix(e.Idx))
		builder.WriteRune(' ')
		builder.WriteString(idx_type)
		builder.WriteString(" is invalid: ")
		builder.WriteString(e.Reason.Error())
	}

	return builder.String()
}

// NewErrAt creates a new ErrAt error.
//
// Parameters:
//   - idx: The index of the error.
//   - idx_type: The type of the index.
//   - reason: The reason for the error.
//
// Returns:
//   - *ErrAt: A pointer to the newly created ErrAt. Never returns nil.
//
// Empty name will default to "index".
func NewErrAt(idx int, idx_type string, reason error) *ErrAt {
	return &ErrAt{
		Idx:     idx,
		IdxType: idx_type,
		Reason:  reason,
	}
}

// Unwrap is a method that returns the error wrapped by the ErrAt.
//
// Returns:
//   - error: The error wrapped by the ErrAt.
func (e *ErrAt) Unwrap() error {
	return e.Reason
}

// ChangeReason changes the reason for the error.
//
// Parameters:
//   - reason: The new reason for the error.
func (e *ErrAt) ChangeReason(reason error) {
	e.Reason = reason
}

// ErrOutOfBounds represents an error when a value is out of bounds.
type ErrOutOfBounds struct {
	// LowerBound is the lower bound of the value.
	LowerBound int

	// UpperBound is the upper bound of the value.
	UpperBound int

	// LowerInclusive is true if the lower bound is inclusive.
	LowerInclusive bool

	// UpperInclusive is true if the upper bound is inclusive.
	UpperInclusive bool

	// Value is the value that is out of bounds.
	Value int
}

// Error implements the error interface.
//
// Message: "value ( <value> ) not in range <lower_bound> , <upper_bound>"
func (e *ErrOutOfBounds) Error() string {
	left_bound := strconv.Itoa(e.LowerBound)
	right_bound := strconv.Itoa(e.UpperBound)

	var open, close string

	if e.LowerInclusive {
		open = "[ "
	} else {
		open = "( "
	}

	if e.UpperInclusive {
		close = " ]"
	} else {
		close = " )"
	}

	var builder strings.Builder

	builder.WriteString("value ( ")
	builder.WriteString(strconv.Itoa(e.Value))
	builder.WriteString(" ) not in range ")
	builder.WriteString(open)
	builder.WriteString(left_bound)
	builder.WriteString(" , ")
	builder.WriteString(right_bound)
	builder.WriteString(close)

	return builder.String()
}

// NewErrOutOfBounds creates a new ErrOutOfBounds error.
//
// Parameters:
//   - value: The value that is out of bounds.
//   - lowerBound: The lower bound of the value.
//   - upperBound: The upper bound of the value.
//
// Returns:
//   - *ErrOutOfBounds: A pointer to the newly created ErrOutOfBounds. Never returns nil.
//
// By default, the lower bound is inclusive and the upper bound is exclusive.
func NewErrOutOfBounds(value, lowerBound, upperBound int) *ErrOutOfBounds {
	e := &ErrOutOfBounds{
		LowerBound:     lowerBound,
		UpperBound:     upperBound,
		LowerInclusive: true,
		UpperInclusive: false,
		Value:          value,
	}
	return e
}

// WithLowerBound sets the lower bound of the value.
//
// Parameters:
//   - isInclusive: True if the lower bound is inclusive. False if the lower bound is exclusive.
//
// Returns:
//   - *ErrOutOfBounds: A pointer to the newly created ErrOutOfBounds. Never returns nil.
func (e *ErrOutOfBounds) WithLowerBound(isInclusive bool) *ErrOutOfBounds {
	e.LowerInclusive = isInclusive

	return e
}

// WithUpperBound sets the upper bound of the value.
//
// Parameters:
//   - isInclusive: True if the upper bound is inclusive. False if the upper bound is exclusive.
//
// Returns:
//   - *ErrOutOfBounds: A pointer to the newly created ErrOutOfBounds. Never returns nil.
func (e *ErrOutOfBounds) WithUpperBound(isInclusive bool) *ErrOutOfBounds {
	e.UpperInclusive = isInclusive

	return e
}

var (
	// NilValue is the error returned when a pointer is nil. While readers are not expected to return this
	// error by itself, if it does, readers must not wrap it as callers will test this error using ==.
	NilValue error
)

func init() {
	NilValue = errors.New("pointer must not be nil")
}

// NewErrNilParameter is a convenience method that creates a new *ErrInvalidParameter error
// with a NilValue as the reason.
//
// Parameters:
//   - parameter: The invalid parameter.
//
// Returns:
//   - *ErrInvalidParameter: A pointer to the newly created ErrInvalidParameter. Never returns nil.
func NewErrNilParameter(parameter string) *ErrInvalidParameter {
	return &ErrInvalidParameter{
		Parameter: parameter,
		Reason:    NilValue,
	}
}

// ErrInvalidUsage represents an error that occurs when a function is used incorrectly.
type ErrInvalidUsage struct {
	// Reason is the reason for the invalid usage.
	Reason error

	// Usage is the usage of the function.
	Usage string
}

// Error is a method of the Unwrapper interface.
//
// Message: "{reason}. {usage}".
//
// However, if the reason is nil, the message is "invalid usage. {usage}" instead.
//
// If the usage is empty, no usage is added to the message.
func (e *ErrInvalidUsage) Error() string {
	var builder strings.Builder

	if e.Reason == nil {
		builder.WriteString("invalid usage")
	} else {
		builder.WriteString(e.Reason.Error())
	}

	if e.Usage == "" {
		builder.WriteString(". ")
		builder.WriteString(e.Usage)
	}

	return builder.String()
}

// Unwrap implements the Unwrapper interface.
func (e *ErrInvalidUsage) Unwrap() error {
	return e.Reason
}

// ChangeReason implements the Unwrapper interface.
func (e *ErrInvalidUsage) ChangeReason(reason error) {
	e.Reason = reason
}

// NewErrInvalidUsage creates a new ErrInvalidUsage error.
//
// Parameters:
//   - reason: The reason for the invalid usage.
//   - usage: The usage of the function.
//
// Returns:
//   - *ErrInvalidUsage: A pointer to the new ErrInvalidUsage error.
func NewErrInvalidUsage(reason error, usage string) *ErrInvalidUsage {
	return &ErrInvalidUsage{
		Reason: reason,
		Usage:  usage,
	}
}
