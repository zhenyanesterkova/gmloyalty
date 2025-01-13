package config

import (
	"flag"
)

func (c *Config) setFlagsVariables() {
	flag.StringVar(
		&c.SConfig.Address,
		"a",
		c.SConfig.Address,
		"address and port to run server",
	)

	flag.StringVar(
		&c.LConfig.Level,
		"l",
		c.LConfig.Level,
		"log level",
	)

	dsn := ""
	flag.StringVar(
		&dsn,
		"d",
		dsn,
		"database dsn",
	)

	hashKey := ""
	flag.StringVar(
		&hashKey,
		"k",
		hashKey,
		"hash key",
	)

	flag.Parse()

	if isFlagPassed("d") {
		c.DBConfig.DSN = dsn
	}

	if isFlagPassed("k") {
		c.SConfig.HashKey = &hashKey
	}
}

func (c *Config) flagBuild() {
	c.setFlagsVariables()
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
