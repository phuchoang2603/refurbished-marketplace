package card

import "github.com/Oudwins/tailwind-merge-go/pkg/twmerge"

func cardVariants(className string) string {
	baseClasses := "bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm"
	return twmerge.Merge(baseClasses, className)
}

func cardHeaderVariants(className string) string {
	baseClasses := "grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6"
	return twmerge.Merge(baseClasses, className)
}

func cardTitleVariants(className string) string {
	return twmerge.Merge("leading-none font-semibold", className)
}

func cardDescriptionVariants(className string) string {
	return twmerge.Merge("text-muted-foreground text-sm", className)
}

func cardActionVariants(className string) string {
	return twmerge.Merge("self-start justify-self-end", className)
}

func cardContentVariants(className string) string {
	return twmerge.Merge("px-6", className)
}

func cardFooterVariants(className string) string {
	return twmerge.Merge("flex items-center px-6", className)
}
