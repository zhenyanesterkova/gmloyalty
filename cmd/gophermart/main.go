package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/handler"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository/retrystorage"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/backoff"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run() error {
	cfg := config.New()
	err := cfg.Build()
	if err != nil {
		log.Printf("can not build config: %v", err)
		return fmt.Errorf("config error: %w", err)
	}

	loggerInst := logger.NewLogrusLogger()
	err = loggerInst.SetLevelForLog(cfg.LConfig.Level)
	if err != nil {
		loggerInst.LogrusLog.Errorf("can not parse log level: %v", err)
		return fmt.Errorf("parse log level error: %w", err)
	}

	backoffInst := backoff.New(
		cfg.RetryConfig.MinDelay,
		cfg.RetryConfig.MaxDelay,
		cfg.RetryConfig.MaxAttempt,
	)

	checkRetryFunc := func(err error) bool {
		var pgErr *pgconn.PgError
		var pgErrConn *pgconn.ConnectError
		res := false
		if errors.As(err, &pgErr) {
			res = pgerrcode.IsConnectionException(pgErr.Code)
		} else if errors.As(err, &pgErrConn) {
			res = true
		}
		return res
	}

	retryStore, err := retrystorage.New(
		cfg.DBConfig,
		loggerInst,
		backoffInst,
		checkRetryFunc,
		cfg.JWTConfig,
	)
	if err != nil {
		loggerInst.LogrusLog.Errorf("failed create storage: %v", err)
		return fmt.Errorf("failed create storage: %w", err)
	}
	defer func() {
		err := retryStore.Close()
		if err != nil {
			loggerInst.LogrusLog.Errorf("can not close storage: %v", err)
		}
	}()

	router := chi.NewRouter()

	repoHandler := handler.NewRepositorieHandler(retryStore, loggerInst, cfg.SConfig.HashKey, cfg.JWTConfig)
	repoHandler.InitChiRouter(router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	serverCtx := context.WithoutCancel(ctx)
	errCh := make(chan error)

	loggerInst.LogrusLog.Infof("Start Server on %s", cfg.SConfig.Address)
	go func(ctx context.Context, errCh chan error) {
		defer close(errCh)

		select {
		case errCh <- http.ListenAndServe(cfg.SConfig.Address, router):
			return
		case <-ctx.Done():
			return
		}
	}(serverCtx, errCh)

	select {
	case <-ctx.Done():
		log.Println("Got stop signal")
	case err := <-errCh:
		stop()
		log.Printf("fatal error: %v", err)
	}

	return nil
}
