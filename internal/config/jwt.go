package config

import "time"

const (
	DefaultTokenExp  = 3
	DefaultSecretKey = "testsecretkey"
)

type JWTConfig struct {
	SecretKey string
	TokenExp  time.Duration
}
