package utils

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const timestampFormat = "2006-01-02T15:04:05Z07:00"

func FormatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(timestampFormat)
}

func FormatCents(v int64) string {
	return "$" + formatInt64(v)
}

func FormatInt32(v int32) string {
	return formatInt64(int64(v))
}

func formatInt64(v int64) string {
	return fmt.Sprintf("%d", v)
}
