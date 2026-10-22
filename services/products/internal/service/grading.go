package service

func validCategory(value string) bool {
	switch value {
	case "phone", "laptop", "tablet", "accessory", "headphones", "smartwatch", "game_console", "camera":
		return true
	default:
		return false
	}
}

func validCondition(value string) bool {
	switch value {
	case "excellent", "good", "fair":
		return true
	default:
		return false
	}
}
