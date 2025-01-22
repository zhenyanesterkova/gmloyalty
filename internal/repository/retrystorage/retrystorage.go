package retrystorage

import (
	"context"
	"fmt"
	"time"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/backoff"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
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
	cfgJWT config.JWTConfig,
) (
	*RetryStorage,
	error,
) {
	retryStore := &RetryStorage{
		checkRetry: checkRetryFunc,
		backoff:    bf,
		logger:     loggerInst,
	}

	store, err := repository.NewStore(cfg, loggerInst, cfgJWT)
	if err != nil {
		if retryStore.checkRetry(err) {
			err = retryStore.retry(func() error {
				store, err = repository.NewStore(cfg, loggerInst, cfgJWT)
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

func (rs *RetryStorage) Register(ctx context.Context, user user.User) (int, error) {
	userID, err := rs.storage.Register(ctx, user)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			userID, err = rs.storage.Register(ctx, user)
			if err != nil {
				return fmt.Errorf("failed retry register user: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return 0, fmt.Errorf("failed register user: %w", err)
	}
	return userID, nil
}

func (rs *RetryStorage) Login(user user.User) (int, error) {
	userID, err := rs.storage.Login(user)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			userID, err = rs.storage.Login(user)
			if err != nil {
				return fmt.Errorf("failed retry login user: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return 0, fmt.Errorf("failed login user: %w", err)
	}
	return userID, nil
}

func (rs *RetryStorage) GetOrderByOrderNum(orderNum string) (order.Order, error) {
	orderData, err := rs.storage.GetOrderByOrderNum(orderNum)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			orderData, err = rs.storage.GetOrderByOrderNum(orderNum)
			if err != nil {
				return fmt.Errorf("failed retry get order by order num: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return order.Order{}, fmt.Errorf("failed get order by order num: %w", err)
	}
	return orderData, nil
}

func (rs *RetryStorage) AddOrder(orderData order.Order) error {
	err := rs.storage.AddOrder(orderData)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			err = rs.storage.AddOrder(orderData)
			if err != nil {
				return fmt.Errorf("failed retry add order to orders: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("failed add order to orders: %w", err)
	}
	return nil
}

func (rs *RetryStorage) UpdateOrderStatus(orderData order.Order) error {
	err := rs.storage.UpdateOrderStatus(orderData)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			err = rs.storage.UpdateOrderStatus(orderData)
			if err != nil {
				return fmt.Errorf("failed retry add order to orders: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("failed add order to orders: %w", err)
	}
	return nil
}

func (rs *RetryStorage) ProcessingOrder(ctx context.Context, orderData order.Order) error {
	err := rs.storage.ProcessingOrder(ctx, orderData)
	if rs.checkRetry(err) {
		err = rs.retry(func() error {
			err = rs.storage.ProcessingOrder(ctx, orderData)
			if err != nil {
				return fmt.Errorf("failed retry processing order: %w", err)
			}
			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("failed processing order: %w", err)
	}
	return nil
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
