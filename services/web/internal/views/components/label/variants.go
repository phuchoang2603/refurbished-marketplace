package label

import "github.com/Oudwins/tailwind-merge-go/pkg/twmerge"

func labelVariants(className string) string {
	baseClasses := "flex items-center gap-2 text-sm leading-none font-medium select-none peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
	return twmerge.Merge(baseClasses, className)
}
