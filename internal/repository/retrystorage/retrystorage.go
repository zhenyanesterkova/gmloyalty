package retrystorage

import (
	"fmt"
	"time"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/backoff"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
)

type RetryStorage struct {
	storage    repository.Store
	backoff    *backoff.Backoff
	logger     logger.LogrusLogger
	checkRetry func(error) bool
}

func New(
	cfg config.DBConfig,
	loggerInst logger.LogrusLogger,
	bf *backoff.Backoff,
	checkRetryFunc func(error) bool,
) (
	*RetryStorage,
	error,
) {
	retryStore := &RetryStorage{
		checkRetry: checkRetryFunc,
		backoff:    bf,
		logger:     loggerInst,
	}

	store, err := repository.NewStore(cfg, loggerInst)
	if err != nil {
		if retryStore.checkRetry(err) {
			err = retryStore.retry(func() error {
				store, err = repository.NewStore(cfg, loggerInst)
				if err != nil {
					return fmt.Errorf("failed retry create storage: %w", err)
				}
				return nil
			})
		}
		loggerInst.LogrusLog.Errorf("can not create storage: %v", err)
		return retryStore, fmt.Errorf("can not create storage: %w", err)
	}

	retryStore.storage = store
	return retryStore, nil
}

func (rs *RetryStorage) Ping() error {
	err := rs.storage.Ping()
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			err = rs.storage.Ping()
			if err != nil {
				return fmt.Errorf("failed retry ping: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("failed ping: %w", err)
	}
	return nil
}

func (rs *RetryStorage) Close() error {
	if err := rs.storage.Close(); err != nil {
		return fmt.Errorf("failed close DB: %w", err)
	}
	return nil
}

func (rs *RetryStorage) retry(work func() error) error {
	log := rs.logger.LogrusLog
	defer rs.backoff.Reset()
	for {
		log.Debug("attempt to repeat ...")
		err := work()

		if err == nil {
			return nil
		}

		if rs.checkRetry(err) {
			var delay time.Duration
			if delay = rs.backoff.Next(); delay == backoff.Stop {
				return err
			}
			time.Sleep(delay)
		}
	}
}
