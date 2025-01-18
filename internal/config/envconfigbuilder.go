package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func (c *Config) setEnvServerConfig() {
	if envEndpoint, ok := os.LookupEnv("RUN_ADDRESS3"); ok {
		c.SConfig.Address = envEndpoint
	}
	if envHashKey, ok := os.LookupEnv("KEY"); ok {
		c.SConfig.HashKey = &envHashKey
	}
}

func (c *Config) setEnvLoggerConfig() {
	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.LConfig.Level = envLogLevel
	}
}

func (c *Config) setDBConfig() {
	if dsn, ok := os.LookupEnv("DATABASE_URI"); ok {
		c.DBConfig.DSN = dsn
	}
}

func (c *Config) setJWTConfig() error {
	if key, ok := os.LookupEnv("SECRET_KEY"); ok {
		c.JWTConfig.SecretKey = key
	}
	if exp, ok := os.LookupEnv("TOKEN_EXP"); ok {
		dur, err := time.ParseDuration(exp + "s")
		if err != nil {
			return errors.New("can not parse token_exp as duration" + err.Error())
		}
		c.JWTConfig.TokenExp = dur
	}
	return nil
}

func (c *Config) envBuild() error {
	c.setEnvServerConfig()
	c.setEnvLoggerConfig()
	c.setDBConfig()
	err := c.setJWTConfig()
	if err != nil {
		return fmt.Errorf("failed set JWT config from env: %w", err)
	}
	return nil
}
