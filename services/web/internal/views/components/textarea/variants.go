package textarea

import "github.com/Oudwins/tailwind-merge-go/pkg/twmerge"

func textareaVariants(className string) string {
	baseClasses := "placeholder:text-muted-foreground border-input flex min-h-24 w-full rounded-md border bg-background px-3 py-2 text-base text-foreground shadow-sm transition-colors outline-none disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
	return twmerge.Merge(baseClasses, className)
}
