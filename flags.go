package generator

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	gcers "github.com/PlayerR9/errors"
	gcmap "github.com/PlayerR9/go-commons/maps"
	"github.com/dustin/go-humanize"
)

// OutputLocVal is the value of the output_flag flag.
type OutputLocVal struct {
	// loc is the location of the output file.
	loc string

	// def_loc is the default location of the output file.
	def_loc string

	// is_required is whether the flag is required or not.
	is_required bool
}

// String implements the flag.Value interface.
func (v OutputLocVal) String() string {
	return v.loc
}

// Set implements the flag.Value interface.
func (v *OutputLocVal) Set(loc string) error {
	if v == nil {
		return nil
	}

	v.loc = loc
	return nil
}

// NewOutputFlag sets the flag that specifies the location of the output file.
//
// Parameters:
//   - def_value: The default value of the output_flag flag.
//   - required: Whether the flag is required or not.
//
// Returns:
//   - *OutputLocVal: The new output_flag flag. Never returns nil.
//
// Here are all the possible valid calls to this function:
//
//	NewOutputFlag("", false) <-> NewOutputFlag("[no location]", false)
//	NewOutputFlag("path/to/file.go", false)
//	NewOutputFlag("", true) <-> NewOutputFlag("path/to/file.go", true)
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
func NewOutputFlag(def_value string, required bool) *OutputLocVal {
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

	value := &OutputLocVal{
		def_loc:     def_value,
		is_required: required,
	}

	flag.Var(value, "o", usage)

	return value
}

// fix_output_loc fixes the output location.
//
// Parameters:
//   - file_name: The name of the file.
//   - suffix: The suffix of the file.
//
// Returns:
//   - string: The output location.
//   - error: An error if any.
//
// Errors:
//   - *common.ErrInvalidParameter: If the file name is empty.
//   - *common.ErrInvalidUsage: If the OutputLoc flag was not set.
//   - error: Any other error that may have occurred.
//
// The suffix parameter must end with the ".go" extension. Plus, the output
// location is always lowercased.
//
// NOTES: This function only sets the output location if the user did not set
// the output flag. If they did, this function won't do anything but the necessary
// checks and validations.
//
// Example:
//
//	loc, err := fix_output_loc("test", ".go")
//	if err != nil {
//	  panic(err)
//	}
//
//	fmt.Println(loc) // test.go
func (o *OutputLocVal) fix(default_file_name string) (string, error) {
	if o.loc == "" {
		if o.is_required {
			return "", errors.New("flag must be set")
		}

		o.loc = default_file_name
	}

	// Assumption: default_file_name is never empty.

	before, after := filepath.Split(o.loc)

	after = strings.ToLower(after)

	ext := filepath.Ext(after)
	if ext == "" {
		return "", errors.New("location cannot be a directory")
	} else if ext != go_ext {
		return "", errors.New("location must be a .go file")
	}

	return before + after, nil
}

// Loc gets the location of the output file.
//
// Returns:
//   - string: The location of the output file.
func (o OutputLocVal) Loc() string {
	return o.loc
}

// struct_fields_vaò is a struct that represents the fields value.
type StructFieldsVal struct {
	// fields is a map of the fields and their types.
	fields *gcmap.OrderedMap[string, string]

	// generics is a map of the generics and their types.
	generics *gcmap.OrderedMap[rune, string]

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
func (s StructFieldsVal) String() string {
	if s.fields.Size() == 0 {
		return ""
	}

	var values []string
	var builder strings.Builder

	for k, v := range s.fields.Entry() {
		builder.WriteString(k)
		builder.WriteRune(' ')
		builder.WriteString(v)

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *StructFieldsVal) Set(value string) error {
	if s == nil {
		return nil
	}

	if value == "" && s.is_required {
		return errors.New("value must be set")
	}

	fields := strings.Split(value, ",")

	s.fields = gcmap.NewOrderedMap[string, string]()

	for i, field := range fields {
		if field == "" {
			continue
		}

		sub_fields := strings.Split(field, "/")

		if len(sub_fields) == 1 {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" field", errors.New("missing type"))
		} else if len(sub_fields) > 2 {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" field", errors.New("too many fields"))
		}

		ok := s.fields.Add(sub_fields[0], sub_fields[1], false)
		if !ok {
			return fmt.Errorf("field %q already exists", sub_fields[0])
		}
	}

	size := s.fields.Size()

	if s.count != -1 && size != s.count {
		return fmt.Errorf("wrong number of fields: expected %d, got %d", s.count, size)
	}

	s.generics = gcmap.NewOrderedMap[rune, string]()

	for _, field_type := range fields {
		if field_type == "" {
			continue
		}

		chars, err := parse_generics(field_type)
		ok := IsErrNotGeneric(err)

		if ok {
			continue
		} else if err != nil {
			return fmt.Errorf("syntax error for type %q: %w", field_type, err)
		}

		for _, char := range chars {
			_ = s.generics.Add(char, "", false)
			// dbg.AssertOk(ok, "s.generics.Add(%s, %q, false)", strconv.QuoteRune(char), "")
		}
	}

	return nil
}

