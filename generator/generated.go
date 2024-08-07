package generator

import (
	"os"
	"path/filepath"

	gcfm "github.com/PlayerR9/go-commons/file_manager"
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
//
// Returns:
//   - string: The location of the generated code.
//   - error: An error if occurred.
//
// The suffix is useful for when generating multiple files as it adds a suffix without
// changing the extension.
func (g Generated) WriteFile(suffix string) (string, error) {
	var loc string

	if suffix != "" {
		loc = gcfm.AddSuffixToFileName(g.DestLoc, suffix, go_ext)
	} else {
		loc = g.DestLoc
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
