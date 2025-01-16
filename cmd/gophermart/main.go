package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
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

	retryStore, err := retrystorage.New(cfg.DBConfig, loggerInst, backoffInst, checkRetryFunc)
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

	repoHandler := handler.NewRepositorieHandler(retryStore, loggerInst, cfg.SConfig.HashKey)
	repoHandler.InitChiRouter(router)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	loggerInst.LogrusLog.Infof("Start Server on %s", cfg.SConfig.Address)
	go func() {
		if err := http.ListenAndServe(cfg.SConfig.Address, router); err != nil {
			loggerInst.LogrusLog.Errorf("server error: %v", err)
		}
	}()

	s := <-c
	loggerInst.LogrusLog.Info("Got signal: ", s)

	return nil
}
