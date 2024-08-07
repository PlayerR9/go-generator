package generator

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

/////////////////////////////////////////////////////

var (
	// output_loc_flag is a flag that specifies the location of the output file.
	output_loc_flag *string

	// is_output_loc_required_flag is a flag that specifies whether the output location is required or not.
	is_output_loc_required_flag bool

	// struct_fields_flag is a pointer to the fields_flag flag.
	struct_fields_flag *struct_fields_val

	// generics_sig_flag is a pointer to the generics_flag flag.
	generics_sig_flag *generics_sign_val

	// type_list_flag is a pointer to the type_list_flag flag.
	type_list_flag *type_list_val
)

// set_output_flag sets the flag that specifies the location of the output file.
//
// Parameters:
//   - def_value: The default value of the output_flag flag.
//   - required: Whether the flag is required or not.
//
// Here are all the possible valid calls to this function:
//
//	set_output_flag("", false) <-> set_output_flag("[no location]", false)
//	set_output_flag("path/to/file.go", false)
//	set_output_flag("", true) <-> set_output_flag("path/to/file.go", true)
//
// However, the def_value parameter does not specify the actual default location of the output file.
// Instead, it is merely used in the "usage" portion of the flag specification in order to give the user
// more information about the location of the output file. Thus, if no output flag is set, the actual
// default location of the flag is an empty string.
//
// Documentation:
//
// **Flag: Output File**
//
// This optional flag is used to specify the output file. If not specified, the output will be written to
// standard output, that is, the file "<type_name>_treenode.go" in the root of the current directory.
func set_output_flag(def_value string, required bool) {
	var usage string

	if required {
		var builder strings.Builder

		builder.WriteString("The location of the output file. ")
		builder.WriteString("It must be set and it must specify a .go file.")

		usage = builder.String()
	} else {
		var def_loc string

		if def_value == "" {
			def_loc = "\"[no location]\""
		} else {
			def_loc = strconv.Quote(def_value)
		}

		var builder strings.Builder

		builder.WriteString("The location of the output file. ")

		builder.WriteString("If set, it must specify a .go file. ")
		builder.WriteString("On the other hand, if not set, the default location of ")
		builder.WriteString(def_loc)
		builder.WriteString(" will be used instead.")

		usage = builder.String()
	}

	output_loc_flag = flag.String("o", "", usage)
	is_output_loc_required_flag = required
}

// get_output_loc gets the location of the output file.
//
// Returns:
//   - string: The location of the output file.
//   - error: An error of type *common.ErrInvalidUsage if the output location was not defined
//     prior to calling this function.
func get_output_loc() (string, error) {
	if output_loc_flag == nil {
		return "", NewErrInvalidUsage(
			errors.New("output location was not defined"),
			"Please call the go_generator.SetOutputFlag() function before calling this function.",
		)
	}

	return *output_loc_flag, nil
}