// NewStructFieldsFlag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Returns:
//   - *StructFieldsVal: The value of the flag.
//
// This function returns nil iff count is 0.
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
func NewStructFieldsFlag(flag_name string, is_required bool, count int, brief string) *StructFieldsVal {
	if count == 0 {
		return nil
	}

	if count < 0 {
		count = -1
	}

	value := &StructFieldsVal{
		fields:      gcmap.NewOrderedMap[string, string](),
		generics:    gcmap.NewOrderedMap[rune, string](),
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

	flag.Var(value, flag_name, usage.String())

	return value
}

// Fields returns the fields of the struct.
//
// Returns:
//   - map[string]string: A map of field names and their types. Never returns nil.
func (s StructFieldsVal) Fields() map[string]string {
	return s.fields.Map()
}

// Generics returns the letters of the generics.
//
// Returns:
//   - []rune: The letters of the generics.
func (s StructFieldsVal) Generics() []rune {
	return s.generics.Keys()
}

// MakeParameterList makes a string representing a list of parameters.
//
// WARNING: Call this function only if StructFieldsFlag is set.
//
// Parameters:
//   - fields: A map of field names and their types.
//
// Returns:
//   - string: A string representing the parameters.
//   - error: An error if any.
func (s StructFieldsVal) MakeParameterList() (string, error) {
	var field_list []string
	var type_list []string

	for k, v := range s.fields.Entry() {
		if k == "" {
			return "", errors.New("found type name with empty name")
		}

		first_letter, _ := utf8.DecodeRuneInString(k)
		if first_letter == utf8.RuneError {
			return "", errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsLetter(first_letter)
		if !ok {
			return "", fmt.Errorf("type name %q must start with a letter", k)
		}

		ok = unicode.IsUpper(first_letter)
		if !ok {
			return "", nil
		}

		pos, ok := slices.BinarySearch(field_list, k)
		if ok {
			panic(fmt.Sprintf("slices.BinarySearch(field_list, %q) should never succeed", k))
		}

		field_list = slices.Insert(field_list, pos, k)
		type_list = slices.Insert(type_list, pos, v)
	}

	var values []string
	var builder strings.Builder

	for i := 0; i < len(field_list); i++ {
		param := strings.ToLower(field_list[i])

		_, ok := slices.BinarySearch(go_reserved_keywords, param)
		if ok {
			param += "_"
		}

		builder.WriteString(param)
		builder.WriteRune(' ')
		builder.WriteString(type_list[i])

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")

	return joined_str, nil
}

// MakeAssignmentList makes a string representing a list of assignments.
//
// WARNING: Call this function only if StructFieldsFlag is set.
//
// Parameters:
//   - fields: A map of field names and their types.
//
// Returns:
//   - string: A string representing the assignments.
//   - error: An error if any.
func (s StructFieldsVal) MakeAssignmentList() (map[string]string, error) {
	var field_list []string
	var type_list []string

	for k, v := range s.fields.Entry() {
		if k == "" {
			return nil, errors.New("found type name with empty name")
		}

		first_letter, _ := utf8.DecodeRuneInString(k)
		if first_letter == utf8.RuneError {
			return nil, errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsLetter(first_letter)
		if !ok {
			return nil, fmt.Errorf("type name %q must start with a letter", k)
		}

		ok = unicode.IsUpper(first_letter)
		if !ok {
			return nil, nil
		}

		pos, ok := slices.BinarySearch(field_list, k)
		if ok {
			panic(fmt.Sprintf("slices.BinarySearch(field_list, %q) should never succeed", k))
		}

		field_list = slices.Insert(field_list, pos, k)
		type_list = slices.Insert(type_list, pos, v)
	}

	assignment_map := make(map[string]string)

	for i := 0; i < len(field_list); i++ {
		param := strings.ToLower(field_list[i])

		_, ok := slices.BinarySearch(go_reserved_keywords, param)
		if ok {
			param = "elem_" + param
		}

		assignment_map[field_list[i]] = param
	}

	return assignment_map, nil
}

// GenericsSignVal is a struct that contains the values of the generics.
type GenericsSignVal struct {
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
func (s GenericsSignVal) String() string {
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
func (s *GenericsSignVal) Set(value string) error {
	if s == nil {
		return nil
	}

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
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" field", err)
		}

		err = s.add(letter, g_type)
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" field", err)
		}
	}

	if s.count != -1 && len(s.letters) != s.count {
		return fmt.Errorf("invalid number of generics: expected %d, got %d", s.count, len(s.letters))
	}

	return nil
}

// NewGenericsSignFlag sets the flag that specifies the generics to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of generics. If -1, no upper bound is set, 0 means no generics.
//
// Returns:
//   - *GenericsSignVal: The value of the flag.
//
// This function returns nil iff count is 0.
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
func NewGenericsSignFlag(flag_name string, is_required bool, count int) *GenericsSignVal {
	if count == 0 {
		return nil
	}

	if count < 0 {
		count = -1
	}

	value := &GenericsSignVal{
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

	flag.Var(value, flag_name, usage.String())

	return value
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
func (gv *GenericsSignVal) add(letter rune, g_type string) error {
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
func (gv GenericsSignVal) Signature() string {
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

// TypeListVal is a struct that represents a list of types.
type TypeListVal struct {
	// fields is a list of types.
	types []string

	// generics is a map of the generics and their types.
	generics *gcmap.OrderedMap[rune, string]

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
func (s TypeListVal) String() string {
	if len(s.types) == 0 {
		return ""
	}

	joined_str := strings.Join(s.types, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *TypeListVal) Set(value string) error {
	if s == nil {
		return gcers.NewErrNilParameter("TypeListVal")
	}

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

	s.generics = gcmap.NewOrderedMap[rune, string]()

	for _, field_type := range s.types {
		if field_type == "" {
			continue
		}

		chars, err := parse_generics(field_type)
		ok := IsErrNotGeneric(err)

		if ok {
			continue
		} else if err != nil {
			return fmt.Errorf("syntax error for type %q: %w", field_type, err)
		}

		for _, char := range chars {
			_ = s.generics.Add(char, "", true)
			// dbg.AssertOk(ok, "s.generics.Add(%s, %q, true)", strconv.QuoteRune(char), "")
		}
	}

	return nil
}

// NewTypeListFlag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Returns:
//   - *TypeListVal: The flag value.
//
// This function returns nil iff count is 0.
//
// Any negative number will be interpreted as unlimited number of fields. Also, the value 0 will not set the flag.
// If value is nil, it will panic.
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
func NewTypeListFlag(flag_name string, is_required bool, count int, brief string) *TypeListVal {
	if count == 0 {
		return nil
	}

	if count < 0 {
		count = -1
	}

	value := &TypeListVal{
		types:       make([]string, 0),
		generics:    gcmap.NewOrderedMap[rune, string](),
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

	flag.Var(value, flag_name, usage.String())

	return value
}

// Type returns the type at the given index.
//
// Parameters:
//   - idx: The index of the type to return.
//
// Return:
//   - string: The type at the given index.
//   - error: An error of type *luc.ErrInvalidParameter if the index is out of bounds.
func (s TypeListVal) Type(idx int) (string, error) {
	if idx < 0 || idx >= len(s.types) {
		err := gcers.NewErrInvalidParameter(fmt.Sprintf("idx must be in the range [0, %d]", len(s.types)))
		err.AddFrame("TypeListVal", "Type()")

		return "", err
	}

	return s.types[idx], nil
}

// Generics returns the generics of the struct.
//
// Returns:
//   - []rune: The generics of the struct.
func (s TypeListVal) Generics() []rune {
	return s.generics.Keys()
}

/////////////////////////////////////////////////////
