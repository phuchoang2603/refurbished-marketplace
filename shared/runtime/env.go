package runtime

import (
	"os"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
)

func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		sharedlog.Fatal("required env missing", "key", key)
	}
	return v
}
