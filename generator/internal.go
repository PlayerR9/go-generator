package generator

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

/////////////////////////////////////////////////////

// get_ordinal_suffix returns the ordinal suffix for a given integer.
//
// Parameters:
//   - number: The integer for which to get the ordinal suffix. Negative
//     numbers are treated as positive.
//
// Returns:
//   - string: The ordinal suffix for the number.
//
// Example:
//   - get_ordinal_suffix(1) returns "1st"
//   - get_ordinal_suffix(2) returns "2nd"
func get_ordinal_suffix(number int) string {
	var builder strings.Builder

	builder.WriteString(strconv.Itoa(number))

	if number < 0 {
		number = -number
	}

	lastTwoDigits := number % 100
	lastDigit := lastTwoDigits % 10

	if lastTwoDigits >= 11 && lastTwoDigits <= 13 {
		builder.WriteString("th")
	} else {
		switch lastDigit {
		case 1:
			builder.WriteString("st")
		case 2:
			builder.WriteString("nd")
		case 3:
			builder.WriteString("rd")
		default:
			builder.WriteString("th")
		}
	}

	return builder.String()
}

// try_insert is a helper function that inserts an element into a slice only
// if the element is not already in the slice.
//
// Parameters:
//   - slc: The slice to insert into.
//   - e: The element to insert.
//
// Returns:
//   - []T: The slice with the inserted element.
//
// This function only works if the slice is sorted.
func try_insert[T cmp.Ordered](slc []T, e T) []T {
	pos, ok := slices.BinarySearch(slc, e)
	if ok {
		return slc
	}

	slc = slices.Insert(slc, pos, e)

	return slc
}
