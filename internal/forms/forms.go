package forms

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/url"
	"strings"
)

type Form struct {
	url.Values
	Errors errors
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Has(field string) bool {
	value := f.Get(field)

	if strings.TrimSpace(value) == "" {
		f.Errors.Add(field, "This field cannot be blank")

		return false
	}

	return true
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		f.Has(field)
	}
}

func (f *Form) MinLength(field string, length int) bool {
	value := f.Get(field)

	if len(value) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))

		return false
	}

	return true
}

func (f *Form) IsEmail(field string) bool {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")

		return false
	}

	return true
}
