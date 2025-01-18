package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"
)

func (c *Config) setFlagsVariables() error {
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

	secretKey := ""
	flag.StringVar(
		&secretKey,
		"s",
		DefaultSecretKey,
		"secret key for jwt",
	)

	var durTokenExp int
	flag.IntVar(&durTokenExp, "e", DefaultTokenExp, "token exp, hour")

	flag.Parse()

	if isFlagPassed("d") {
		c.DBConfig.DSN = dsn
	}

	if isFlagPassed("k") {
		c.SConfig.HashKey = &hashKey
	}

	if isFlagPassed("s") {
		c.JWTConfig.SecretKey = secretKey
	}

	if isFlagPassed("e") {
		dur, err := time.ParseDuration(strconv.Itoa(durTokenExp) + "h")
		if err != nil {
			return errors.New("can not parse token exp as duration " + err.Error())
		}
		c.JWTConfig.TokenExp = dur
	}

	return nil
}

func (c *Config) flagBuild() error {
	err := c.setFlagsVariables()
	if err != nil {
		return fmt.Errorf("failed set flags config: %w", err)
	}

	return nil
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
