package form

import "github.com/a-h/templ"

type FormArgs struct {
	Action     string
	Class      string
	Attributes templ.Attributes
}

type FormItemArgs struct {
	Class      string
	Attributes templ.Attributes
}

type FormLabelArgs struct {
	For        string
	HasError   bool
	Class      string
	Attributes templ.Attributes
}

type FormDescriptionArgs struct {
	ID         string
	Class      string
	Attributes templ.Attributes
}

type FormMessageArgs struct {
	ID         string
	Message    string
	Class      string
	Attributes templ.Attributes
}