// struct_fields_vaò is a struct that represents the fields value.
type struct_fields_val struct {
	// fields is a map of the fields and their types.
	fields *ordered_map[string, string]

	// generics is a map of the generics and their types.
	generics *ordered_map[rune, string]

	// is_required is a flag that specifies whether the fields value is required or not.
	is_required bool

	// count is the number of fields expected. -1 for unlimited number of fields.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	"<value1> <type1>, <value2> <type2>, ..."
func (s struct_fields_val) String() string {
	if s.fields.size() == 0 {
		return ""
	}

	var values []string
	var builder strings.Builder

	iter := s.fields.iterator()
	// dbg.Assert(iter != nil, "entry iterator is nil")

	for {
		entry, err := iter.Consume()
		if err != nil {
			break
		}

		builder.WriteString(entry.Key)
		builder.WriteRune(' ')
		builder.WriteString(entry.Value)

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *struct_fields_val) Set(value string) error {
	if value == "" && s.is_required {
		return errors.New("value must be set")
	}

	fields := strings.Split(value, ",")

	s.fields = new_ordered_map[string, string]()

	for i, field := range fields {
		if field == "" {
			continue
		}

		sub_fields := strings.Split(field, "/")

		if len(sub_fields) == 1 {
			reason := errors.New("missing type")
			err := NewErrAt(i+1, "field", reason)
			return err
		} else if len(sub_fields) > 2 {
			reason := errors.New("too many fields")
			err := NewErrAt(i+1, "field", reason)
			return err
		}

		ok := s.fields.add(sub_fields[0], sub_fields[1], false)
		if !ok {
			return fmt.Errorf("field %q already exists", sub_fields[0])
		}
	}

	size := s.fields.size()

	if s.count != -1 && size != s.count {
		return fmt.Errorf("wrong number of fields: expected %d, got %d", s.count, size)
	}

	s.generics = new_ordered_map[rune, string]()

	for _, field_type := range fields {
		chars, err := parse_generics(field_type)
		ok := IsErrNotGeneric(err)

		if ok {
			continue
		} else if err != nil {
			return fmt.Errorf("syntax error for type %q: %w", field_type, err)
		}

		for _, char := range chars {
			_ = s.generics.add(char, "", false)
			// dbg.AssertOk(ok, "s.generics.Add(%s, %q, false)", strconv.QuoteRune(char), "")
		}
	}

	return nil
}

// set_struct_fields_flag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Any negative number will be interpreted as unlimited number of fields. Also, the value 0 will not set the flag.
//
// Documentation:
//
// **Flag: Fields**
//
// The "fields" flag is used to specify the fields that the tree node contains. Because it doesn't make
// a lot of sense to have a tree node without fields, this flag must be set.
//
// Its argument is specified as a list of key-value pairs where each pair is separated by a comma (",") and
// a slash ("/") is used to separate the key and the value.
//
// The key indicates the name of the field while the value indicates the type of the field.
//
// For instance, running the following command:
//
//	//go:generate treenode -type="TreeNode" -fields=a/int,b/int,name/string
//
// will generate a tree node with the following fields:
//
//	type TreeNode struct {
//		// Node pointers.
//
//		a int
//		b int
//		name string
//	}
//
// It is important to note that spaces are not allowed.
//
// Also, it is possible to specify generics by following the value with the generics between square brackets;
// like so: "a/MyType[T,C]"
func set_struct_fields_flag(flag_name string, is_required bool, count int, brief string) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	struct_fields_flag = &struct_fields_val{
		fields:      new_ordered_map[string, string](),
		generics:    new_ordered_map[rune, string](),
		is_required: is_required,
		count:       count,
	}

	var usage strings.Builder

	usage.WriteString(brief)

	if is_required {
		if count == -1 {
			usage.WriteString("It must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It must be set with exactly %d fields.", count))
		}
	} else {
		if count == -1 {
			usage.WriteString("It is optional but, if set, it must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It is optional but, if set, it must be set with exactly %d fields.", count))
		}
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(struct_fields_flag, flag_name, usage.String())
}

// Fields returns the fields of the struct.
//
// Returns:
//   - map[string]string: A map of field names and their types. Never returns nil.
func (s struct_fields_val) Fields() map[string]string {
	return s.fields.Map()
}

