package file_manager

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	gcstr "github.com/PlayerR9/go-commons/strings"
)

// AddSuffixToFileName adds a suffix to a file name.
//
// Parameters:
//   - filename: The name of the file.
//   - new_suffix: The new suffix to add.
//   - ext: The extension of the file. If not provided, it will be inferred
//     from the filename.
//
// Returns:
//   - string: The new file name with the suffix added.
//
// This function returns an empty string if the filename is empty.
func AddSuffixToFileName(filename, new_suffix string, ext string) string {
	if filename == "" || new_suffix == "" {
		return filename
	}

	if ext == "" {
		ext = filepath.Ext(filename)

		if ext == "" {
			return filename + new_suffix
		}
	}

	loc := strings.TrimSuffix(filename, ext)

	var builder strings.Builder

	builder.WriteString(loc)
	builder.WriteString(new_suffix)
	builder.WriteString(ext)

	return builder.String()
}

func ErrIfInvalidExt(file_name string, exts ...string) error {
	if file_name == "" {
		return errors.New("no file name provided")
	}

	ext := filepath.Ext(file_name)

	if ext == "" {
		return errors.New("expected file, got directory instead")
	}

	for _, e := range exts {
		if ext == e {
			return nil
		}
	}

	return fmt.Errorf("invalid file extension: %s", gcstr.OrString(exts, true, true))
}
