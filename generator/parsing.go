package generator

import (
	"errors"
	"flag"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	gcers "github.com/PlayerR9/go-commons/errors"
	gcint "github.com/PlayerR9/go-commons/ints"
	gcslc "github.com/PlayerR9/go-commons/slices"
)

// PrintFlags prints the default values of the flags.
//
// It is useful for debugging and for error messages.
func PrintFlags() {
	flag.PrintDefaults()
}

// ParseFlags parses the command line flags.
func ParseFlags() {
	flag.Parse()
}

// AlignGenerics aligns the generics in the given values.
//
// Parameters:
//   - g: The *GenericsSignVal to align.
//   - values: The values to align.
//
// Returns:
//   - error: An error if occurred.
func AlignGenerics(g *GenericsSignVal, values ...flag.Value) error {
	var top int

	for i := 0; i < len(values); i++ {
		if values[i] != nil {
			values[top] = values[i]
			top++
		}
	}

	values = values[:top]

	if g != nil && len(values) == 0 {
		return gcers.NewErrInvalidUsage(
			errors.New("not specified any values that have generics, yet *GenericsSignVal is specified"),
			"Make sure to call a flag that sets the *GenericsSignVal such as go-generator.NewTypeListFlag()",
		)
	}

	var all_generics []rune

	for _, value := range values {
		switch value := value.(type) {
		case interface{ Generics() []rune }:
			for _, key := range value.Generics() {
				all_generics = gcslc.TryInsert(all_generics, key)
			}
		}
	}

	if len(all_generics) > 0 && g == nil {
		return gcers.NewErrInvalidUsage(
			errors.New("specified at least one value that has generics but not specified the *GenericsSignVal"),
			"Make sure to call a flag that sets the *GenericsSignVal such as go-generator.NewTypeListFlag()",
		)
	}

	for _, generic_id := range all_generics {
		pos, ok := slices.BinarySearch(g.letters, generic_id)
		if ok {
			continue
		}

		g.letters = slices.Insert(g.letters, pos, generic_id)
		g.types = slices.Insert(g.types, pos, "any")
	}

	return nil
}

// MakeTypeSign creates a type signature from a type name and a suffix.
//
// It also adds the generic signature if it exists.
//
// Parameters:
//   - type_name: The name of the type.
//   - suffix: The suffix of the type.
//
// Returns:
//   - string: The type signature.
//   - error: An error if the type signature cannot be created. (i.e., the type name is empty)
func MakeTypeSign(g *GenericsSignVal, type_name string, suffix string) (string, error) {
	if type_name == "" {
		return "", gcers.NewErrInvalidParameter("type_name", gcers.NewErrEmpty(type_name))
	}

	var builder strings.Builder

	builder.WriteString(type_name)
	builder.WriteString(suffix)

	if g == nil {
		return builder.String(), nil
	}

	if len(g.letters) > 0 {
		builder.WriteString(g.Signature())
	}

	return builder.String(), nil
}

var (
	// go_reserved_keywords is a list of Go reserved keywords.
	go_reserved_keywords []string
)

func init() {
	keys := []string{
		"break", "case", "chan", "const", "continue", "default", "defer", "else",
		"fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
		"map", "package", "range", "return", "select", "struct", "switch", "type",
		"var",
	}

	for _, key := range keys {
		pos, _ := slices.BinarySearch(go_reserved_keywords, key)
		// dbg.AssertOk(!ok, "slices.BinarySearch(GoReservedKeywords, %q)", key)

		go_reserved_keywords = slices.Insert(go_reserved_keywords, pos, key)
	}
}

// is_generics_id checks if the input string is a valid single upper case letter and returns it as a rune.
//
// Parameters:
//   - id: The id to check.
//
// Returns:
//   - rune: The valid single upper case letter.
//   - error: An error of type *ErrInvalidID if the input string is not a valid identifier.
func is_generics_id(id string) (rune, error) {
	if id == "" {
		return '\000', gcers.NewErrEmpty(id)
	}

	size := utf8.RuneCountInString(id)
	if size > 1 {
		return '\000', errors.New("value must be a single character")
	}

	letter, _ := utf8.DecodeRuneInString(id)
	if letter == utf8.RuneError {
		return '\000', errors.New("value is not a valid unicode character")
	}

	ok := unicode.IsUpper(letter)
	if !ok {
		return '\000', errors.New("value must be an upper case letter")
	}

	return letter, nil
}

// parse_generics parses a string representing a list of generic types enclosed in square brackets.
//
// Parameters:
//   - str: The string to parse.
//
// Returns:
//   - []rune: An array of runes representing the parsed generic types.
//   - error: An error if the parsing fails.
//
// Errors:
//   - *ErrNotGeneric: The string is not a valid list of generic types.
//   - error: An error if the string is a possibly valid list of generic types but fails to parse.
func parse_generics(str string) ([]rune, error) {
	if str == "" {
		return nil, NewErrNotGeneric(gcers.NewErrEmpty(str))
	}

	var letters []rune

	ok := strings.HasSuffix(str, "]")
	if ok {
		idx := strings.Index(str, "[")
		if idx == -1 {
			err := errors.New("missing opening square bracket")
			return nil, err
		}

		generic := str[idx+1 : len(str)-1]
		if generic == "" {
			err := errors.New("empty generic type")
			return nil, err
		}

		fields := strings.Split(generic, ",")

		for i, field := range fields {
			letter, err := is_generics_id(field)
			if err != nil {
				return nil, gcint.NewErrAt(i+1, "field", err)
			}

			letters = append(letters, letter)
		}
	} else {
		letter, err := is_generics_id(str)
		if err != nil {
			err := NewErrNotGeneric(err)
			return nil, err
		}

		letters = append(letters, letter)
	}

	return letters, nil
}

// parse_generics_value is a helper function that is used to parse the generics
// values.
//
// Parameters:
//   - field: The field to parse.
//
// Returns:
//   - rune: The letter of the generic.
//   - string: The type of the generic.
//   - error: An error if the parsing fails.
//
// Errors:
//   - *ErrInvalidID: If the id is invalid.
//   - error: If the parsing fails.
//
// Assertions:
//   - field != ""
func parse_generics_value(field string) (rune, string, error) {
	// dbg.Assert(field != "", "field must not be an empty string")

	sub_fields := strings.Split(field, "/")

	if len(sub_fields) == 1 {
		return '\000', "", errors.New("missing type of generic")
	} else if len(sub_fields) > 2 {
		return '\000', "", errors.New("too many fields")
	}

	left := sub_fields[0]

	letter, err := is_generics_id(left)
	if err != nil {
		return '\000', "", NewErrInvalidID(left, err)
	}

	right := sub_fields[1]

	return letter, right, nil
}
