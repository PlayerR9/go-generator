package generator

import (
	"testing"
)

func TestIsValidName(t *testing.T) {
	err := IsValidVariableName("tn", []string{"child"}, NotExported)
	if err != nil {
		t.Errorf("IsValidName failed: %s", err.Error())
	}
}
