package form

import "refurbished-marketplace/services/web/internal/utils"

func formItemVariants(className string) string {
	return utils.TwMerge("grid gap-2", className)
}

func formLabelVariants(className string, hasError bool) string {
	base := ""
	if hasError {
		base = "text-destructive"
	}
	return utils.TwMerge(base, className)
}

func formDescriptionVariants(className string) string {
	return utils.TwMerge("text-muted-foreground text-sm", className)
}

func formMessageVariants(className string) string {
	return utils.TwMerge("text-destructive text-sm", className)
}
