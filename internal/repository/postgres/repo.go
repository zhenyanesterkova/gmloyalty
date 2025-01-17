package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
	log  logger.LogrusLogger
}

func New(dsn string, lg logger.LogrusLogger) (*PostgresStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &PostgresStorage{
		pool: pool,
		log:  lg,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (psg *PostgresStorage) Register(ctx context.Context, user user.User) error {
	log := psg.log.LogrusLog

	hashPassWD, err := user.HashPassword()
	if err != nil {
		return fmt.Errorf("failed calc hash password: %w", err)
	}

	row := psg.pool.QueryRow(
		context.TODO(),
		`INSERT INTO users (user_login, hashed_password)
			VALUES ($1, $2)
			RETURNING id;
			`,
		user.Login,
		hashPassWD,
	)

	var id string
	err = row.Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to scan row when create user: %w", err)
	}
	log.WithFields(logrus.Fields{
		"id":           id,
		"login":        user.Login,
		"hashPassword": hashPassWD,
	}).Debug("register user")

	return nil
}

func (psg *PostgresStorage) Login(user user.User) {

}

func (psg *PostgresStorage) Ping() error {
	if err := psg.pool.Ping(context.TODO()); err != nil {
		return fmt.Errorf("failed to ping the DB: %w", err)
	}

	return nil
}

func (psg *PostgresStorage) Close() error {
	psg.pool.Close()
	return nil
}
