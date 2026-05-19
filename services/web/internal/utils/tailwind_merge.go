package utils

import twmerge "github.com/Oudwins/tailwind-merge-go"

func TwMerge(classes ...string) string {
	return twmerge.Merge(classes...)
}
