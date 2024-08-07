package generator

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	gcers "github.com/PlayerR9/go-commons/errors"
)

// GoExport is an enum that represents whether a variable is exported or not.
type GoExport int

const (
	// NotExported represents a variable that is not exported.
	NotExported GoExport = iota

	// Exported represents a variable that is exported.
	Exported

	// Either represents a variable that is either exported or not exported.
	Either
)

// FixVariableName acts in the same way as IsValidName but fixes the variable name if it is invalid.
//
// Parameters:
//   - variable_name: The variable name to check.
//   - keywords: The list of keywords to check against.
//   - exported: Whether the variable is exported or not.
//
// Returns:
//   - string: The fixed variable name.
//   - error: An error if the variable name is invalid.
func FixVariableName(variable_name string, keywords []string, exported GoExport) (string, error) {
	if variable_name == "" {
		return "", gcers.NewErrEmpty(variable_name)
	}

	switch exported {
	case NotExported:
		r, size := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return "", errors.New("invalid UTF-8 encoding")
		}

		if !unicode.IsLetter(r) {
			return "", errors.New("identifier must start with a letter")
		}

		ok := unicode.IsLower(r)
		if !ok {
			r = unicode.ToLower(r)
			variable_name = variable_name[size:]

			var builder strings.Builder

			builder.WriteRune(r)
			builder.WriteString(variable_name)

			variable_name = builder.String()
		}

		_, ok = slices.BinarySearch(go_reserved_keywords, variable_name)
		if ok {
			return "", fmt.Errorf("variable (%q) is a reserved keyword", variable_name)
		}

		return variable_name, nil
	case Exported:
		r, size := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return "", errors.New("invalid UTF-8 encoding")
		}

		if !unicode.IsLetter(r) {
			return "", errors.New("identifier must start with a letter")
		}

		ok := unicode.IsUpper(r)
		if !ok {
			r = unicode.ToUpper(r)
			variable_name = variable_name[size:]

			var builder strings.Builder

			builder.WriteRune(r)
			builder.WriteString(variable_name)

			variable_name = builder.String()
		}

		return variable_name, nil
	}

	ok := slices.Contains(keywords, variable_name)
	if ok {
		return "", fmt.Errorf("variable (%q) is already used", variable_name)
	}

	return variable_name, nil
}

// IsValidVariableName checks if the given variable name is valid.
//
// This function checks if the variable name is not empty and if it is not a
// Go reserved keyword. It also checks if the variable name is not in the list
// of keywords.
//
// Parameters:
//   - variable_name: The variable name to check.
//   - keywords: The list of keywords to check against.
//   - exported: Whether the variable is exported or not.
//
// Returns:
//   - error: An error if the variable name is invalid.
//
// If the variable is exported, the function checks if the variable name starts
// with an uppercase letter. If the variable is not exported, the function checks
// if the variable name starts with a lowercase letter. Any other case, the
// function does not perform any checks.
func IsValidVariableName(variable_name string, keywords []string, exported GoExport) error {
	if variable_name == "" {
		return gcers.NewErrEmpty(variable_name)
	}

	switch exported {
	case NotExported:
		r, _ := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsLower(r)
		if !ok {
			return errors.New("identifier must start with a lowercase letter")
		}

		_, ok = slices.BinarySearch(go_reserved_keywords, variable_name)
		if ok {
			return fmt.Errorf("identifier (%q) is a Go reserved keyword", variable_name)
		}
	case Exported:
		r, _ := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsUpper(r)
		if !ok {
			return errors.New("identifier must start with an uppercase letter")
		}
	}

	ok := slices.Contains(keywords, variable_name)
	if ok {
		err := errors.New("name is not allowed")
		return err
	}

	return nil
}

// GetPackages returns a list of packages from a list of strings.
//
// Parameters:
//   - packages: The list of strings to get the packages from.
//
// Returns:
//   - []string: The list of packages. Never returns nil.
func GetPackages(packages []string) []string {
	if len(packages) == 0 {
		return make([]string, 0)
	}

	var unique []string

	for _, elem := range packages {
		pos, ok := slices.BinarySearch(unique, elem)
		if !ok {
			unique = slices.Insert(unique, pos, elem)
		}
	}

	return unique
}

var (
	// zero_value_types is a list of types that have a default value of zero.
	zero_value_types []string

	// nillable_prefix is a list of prefixes that indicate a type is nillable.
	nillable_prefix []string
)

func init() {
	zero_value_types = []string{
		"byte",
		"complex64",
		"complex128",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"uintptr",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
	}

	nillable_prefix = []string{
		"[]",
		"map",
		"*",
		"chan",
		"func",
		"interface",
		"<-",
	}
}

