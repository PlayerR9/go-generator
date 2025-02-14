package generator

import (
	"os"
	"path/filepath"
	"strings"
)

// go_ext is the extension of Go files.
const go_ext string = ".go"

// Generated is the type containing the generated code and its location.
type Generated struct {
	// DestLoc is the destination location of the generated code.
	DestLoc string

	// Data is the data to use for the generated code.
	Data []byte
}

// ModifySuffixPath modifies the path of the generated code.
//
// Parameters:
//   - suffix: The suffix to add to the file name. If empty, no suffix is added.
//   - sub_directories: The sub directories to create the file in.
//
// Returns:
//   - bool: True if the receiver is not nil, false otherwise.
//
// The suffix is useful for when generating multiple files as it adds a suffix without
// changing the extension.
func (g *Generated) ModifySuffixPath(suffix string, sub_directories ...string) bool {
	if g == nil {
		return false
	}

	var loc string

	if len(sub_directories) > 0 {
		dir, file := filepath.Split(g.DestLoc)
		loc = filepath.Join(dir, filepath.Join(sub_directories...), file)
	} else {
		loc = g.DestLoc
	}

	if suffix != "" {
		loc = strings.TrimSuffix(loc, go_ext) + suffix + go_ext
	}

	g.DestLoc = loc

	return true
}

// ModifyPrefixPath modifies the path of the generated code.
//
// Parameters:
//   - prefix: The prefix to add to the file name. If empty, no prefix is added.
//   - sub_directories: The sub directories to create the file in.
//
// Returns:
//   - bool: True if the receiver is not nil, false otherwise.
//
// The prefix is useful for when generating multiple files as it adds a prefix without
// changing the extension.
func (g *Generated) ModifyPrefixPath(prefix string, sub_directories ...string) bool {
	if g == nil {
		return false
	}

	var loc string

	dir, file := filepath.Split(g.DestLoc)

	if len(sub_directories) > 0 {
		loc = filepath.Join(dir, filepath.Join(sub_directories...), prefix+file)
	} else {
		loc = filepath.Join(dir, prefix+file)
	}

	g.DestLoc = loc

	return true
}

// ReplaceFileName replaces the file name of the generated code.
//
// Parameters:
//   - file_name: The file name to use for the generated code.
//
// Returns:
//   - bool: True if the receiver is not nil, false otherwise.
//
// The file name is useful for when generating multiple files as it adds a prefix without
// changing the extension.
func (g *Generated) ReplaceFileName(file_name string) bool {
	if g == nil || file_name == "" {
		return false
	}

	dir, _ := filepath.Split(g.DestLoc)
	g.DestLoc = filepath.Join(dir, file_name)

	return true
}

// WriteFile writes the generated code to the destination file.
//
// Parameters:
//   - suffix: The suffix to add to the file name. If empty, no suffix is added.
//   - sub_directories: The sub directories to create the file in.
//
// Returns:
//   - error: An error if occurred.
//
// The suffix is useful for when generating multiple files as it adds a suffix without
// changing the extension.
func (g Generated) WriteFile() error {
	dir := filepath.Dir(g.DestLoc)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(g.DestLoc, g.Data, 0644)
	if err != nil {
		return err
	}

	return nil
}
