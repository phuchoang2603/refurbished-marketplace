package form

import "github.com/Oudwins/tailwind-merge-go/pkg/twmerge"

func formItemVariants(className string) string {
	return twmerge.Merge("grid gap-2", className)
}

func formLabelVariants(className string, hasError bool) string {
	base := ""
	if hasError {
		base = "text-destructive"
	}
	return twmerge.Merge(base, className)
}

func formDescriptionVariants(className string) string {
	return twmerge.Merge("text-muted-foreground text-sm", className)
}

func formMessageVariants(className string) string {
	return twmerge.Merge("text-destructive text-sm", className)
}
