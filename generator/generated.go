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

// WriteFile writes the generated code to the destination file.
//
// Parameters:
//   - suffix: The suffix to add to the file name. If empty, no suffix is added.
//   - sub_directories: The sub directories to create the file in.
//
// Returns:
//   - string: The location of the generated code.
//   - error: An error if occurred.
//
// The suffix is useful for when generating multiple files as it adds a suffix without
// changing the extension.
func (g Generated) WriteFile(suffix string, sub_directories ...string) (string, error) {
	var loc string

	if len(sub_directories) > 0 {
		dir, file := filepath.Split(g.DestLoc)

		loc = filepath.Join(dir, filepath.Join(sub_directories...), file)
	}

	if suffix != "" {
		loc = strings.TrimSuffix(loc, go_ext) + suffix + go_ext
	}

	dir := filepath.Dir(loc)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return loc, err
	}

	err = os.WriteFile(loc, g.Data, 0644)
	if err != nil {
		return loc, err
	}

	return loc, nil
}
