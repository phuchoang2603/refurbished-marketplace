package shared

import "fmt"

func FormatCents(v int64) string {
	return "$" + formatInt64(v)
}

func FormatInt32(v int32) string {
	return formatInt64(int64(v))
}

func formatInt64(v int64) string {
	return fmt.Sprintf("%d", v)
}
