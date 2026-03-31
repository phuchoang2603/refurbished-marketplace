package config

import "time"

type Config struct {
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
}

const (
	DefaultJWTIssuer   = "refurbished-marketplace"
	DefaultJWTAudience = "refurbished-marketplace-api"
)

const (
	DefaultJWTAccessTTL  = 15 * time.Minute
	DefaultJWTRefreshTTL = 168 * time.Hour
)

func DefaultConfig(secret string) Config {
	return Config{
		JWTSecret:   secret,
		JWTIssuer:   DefaultJWTIssuer,
		JWTAudience: DefaultJWTAudience,
	}
}