// generics_sign_val is a struct that contains the values of the generics.
type generics_sign_val struct {
	// letters is a slice that contains the letters of the generics.
	letters []rune

	// types is a slice that contains the types of the generics.
	types []string

	// is_required is a flag that specifies whether the generics value is required or not.
	is_required bool

	// count is a flag that specifies the number of generics.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	[letter1 type1, letter2 type2, ...]
func (s generics_sign_val) String() string {
	if len(s.letters) == 0 {
		return ""
	}

	var values []string
	var builder strings.Builder

	for i, letter := range s.letters {
		builder.WriteRune(letter)
		builder.WriteRune(' ')
		builder.WriteString(s.types[i])

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")

	builder.WriteRune('[')
	builder.WriteString(joined_str)
	builder.WriteRune(']')

	str := builder.String()
	return str
}

// Set implements the flag.Value interface.
func (s *generics_sign_val) Set(value string) error {
	if value == "" {
		return nil
	}

	fields := strings.Split(value, ",")

	for i, field := range fields {
		if field == "" {
			continue
		}

		letter, g_type, err := parse_generics_value(field)
		if err != nil {
			return NewErrAt(i+1, "field", err)
		}

		err = s.add(letter, g_type)
		if err != nil {
			return NewErrAt(i+1, "field", err)
		}
	}

	if s.count != -1 && len(s.letters) != s.count {
		return fmt.Errorf("invalid number of generics: expected %d, got %d", s.count, len(s.letters))
	}

	return nil
}

// set_generics_sign_flag sets the flag that specifies the generics to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of generics. If -1, no upper bound is set, 0 means no generics.
//
// Documentation:
//
// **Flag: Generics**
//
// This optional flag is used to specify the type(s) of the generics. However, this only applies if at least one
// generic type is specified in the fields flag. If none, then this flag is ignored.
//
// As an edge case, if this flag is not specified but the fields flag contains generics, then
// all generics are set to the default value of "any".
//
// As with the fields flag, its argument is specified as a list of key-value pairs where each pair is separated
// by a comma (",") and a slash ("/") is used to separate the key and the value. The key indicates the name of
// the generic and the value indicates the type of the generic.
//
// For instance, running the following command:
//
//	//go:generate treenode -type="TreeNode" -fields=a/MyType[T],b/MyType[C] -g=T/any,C/int
//
// will generate a tree node with the following fields:
//
//	type TreeNode[T any, C int] struct {
//		// Node pointers.
//
//		a T
//		b C
//	}
func set_generics_sign_flag(flag_name string, is_required bool, count int) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	generics_sig_flag = &generics_sign_val{
		letters:     make([]rune, 0),
		types:       make([]string, 0),
		is_required: is_required,
		count:       count,
	}

	var usage strings.Builder

	usage.WriteString("The signature of generics.")

	if is_required {
		usage.WriteString("It must be set.")
	} else {
		usage.WriteString("It is optional.")
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(generics_sig_flag, flag_name, usage.String())
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

// add is a helper function that is used to add a generic to the GenericsValue.
//
// Parameters:
//   - letter: The letter of the generic.
//   - g_type: The type of the generic.
//
// Errors:
//   - error: If the parsing fails.
//
// Assertions:
//   - letter is an upper case letter.
//   - g_type != ""
func (gv *generics_sign_val) add(letter rune, g_type string) error {
	// dbg.AssertParam("letter", unicode.IsUpper(letter), errors.New("letter must be an upper case letter"))
	// dbg.AssertParam("g_type", g_type != "", errors.New("type must be set"))

	pos, ok := slices.BinarySearch(gv.letters, letter)
	if !ok {
		gv.letters = slices.Insert(gv.letters, pos, letter)
		gv.types = slices.Insert(gv.types, pos, g_type)

		return nil
	}

	if gv.types[pos] != g_type {
		err := fmt.Errorf("duplicate definition for generic %q: %s and %s", string(letter), gv.types[pos], g_type)
		return err
	}

	return nil
}

// Signature returns the signature of the generics.
//
// Format:
//
//	[T1, T2, T3]
//
// Returns:
//   - string: The list of generics.
func (gv generics_sign_val) Signature() string {
	if len(gv.letters) == 0 {
		return ""
	}

	values := make([]string, 0, len(gv.letters))

	for _, letter := range gv.letters {
		str := string(letter)
		values = append(values, str)
	}

	joined_str := strings.Join(values, ", ")

	var builder strings.Builder

	builder.WriteRune('[')
	builder.WriteString(joined_str)
	builder.WriteRune(']')

	str := builder.String()

	return str
}

// type_list_val is a struct that represents a list of types.
type type_list_val struct {
	// fields is a list of types.
	types []string

	// generics is a map of the generics and their types.
	generics *ordered_map[rune, string]

	// is_required is a flag that specifies whether the fields value is required or not.
	is_required bool

	// count is the number of fields expected.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	"<type1>, <type2>, ..."
func (s type_list_val) String() string {
	if len(s.types) == 0 {
		return ""
	}

	joined_str := strings.Join(s.types, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *type_list_val) Set(value string) error {
	if value == "" && s.is_required {
		return errors.New("value must be set")
	}

	parsed := strings.Split(value, ",")

	var top int

	for i := 0; i < len(parsed); i++ {
		if parsed[i] != "" {
			parsed[top] = parsed[i]
			top++
		}
	}

	parsed = parsed[:top]

	if s.count != -1 && len(parsed) != s.count {
		return fmt.Errorf("wrong number of types: expected %d, got %d", s.count, len(parsed))
	}

	if s.count != -1 && len(parsed) != s.count {
		return fmt.Errorf("wrong number of fields: expected %d, got %d", s.count, len(parsed))
	}

	s.types = parsed

	// Find generics

	s.generics = new_ordered_map[rune, string]()

	for _, field_type := range s.types {
		chars, err := parse_generics(field_type)
		ok := IsErrNotGeneric(err)

		if ok {
			continue
		} else if err != nil {
			return fmt.Errorf("syntax error for type %q: %w", field_type, err)
		}

		for _, char := range chars {
			_ = s.generics.add(char, "", true)
			// dbg.AssertOk(ok, "s.generics.Add(%s, %q, true)", strconv.QuoteRune(char), "")
		}
	}

	return nil
}

// set_type_list_flag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Any negative number will be interpreted as unlimited number of fields. Also, the value 0 will not set the flag.
//
// Documentation:
//
// **Flag: Types**
//
// The "types" flag is used to specify a list of types that are accepted by the generator.
//
// Its argument is specidied as a list of Go types separated by commas without spaces.
//
// For instance, running the following command:
//
//	//go:generate table -name=IntTable -type=int -fields=a/int,b/int,name/string
//
// will generate a tree node with the following fields:
//
//	type TreeNode struct {
//		// Node pointers.
//
//		a int
//		b int
//		name string
//	}
//
// It is important to note that spaces are not allowed.
//
// Also, it is possible to specify generics by following the value with the generics between square brackets;
// like so: "a/MyType[T,C]"
func set_type_list_flag(flag_name string, is_required bool, count int, brief string) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	type_list_flag = &type_list_val{
		types:       make([]string, 0),
		is_required: is_required,
		count:       count,
		generics:    new_ordered_map[rune, string](),
	}

	var usage strings.Builder

	usage.WriteString(brief)

	if is_required {
		if count == -1 {
			usage.WriteString("It must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It must be set with exactly %d fields.", count))
		}
	} else {
		if count == -1 {
			usage.WriteString("It is optional but, if set, it must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It is optional but, if set, it must be set with exactly %d fields.", count))
		}
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(type_list_flag, flag_name, usage.String())
}

// Type returns the type at the given index.
//
// Parameters:
//   - idx: The index of the type to return.
//
// Return:
//   - string: The type at the given index.
//   - error: An error of type *luc.ErrInvalidParameter if the index is out of bounds.
func (s type_list_val) Type(idx int) (string, error) {
	if idx < 0 || idx >= len(s.types) {
		return "", NewErrInvalidParameter("idx", NewErrOutOfBounds(idx, 0, len(s.types)))
	}

	return s.types[idx], nil
}

// print_flags prints the default values of the flags.
//
// It is useful for debugging and for error messages.
func print_flags() {
	flag.PrintDefaults()
}
