package config

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	SConfig     ServerConfig
	DBConfig    DBConfig
	LConfig     LoggerConfig
	JWTConfig   JWTConfig
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
		JWTConfig: JWTConfig{
			TokenExp:  DefaultTokenExp * time.Hour,
			SecretKey: DefaultSecretKey,
		},
	}
}

func (c *Config) Build() error {
	err := c.envBuild()
	if err != nil {
		return fmt.Errorf("error build env config: %w", err)
	}

	err = c.flagBuild()
	if err != nil {
		return fmt.Errorf("error build flags config: %w", err)
	}

	if c.DBConfig.DSN == "" {
		return fmt.Errorf("error build config: %w", errors.New("database source name is empty"))
	}

	return nil
}
