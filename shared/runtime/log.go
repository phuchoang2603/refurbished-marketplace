package runtime

import (
	sharedlog "refurbished-marketplace/shared/observe/log"
)

// InitLogging configures JSON slog for the process. Call before other bootstrap logs.
func InitLogging(serviceName string) {
	sharedlog.Init(serviceName)
}
