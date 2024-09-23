package generator

import (
	"testing"
)

func TestFixImportDir(t *testing.T) {
	fixed, err := fix_import_dir("test/stack.go")
	if err != nil {
		t.Errorf("FixImportDir failed: %s", err.Error())
	}

	if fixed != "test" {
		t.Errorf("FixImportDir failed: expected %s, got %s", "stack.go", fixed)
	}
}
