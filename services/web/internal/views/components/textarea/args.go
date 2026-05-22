package textarea

import "github.com/a-h/templ"

type TextareaArgs struct {
	Class       string
	Placeholder string
	Value       string
	Name        string
	ID          string
	FormID      string
	Rows        int
	Disabled    bool
	Required    bool
	Attributes  templ.Attributes
}