// ZeroValueOf returns the zero value of a type.
//
// Parameters:
//   - type_name: The name of the type.
//   - custom: A map of custom types and their zero values.
//
// Returns:
//   - string: The zero value of the type.
func ZeroValueOf(type_name string, custom map[string]string) string {
	if type_name == "" {
		return ""
	}

	if custom != nil {
		zero, ok := custom[type_name]
		if ok {
			return zero
		}
	}

	for _, prefix := range nillable_prefix {
		if strings.HasPrefix(type_name, prefix) {
			return "nil"
		}
	}

	switch type_name {
	case "bool":
		return "false"
	case "error", "any":
		return "nil"
	case "float32", "float64":
		return "0.0"
	case "rune":
		return "'\\u0000'"
	case "string":
		return "\"\""
	}

	ok := slices.Contains(zero_value_types, type_name)
	if ok {
		return "0"
	}

	return "*new(" + type_name + ")"
}

// GetStringFnCall returns the string function call for the given element. It is
// just a wrapper around the reflect.GetStringOf function.
//
// Parameters:
//   - var_name: The name of the variable.
//   - type_name: The name of the type.
//   - custom: The custom strings to use. Empty values are ignored.
//
// Returns:
//   - string: The string function call.
//   - []string: The dependencies of the string function call.
func GetStringFnCall(var_name string, type_name string, custom map[string][]string) (string, []string) {
	if type_name == "" {
		return "\"nil\"", nil
	}

	if custom != nil {
		values, ok := custom[type_name]
		if ok && len(values) > 0 {
			return values[0], values[1:]
		}
	}

	var builder strings.Builder
	var dependencies []string

	switch type_name {
	case "bool":
		builder.WriteString("strconv.FormatBool(")
		builder.WriteString(var_name)
		builder.WriteString(")")

		dependencies = append(dependencies, "strconv")
	case "byte":
		builder.WriteString("string(")
		builder.WriteString(var_name)
		builder.WriteString(")")
	case "complex64":
		builder.WriteString("strconv.FormatComplex(complex128(")
		builder.WriteString(var_name)
		builder.WriteString("), 'f', -1, 64)")

		dependencies = append(dependencies, "strconv")
	case "complex128":
		builder.WriteString("strconv.FormatComplex(")
		builder.WriteString(var_name)
		builder.WriteString(", 'f', -1, 128)")

		dependencies = append(dependencies, "strconv")
	case "float32":
		builder.WriteString("strconv.FormatFloat(float64(")
		builder.WriteString(var_name)
		builder.WriteString("), 'f', -1, 32)")

		dependencies = append(dependencies, "strconv")
	case "float64":
		builder.WriteString("strconv.FormatFloat(")
		builder.WriteString(var_name)
		builder.WriteString(", 'f', -1, 64)")

		dependencies = append(dependencies, "strconv")
	case "int", "int8", "int16", "int32":
		builder.WriteString("strconv.FormatInt(int64(")
		builder.WriteString(var_name)
		builder.WriteString("), 10)")

		dependencies = append(dependencies, "strconv")
	case "int64":
		builder.WriteString("strconv.FormatInt(")
		builder.WriteString(var_name)
		builder.WriteString(", 10)")

		dependencies = append(dependencies, "strconv")
	case "rune":
		builder.WriteString("string(")
		builder.WriteString(var_name)
		builder.WriteString(")")
	case "string":
		builder.WriteString(var_name)
	case "uint", "uint8", "uint16", "uint32", "uintptr":
		builder.WriteString("strconv.FormatUint(uint64(")
		builder.WriteString(var_name)
		builder.WriteString("), 10)")

		dependencies = append(dependencies, "strconv")
	case "uint64":
		builder.WriteString("strconv.FormatUint(")
		builder.WriteString(var_name)
		builder.WriteString(", 10)")

		dependencies = append(dependencies, "strconv")
	case "error":
		builder.WriteString(var_name)
		builder.WriteString(".Error()")
	default:
		builder.WriteString("fmt.Sprintf(\"%v\", ")
		builder.WriteString(var_name)
		builder.WriteString(")")

		dependencies = append(dependencies, "fmt")
	}

	return builder.String(), dependencies
}

// InitLogger initializes the logger.
//
// Parameters:
//   - out: The output writer. Defaults to os.Stdout.
//   - name: The name of the logger. Defaults to "go-generator".
//
// Returns:
//   - *log.Logger: The initialized logger. Never returns nil.
func InitLogger(out io.Writer, name string) *log.Logger {
	if out == nil {
		out = os.Stdout
	}

	if name == "" {
		name = "go-generator"
	}

	var builder strings.Builder

	builder.WriteString("[")
	builder.WriteString(name)
	builder.WriteString("]: ")

	return log.New(out, builder.String(), log.Lshortfile)
}
