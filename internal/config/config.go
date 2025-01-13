package config

import (
	"errors"
	"fmt"
)

type Config struct {
	DBConfig    DBConfig
	SConfig     ServerConfig
	LConfig     LoggerConfig
	RetryConfig RetryConfig
}

func New() *Config {
	return &Config{
		SConfig: ServerConfig{
			Address: DefaultServerAddress,
		},
		LConfig: LoggerConfig{
			Level: DefaultLogLevel,
		},
		DBConfig: DBConfig{},
		RetryConfig: RetryConfig{
			MinDelay:   DefaultMinRetryDelay,
			MaxDelay:   DefaultMaxRetryDelay,
			MaxAttempt: DefaultMaxRetryAttempt,
		},
	}
}

func (c *Config) Build() error {
	c.flagBuild()

	c.envBuild()

	if c.DBConfig.DSN == "" {
		return fmt.Errorf("error build config: %w", errors.New("database source name is empty"))
	}

	return nil
}
