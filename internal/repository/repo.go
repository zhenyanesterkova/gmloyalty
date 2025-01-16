package repository

import (
	"fmt"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository/postgres"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
)

type Store interface {
	Close() error
	Ping() error
}

func NewStore(conf config.DBConfig, log logger.LogrusLogger) (Store, error) {
	store, err := postgres.New(conf.DSN, log)
	if err != nil {
		return nil, fmt.Errorf("failed create postgres storage: %w", err)
	}

	return store, nil
}
