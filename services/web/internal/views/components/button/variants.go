package button

import "github.com/Oudwins/tailwind-merge-go/pkg/twmerge"

func buttonVariants(variant, size, className string) string {
	baseClasses := "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"

	variantClasses := map[string]string{
		"default":     "bg-primary text-primary-foreground shadow-sm hover:bg-primary/90",
		"destructive": "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
		"outline":     "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
		"secondary":   "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
		"ghost":       "hover:bg-accent hover:text-accent-foreground",
		"link":        "text-primary underline-offset-4 hover:underline",
	}

	sizeClasses := map[string]string{
		"default": "h-9 px-4 py-2",
		"sm":      "h-8 px-3",
		"lg":      "h-10 px-6",
		"icon":    "size-9",
	}

	if variant == "" {
		variant = "default"
	}
	if size == "" {
		size = "default"
	}

	classes := []string{baseClasses}
	if variantClass := variantClasses[variant]; variantClass != "" {
		classes = append(classes, variantClass)
	}
	if sizeClass := sizeClasses[size]; sizeClass != "" {
		classes = append(classes, sizeClass)
	}
	if className != "" {
		classes = append(classes, className)
	}

	return twmerge.Merge(classes...)
}
