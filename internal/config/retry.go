package config

import "time"

const (
	DefaultMinRetryDelay   = time.Second
	DefaultMaxRetryDelay   = 5 * time.Second
	DefaultMaxRetryAttempt = 3
)

type RetryConfig struct {
	MinDelay   time.Duration
	MaxDelay   time.Duration
	MaxAttempt int
}
