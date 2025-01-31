package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
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
	log := psg.log.LogrusLog

	salt, err := user.CreateSalt()
	if err != nil {
		return 0, fmt.Errorf("failed generate salt for calc hash password: %w", err)
	}

	hashPassWD, err := userData.HashPassword(salt)
	if err != nil {
		return 0, fmt.Errorf("failed calc hash password: %w", err)
	}

	tx, err := psg.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed start a register transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				log.Errorf("failed rolls back the register transaction: %v", err)
			}
		}
	}()

	row := tx.QueryRow(
		ctx,
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

	row = tx.QueryRow(
		ctx,
		`INSERT INTO accounts (user_id, balance, withdrawn)
			VALUES ($1, 0, 0)
			RETURNING id;
			`,
		id,
	)

	var idAccaunts int
	err = row.Scan(&idAccaunts)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row when create user accaunt: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed commits the transaction register user: %w", err)
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

func (psg *PostgresStorage) GetUserAccaunt(userID int) (user.Accaunt, error) {
	row := psg.pool.QueryRow(
		context.TODO(),
		`SELECT id, balance, withdrawn FROM accounts 
			WHERE user_id = $1;
		`,
		userID,
	)

	acc := user.Accaunt{}
	acc.UserID = userID
	err := row.Scan(&acc.ID, &acc.Balance, &acc.Withdrawn)
	if err != nil {
		return user.Accaunt{}, fmt.Errorf("failed to scan row when get user accaunt by userID: %w", err)
	}

	return acc, nil
}

func (psg *PostgresStorage) GetOrderByOrderNum(orderNum string) (order.Order, error) {
	row := psg.pool.QueryRow(
		context.TODO(),
		`SELECT order_status, upload_time, user_id FROM orders 
			WHERE order_num = $1;
		`,
		orderNum,
	)

	var (
		userID      int
		orderStatus string
		uploadTime  time.Time
	)
	err := row.Scan(&orderStatus, &uploadTime, &userID)
	if err != nil {
		return order.Order{}, fmt.Errorf("failed to scan row when get user by order num: %w", err)
	}

	return order.Order{
		Number:     orderNum,
		UploadTime: uploadTime,
		Status:     orderStatus,
		UserID:     userID,
	}, nil
}

func (psg *PostgresStorage) AddOrder(orderData order.Order) error {
	_, err := psg.pool.Exec(
		context.TODO(),
		`INSERT INTO orders (order_num, user_id, order_status)
			VALUES ($1, $2, $3);`,
		orderData.Number,
		orderData.UserID,
		orderData.Status,
	)
	if err != nil {
		return fmt.Errorf("failed add order to orders: %w", err)
	}
	return nil
}

func (psg *PostgresStorage) UpdateOrderStatus(orderData order.Order) error {
	_, err := psg.pool.Exec(
		context.TODO(),
		`UPDATE orders SET
			order_status = $1
		WHERE 
			order_num = $2;`,
		orderData.Status,
		orderData.Number,
	)
	if err != nil {
		return fmt.Errorf("failed update order in orders: %w", err)
	}
	return nil
}

func (psg *PostgresStorage) ProcessingOrder(ctx context.Context, orderData order.Order) error {
	log := psg.log.LogrusLog

	log.Debug("start save info about order to DB ...")

	log.WithFields(logrus.Fields{
		"Number":  orderData.Number,
		"Status":  orderData.Status,
		"Accrual": orderData.Accrual,
	}).Info("set order info")

	accaunt, err := psg.GetUserAccaunt(orderData.UserID)
	if err != nil {
		return fmt.Errorf("failed get accaunt to process order: %w", err)
	}

	tx, err := psg.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed start a processing order transaction: %w", err)
	}

	defer func() {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				log.Errorf("failed rolls back the processing order transaction: %v", errRollback)
			}
		}
	}()

	_, err = tx.Exec(
		ctx,
		`INSERT INTO history (order_num, item_type, sum) 
		VALUES ($1, $2, $3);`,
		orderData.Number,
		"accrual",
		orderData.Accrual,
	)
	if err != nil {
		return fmt.Errorf("failed exec query add history item in processing order transaction: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE accounts SET
			balance = $1
		WHERE 
			id = $2;`,
		accaunt.Balance+orderData.Accrual,
		accaunt.ID,
	)
	if err != nil {
		return fmt.Errorf("failed exec query update user accaut in processing order transaction: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE orders SET
			order_status = $1
		WHERE 
			order_num = $2;`,
		orderData.Status,
		orderData.Number,
	)
	if err != nil {
		return fmt.Errorf("failed update order in orders in processing order transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed commits the transaction processing order: %w", err)
	}
	return nil
}

