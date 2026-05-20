package button

import "github.com/a-h/templ"

type ButtonArgs struct {
	Variant    string
	Size       string
	AsChild    bool
	Class      string
	Attributes templ.Attributes
	Disabled   bool
	Type       string
}
