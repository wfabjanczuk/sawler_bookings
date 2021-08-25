package forms

type errors map[string][]string

func (e errors) Add(field string, message string) {
	e[field] = append(e[field], message)
}

func (e errors) GetFirst(field string) string {
	es := e[field]

	if len(es) == 0 {
		return ""
	}

	return es[0]
}
