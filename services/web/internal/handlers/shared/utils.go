package shared

import "google.golang.org/protobuf/types/known/timestamppb"

const timestampFormat = "2006-01-02T15:04:05Z07:00"

func FormatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(timestampFormat)
}
