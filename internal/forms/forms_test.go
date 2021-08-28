package forms

import (
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	form := New(url.Values{
		"test": []string{"value"},
	})

	if !form.Valid() {
		t.Error("Form with no errors is incorrectly considered as invalid")
	}

	form.Errors.Add("test", "Some error")

	if form.Valid() {
		t.Error("Form with an error is incorrectly considered as valid")
	}
}

func TestForm_Required(t *testing.T) {
	form := New(url.Values{
		"req1": []string{"value"},
		"req2": []string{"value"},
		"skip": []string{"value"},
	})

	if !form.Required("req1", "req2") {
		t.Error("Form has all the required fields and func Required returns false")
	}

	if !form.Valid() {
		t.Error("Form has all the required fields and is incorrectly considered as invalid")
	}

	if form.Required("req3") {
		t.Error("Form does not have one field and func Required returns true")
	}

	if form.Valid() {
		t.Error("Form is missing one field and is incorrectly considered as valid")
	}
}

func TestForm_MinLength(t *testing.T) {
	form := New(url.Values{
		"short": []string{"value"},
		"long":  []string{"value-value-value-value-value"},
	})

	if !form.MinLength("long", 10) {
		t.Error("Field satisfies minimum length and func MinLength returns false")
	}

	if !form.Valid() {
		t.Error("Field satisfies minimum length and the form is incorrectly considered as invalid")
	}

	if form.MinLength("short", 10) {
		t.Error("Field does not satisfy minimum length and func MinLength returns true")
	}

	if form.Valid() {
		t.Error("Field is too short and the form is incorrectly considered as valid")
	}
}

func TestForm_IsEmail(t *testing.T) {
	form := New(url.Values{
		"email":     []string{"trevor@sawler.com"},
		"not-email": []string{"trevor.at.sawler.com"},
	})

	if !form.IsEmail("email") {
		t.Error("Field is an email and func IsEmail returns false")
	}

	if !form.Valid() {
		t.Error("Field is an email and the form is incorrectly considered as invalid")
	}

	if form.IsEmail("not-email") {
		t.Error("Field is not an email and func IsEmail returns true")
	}

	if form.Valid() {
		t.Error("Field is not an email and the form is incorrectly considered as valid")
	}
}
