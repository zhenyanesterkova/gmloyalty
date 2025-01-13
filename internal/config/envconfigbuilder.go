package config

import (
	"os"
)

func (c *Config) setEnvServerConfig() {
	if envEndpoint, ok := os.LookupEnv("ADDRESS"); ok {
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
	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DBConfig.DSN = dsn
	}
}

func (c *Config) envBuild() {
	c.setEnvServerConfig()
	c.setEnvLoggerConfig()
	c.setDBConfig()
}
