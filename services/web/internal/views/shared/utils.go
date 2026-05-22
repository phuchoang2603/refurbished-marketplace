package shared

import "fmt"

func FormatCents(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s$%d.%02d", sign, v/100, v%100)
}

func FormatInt32(v int32) string {
	return formatInt64(int64(v))
}

func formatInt64(v int64) string {
	return fmt.Sprintf("%d", v)
}
