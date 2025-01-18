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

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
	log  logger.LogrusLogger
}

func New(
	dsn string,
	lg logger.LogrusLogger,
	cfgJWT config.JWTConfig,
) (*PostgresStorage, error) {
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

func (psg *PostgresStorage) Register(ctx context.Context, userData user.User) (int, error) {
	salt, err := user.CreateSalt()
	if err != nil {
		return 0, fmt.Errorf("failed generate salt for calc hash password: %w", err)
	}

	hashPassWD, err := userData.HashPassword(salt)
	if err != nil {
		return 0, fmt.Errorf("failed calc hash password: %w", err)
	}

	row := psg.pool.QueryRow(
		context.TODO(),
		`INSERT INTO users (user_login, hashed_password)
			VALUES ($1, $2)
			RETURNING id;
			`,
		userData.Login,
		hashPassWD,
	)

	var id int
	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row when create user: %w", err)
	}

	return id, nil
}

func (psg *PostgresStorage) Login(userData user.User) (int, error) {
	row := psg.pool.QueryRow(
		context.TODO(),
		`SELECT id, hashed_password FROM users 
			WHERE user_login = $1;
		`,
		userData.Login,
	)

	var passWD string
	var userID int
	err := row.Scan(&userID, &passWD)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row when login user: %w", err)
	}

	err = userData.CheckPassword(passWD)
	if err != nil {
		return 0, fmt.Errorf("failed check password: %w", err)
	}

	return userID, nil
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
