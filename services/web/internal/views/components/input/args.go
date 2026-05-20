package input

import "github.com/a-h/templ"

type InputArgs struct {
	Type         string
	Class        string
	Placeholder  string
	Value        string
	Name         string
	ID           string
	FormID       string
	Autocomplete string
	Min          string
	Disabled     bool
	Required     bool
	Attributes   templ.Attributes
}
