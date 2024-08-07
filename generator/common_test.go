package generator

import (
	"testing"
)

func TestIsValidName(t *testing.T) {
	err := is_valid_name("tn", []string{"child"}, not_exported)
	if err != nil {
		t.Errorf("IsValidName failed: %s", err.Error())
	}
}

func TestFixImportDir(t *testing.T) {
	fixed, err := fix_import_dir("stack.go")
	if err != nil {
		t.Errorf("FixImportDir failed: %s", err.Error())
	}

	if fixed != "stack.go" {
		t.Errorf("FixImportDir failed: expected %s, got %s", "stack.go", fixed)
	}
}
