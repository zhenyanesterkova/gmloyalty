package repository

import (
	"context"
	"fmt"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository/postgres"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

type Store interface {
	Close() error
	Ping() error
	Register(ctx context.Context, user user.User) (int, error)
	Login(userData user.User) (int, error)
	GetOrderByOrderNum(orderNum string) (order.Order, error)
	AddOrder(orderData order.Order) error
	UpdateOrderStatus(orderData order.Order) error
	ProcessingOrder(ctx context.Context, orderData order.Order) error
	GetOrderList(userID int) ([]order.Order, error)
	GetUserAccaunt(userID int) (user.Accaunt, error)
	Withdraw(ctx context.Context, userID int, withdrawInst order.Withdraw) error
	Withdrawals(ctx context.Context, userID int) ([]order.Withdraw, error)
}

func NewStore(
	conf config.DBConfig,
	log logger.LogrusLogger,
	cfgJWT config.JWTConfig,
) (Store, error) {
	store, err := postgres.New(conf.DSN, log, cfgJWT)
	if err != nil {
		return nil, fmt.Errorf("failed create postgres storage: %w", err)
	}

	return store, nil
}
