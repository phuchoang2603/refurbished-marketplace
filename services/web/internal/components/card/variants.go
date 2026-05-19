package card

import "refurbished-marketplace/services/web/internal/utils"

func cardVariants(className string) string {
	baseClasses := "bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm"
	return utils.TwMerge(baseClasses, className)
}

func cardHeaderVariants(className string) string {
	baseClasses := "grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6"
	return utils.TwMerge(baseClasses, className)
}

func cardTitleVariants(className string) string {
	return utils.TwMerge("leading-none font-semibold", className)
}

func cardDescriptionVariants(className string) string {
	return utils.TwMerge("text-muted-foreground text-sm", className)
}

func cardActionVariants(className string) string {
	return utils.TwMerge("self-start justify-self-end", className)
}

func cardContentVariants(className string) string {
	return utils.TwMerge("px-6", className)
}

func cardFooterVariants(className string) string {
	return utils.TwMerge("flex items-center px-6", className)
}
