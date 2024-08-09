package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
	"text/template"

	gcers "github.com/PlayerR9/go-commons/errors"
)

// PackageNameSetter is the interface that all generators must implement.
type PackageNameSetter interface {
	// SetPackageName sets the package name for the generated code.
	//
	// Parameters:
	//   - pkg_name: The package name to use for the generated code.
	SetPackageName(pkg_name string)
}

// DoFunc is the type of the function to perform on the data before generating the code.
//
// Parameters:
//   - T: The data to perform the function on.
//
// Returns:
//   - error: An error if occurred.
type DoFunc[T PackageNameSetter] func(T) error

// CodeGenerator is the code generator.
type CodeGenerator[T PackageNameSetter] struct {
	// t is the template to use for the generated code.
	templ *template.Template

	// do_funcs is the list of functions to perform on the data before generating the code.
	do_funcs []DoFunc[T]
}

// NewCodeGenerator creates a new code generator.
//
// Parameters:
//   - templ: The template to use for the generated code.
//
// Returns:
//   - *CodeGenerator: The code generator.
//
// This function returns nil iff templ is nil.
func NewCodeGenerator[T PackageNameSetter](templ *template.Template) *CodeGenerator[T] {
	if templ == nil {
		return nil
	}

	return &CodeGenerator[T]{
		templ:    templ,
		do_funcs: make([]DoFunc[T], 0),
	}
}

// NewCodeGeneratorFromTemplate creates a new code generator from a template. Panics
// if the template is not valid.
//
// Parameters:
//   - name: The name of the template.
//   - templ: The template to use for the generated code.
//
// Returns:
//   - *CodeGenerator: The code generator.
//   - error: An error if template.Parse fails.
func NewCodeGeneratorFromTemplate[T PackageNameSetter](name, templ string) (*CodeGenerator[T], error) {
	t, err := template.New(name).Parse(templ)
	if err != nil {
		return nil, err
	}

	return &CodeGenerator[T]{
		templ:    t,
		do_funcs: make([]DoFunc[T], 0),
	}, nil
}

// AddDoFunc adds a function to perform on the data before generating the code.
//
// Parameters:
//   - do_func: The function to perform on the data before generating the code.
//
// Does nothing if the do_func is nil.
func (cg *CodeGenerator[T]) AddDoFunc(do_func DoFunc[T]) {
	if do_func == nil {
		return
	}

	cg.do_funcs = append(cg.do_funcs, do_func)
}

// fix_import_dir takes a destination string and manipulates it to get the correct import path.
//
// Parameters:
//   - dest: The destination path.
//
// Returns:
//   - string: The correct import path.
//   - error: An error if there is any.
func fix_import_dir(dest string) (string, error) {
	if dest == "" {
		dest = "."
	}

	dir := filepath.Dir(dest)
	if dir == "." {
		pkg, err := build.ImportDir(".", 0)
		if err != nil {
			return "", err
		}

		return pkg.Name, nil
	}

	_, right := filepath.Split(dir)
	return right, nil
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
func fix_loc(loc string) (string, error) {
	if loc == "" {
		return "", errors.New("flag must be set")
	}

	// Assumption: default_file_name is never empty.

	before, after := filepath.Split(loc)

	after = strings.ToLower(after)

	ext := filepath.Ext(after)
	if ext == "" {
		return "", errors.New("location cannot be a directory")
	} else if ext != go_ext {
		return "", errors.New("location must be a .go file")
	}

	return before + after, nil
}

// Generate generates code using the given generator and writes it to the given destination file.
//
// WARNING:
//   - Remember to call this function iff the function go-generator.SetOutputFlag() was called
//     and only after the function flag.Parse() was called.
//
// Parameters:
//   - file_name: The file name to use for the generated code.
//   - suffix: The suffix to use for the generated code. This should end with the ".go" extension.
//   - data: The data to use for the generated code.
//
// Returns:
//   - string: The output location of the generated code.
//   - error: An error if occurred.
//
// Errors:
//   - *common.ErrInvalidParameter: If the file_name or suffix is an empty string.
//   - error: Any other type of error that may have occurred.
func (cg CodeGenerator[T]) GenerateWithLoc(loc string, data T) (*Generated, error) {
	if loc == "" {
		return nil, gcers.NewErrInvalidParameter("loc", gcers.NewErrEmpty(loc))
	}

	if cg.templ == nil {
		panic("cg.templ is nil")
	}

	// dbg.AssertNil(cg.templ, "cg.templ")

	// NOTES: By extracting FixOutputLoc and FixImportDir to a separate function,
	// we can remove the dependency on the Generater interface. Suggested to do so
	// as part of the refactoring.

	g := &Generated{}

	output_loc, err := fix_loc(loc)
	if err != nil {
		return g, fmt.Errorf("failed to fix output location: %w", err)
	}

	g.DestLoc = output_loc

	pkg_name, err := fix_import_dir(output_loc)
	if err != nil {
		return g, fmt.Errorf("failed to fix import path: %w", err)
	}

	data.SetPackageName(pkg_name)

	for _, f := range cg.do_funcs {
		err := f(data)
		if err != nil {
			return g, err
		}
	}

	var buff bytes.Buffer

	err = cg.templ.Execute(&buff, data)
	if err != nil {
		return g, err
	}

	g.Data = buff.Bytes()

	return g, nil
}

// Generate generates code using the given generator and writes it to the given destination file.
//
// WARNING:
//   - Remember to call this function iff the function go-generator.SetOutputFlag() was called
//     and only after the function flag.Parse() was called.
//
// Parameters:
//   - file_name: The file name to use for the generated code.
//   - suffix: The suffix to use for the generated code. This should end with the ".go" extension.
//   - data: The data to use for the generated code.
//
// Returns:
//   - string: The output location of the generated code.
//   - error: An error if occurred.
//
// Errors:
//   - *common.ErrInvalidParameter: If the file_name or suffix is an empty string.
//   - error: Any other type of error that may have occurred.
func (cg CodeGenerator[T]) Generate(o *OutputLocVal, default_file_name string, data T) (*Generated, error) {
	if o == nil {
		return nil, gcers.NewErrInvalidUsage(
			errors.New("output location was not defined"),
			"Please call the go-generator.NewOutputFlag() function before calling this function.",
		)
	}

	if cg.templ == nil {
		panic("cg.templ is nil")
	}

	if default_file_name == "" {
		return nil, gcers.NewErrInvalidParameter("file_name", gcers.NewErrEmpty(default_file_name))
	}

	// dbg.AssertNil(cg.templ, "cg.templ")

	// NOTES: By extracting FixOutputLoc and FixImportDir to a separate function,
	// we can remove the dependency on the Generater interface. Suggested to do so
	// as part of the refactoring.

	g := &Generated{}

	output_loc, err := o.fix(default_file_name)
	if err != nil {
		return g, fmt.Errorf("failed to fix output location: %w", err)
	}

	g.DestLoc = output_loc

	pkg_name, err := fix_import_dir(output_loc)
	if err != nil {
		return g, fmt.Errorf("failed to fix import path: %w", err)
	}

	data.SetPackageName(pkg_name)

	for _, f := range cg.do_funcs {
		err := f(data)
		if err != nil {
			return g, err
		}
	}

	var buff bytes.Buffer

	err = cg.templ.Execute(&buff, data)
	if err != nil {
		return g, err
	}

	g.Data = buff.Bytes()

	return g, nil
}
