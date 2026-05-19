package label

import "refurbished-marketplace/services/web/internal/utils"

func labelVariants(className string) string {
	baseClasses := "flex items-center gap-2 text-sm leading-none font-medium select-none peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
	return utils.TwMerge(baseClasses, className)
}
