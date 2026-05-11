package views

import (
	"maps"
	"strings"

	"github.com/a-h/templ"
)

type ButtonArgs struct {
	Type       string
	Variant    string
	Class      string
	Disabled   bool
	Attributes templ.Attributes
}

type CardArgs struct {
	ID         string
	Class      string
	Attributes templ.Attributes
}

type SlotArgs struct {
	ID         string
	Class      string
	Attributes templ.Attributes
}

type FormArgs struct {
	ID         string
	Action     string
	Method     string
	Datastar   string
	Class      string
	Attributes templ.Attributes
}

type LabelArgs struct {
	For        string
	Class      string
	HasError   bool
	Attributes templ.Attributes
}

type InputArgs struct {
	ID           string
	Name         string
	Type         string
	Value        string
	Autocomplete string
	Min          string
	Required     bool
	Disabled     bool
	Class        string
	Attributes   templ.Attributes
}

func classNames(values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " ")
}

func componentAttrs(attrs templ.Attributes, slot string, class string, id string) templ.Attributes {
	merged := templ.Attributes{}
	maps.Copy(merged, attrs)
	if id != "" {
		merged["id"] = id
	}
	if class != "" {
		merged["class"] = class
	}
	if slot != "" {
		merged["data-slot"] = slot
	}
	return merged
}

func formAttrs(args FormArgs) templ.Attributes {
	method := args.Method
	if method == "" {
		method = "post"
	}
	attrs := componentAttrs(args.Attributes, "form", classNames("form-stack", args.Class), args.ID)
	attrs["method"] = method
	if args.Action != "" {
		attrs["action"] = args.Action
	}
	if args.Datastar != "" {
		attrs["data-on-submit__prevent"] = args.Datastar
		attrs["data-indicator-fetching"] = ""
		attrs["data-attr-aria-busy"] = "$fetching ? 'true' : 'false'"
	}
	return attrs
}

func datastarFormAction(method string, action string) string {
	return "@" + method + "('" + action + "', {contentType: 'form'})"
}
