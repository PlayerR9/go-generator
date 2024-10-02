package generator

// ErrorCode represents an error code.
type ErrorCode int

const (
	// BadID occurs when an identifier is invalid.
	BadID ErrorCode = iota

	// BadType occurs when the type of the generic is invalid.
	BadGeneric
)

// Error implements the errors.ErrorCoder interface.
func (e ErrorCode) Int() int {
	return int(e)
}
