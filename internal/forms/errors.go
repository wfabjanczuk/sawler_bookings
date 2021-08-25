package forms

type errors map[string][]string

func (e errors) Add(field string, message string) {
	e[field] = append(e[field], message)
}

func (e errors) GetFirst(field string) string {
	fieldErrors := e[field]

	if len(fieldErrors) == 0 {
		return ""
	}

	return fieldErrors[0]
}
