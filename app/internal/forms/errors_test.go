package forms

import "testing"

func TestErrors(t *testing.T) {
	testErrors := errors{}

	testErrors.Add("first_field", "first_error")
	testErrors.Add("first_field", "second_error")

	if testErrors.GetFirst("non-existent") != "" {
		t.Error("Func GetFirst returned non-empty string for non-existent field")
	}

	if testErrors.GetFirst("first_field") != "first_error" {
		t.Error("Func GetFirst returned wrong string for field with errors")
	}
}