func (psg *PostgresStorage) GetOrderList(userID int) ([]order.Order, error) {
	rows, err := psg.pool.Query(
		context.TODO(),
		`SELECT 
			orders.order_num, 
			orders.order_status, 
			orders.upload_time, 
			orders.user_id, 
			history.sum
		FROM orders
		LEFT JOIN history
		ON orders.order_num = history.order_num AND history.item_type != 'withdrawn'
		WHERE orders.user_id = $1 
		ORDER BY orders.upload_time DESC;
		`,
		userID,
	)
	if err != nil {
		return []order.Order{}, fmt.Errorf("failed query get orders list: %w", err)
	}
	defer rows.Close()

	orderList := []order.Order{}
	var (
		orderNum     string
		orderStatus  string
		uploadTime   time.Time
		userIDFromDB int
		sum          sql.NullFloat64
	)
	for rows.Next() {
		err := rows.Scan(
			&orderNum,
			&orderStatus,
			&uploadTime,
			&userIDFromDB,
			&sum,
		)
		if err != nil {
			return []order.Order{}, fmt.Errorf("failed scan rows when get orders list: %w", err)
		}
		orderList = append(orderList, order.Order{
			Number:     orderNum,
			Status:     orderStatus,
			UploadTime: uploadTime,
			UserID:     userIDFromDB,
			Accrual:    sum.Float64,
		})
	}

	return orderList, nil
}

func (psg *PostgresStorage) Withdraw(ctx context.Context, userID int, withdrawInst order.Withdraw) error {
	log := psg.log.LogrusLog

	accaunt, err := psg.GetUserAccaunt(userID)
	if err != nil {
		return fmt.Errorf("failed get accaunt to add withdraw: %w", err)
	}

	tx, err := psg.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed start add withdraw transaction: %w", err)
	}

	defer func() {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				log.Errorf("failed rolls back add withdraw transaction: %v", errRollback)
			}
		}
	}()

	_, err = tx.Exec(
		ctx,
		`INSERT INTO orders (order_num, user_id, order_status)
			VALUES ($1, $2, $3);`,
		withdrawInst.Number,
		userID,
		order.StatusNew,
	)
	if err != nil {
		return fmt.Errorf("failed add order to orders in add withdraw transaction: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO history (order_num, item_type, sum) 
		VALUES ($1, $2, $3);`,
		withdrawInst.Number,
		"withdrawn",
		withdrawInst.Sum,
	)
	if err != nil {
		return fmt.Errorf("failed exec query add history item in add withdraw transaction: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE accounts SET
			balance = $1,
			withdrawn = $2
		WHERE 
			id = $3;`,
		accaunt.Balance-withdrawInst.Sum,
		accaunt.Withdrawn+withdrawInst.Sum,
		accaunt.ID,
	)
	if err != nil {
		return fmt.Errorf("failed exec query update user accaunt in add withdraw transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed commits the transaction add withdraw: %w", err)
	}
	return nil
}

func (psg *PostgresStorage) Withdrawals(ctx context.Context, userID int) ([]order.Withdraw, error) {
	rows, err := psg.pool.Query(
		ctx,
		`SELECT 
			history.order_num, 
			history.sum,
			history.item_timestamp
		FROM history
		INNER JOIN orders
		ON orders.order_num = history.order_num AND history.item_type = 'withdrawn'
		WHERE orders.user_id = $1
		ORDER BY orders.upload_time DESC;
		`,
		userID,
	)
	if err != nil {
		return []order.Withdraw{}, fmt.Errorf("failed query get withdrawals: %w", err)
	}
	defer rows.Close()

	withdrawals := []order.Withdraw{}
	var (
		orderNum   string
		uploadTime time.Time
		sum        sql.NullFloat64
	)
	for rows.Next() {
		err := rows.Scan(
			&orderNum,
			&sum,
			&uploadTime,
		)
		if err != nil {
			return []order.Withdraw{}, fmt.Errorf("failed scan rows when get withdrawals: %w", err)
		}
		withdrawals = append(withdrawals, order.Withdraw{
			Number:    orderNum,
			Timestamp: uploadTime,
			Sum:       sum.Float64,
		})
	}

	return withdrawals, nil
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
