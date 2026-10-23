package productsv1

func ValidCategory(value string) bool {
	switch value {
	case "phone", "laptop", "tablet", "accessory", "headphones", "smartwatch", "game_console", "camera":
		return true
	default:
		return false
	}
}

func ValidCondition(value string) bool {
	switch value {
	case "excellent", "good", "fair":
		return true
	default:
		return false
	}
}
