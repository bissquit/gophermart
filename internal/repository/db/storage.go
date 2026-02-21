package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bissquit/gophermart/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewDBStorage(p *pgxpool.Pool, l *slog.Logger) *PGStorage {
	return &PGStorage{
		pool:   p,
		logger: l,
	}
}

func (s *PGStorage) CreateUser(login, passwordHash string) (userID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = s.pool.QueryRow(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		login, passwordHash,
	).Scan(&userID)

	if err == nil {
		return userID, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return "", fmt.Errorf("%w: %s", repository.ErrUserAlreadyExists, login)
	}

	s.logger.Error("create user error", "err", err)
	return "", err
}

func (s *PGStorage) GetUserByLogin(login string) (user repository.User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = s.pool.QueryRow(ctx,
		"SELECT id, login, password_hash FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)

	if err == nil {
		return user, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.User{}, repository.ErrUserNotFound
	}

	s.logger.Error("get user by login error", "err", err, "login", login)
	return repository.User{}, err
}

func (s *PGStorage) CreateUserOrder(userID, orderNumber string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		"INSERT INTO orders (user_id, order_number) VALUES ($1, $2)",
		userID, orderNumber,
	)

	if err == nil {
		return nil // 202
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		var existingUserID string
		err = s.pool.QueryRow(ctx,
			"SELECT user_id FROM orders WHERE order_number = $1",
			orderNumber,
		).Scan(&existingUserID)

		if err != nil {
			s.logger.Error("failed to check order owner", "err", err)
			return err
		}

		if existingUserID == userID {
			return repository.ErrOrderAlreadyCreatedByUser // 200
		} else {
			return repository.ErrOrderAlreadyCreatedByAnotherUser // 409
		}
	}

	s.logger.Error("create order error", "err", err)
	return err
}

func (s *PGStorage) GetUserOrders(UserID string) (orders []repository.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, order_number, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC
	`, UserID)
	if err != nil {
		s.logger.Error("query orders error", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order repository.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderNumber,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			s.logger.Error("scan order error", "err", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("rows iteration error", "err", err)
		return nil, err
	}

	return orders, nil
}

func (s *PGStorage) GetUserBalance(userID string) (current, withdrawn float64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = s.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'), 0)
			-
			COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0) as current,
			
			COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0) as withdrawn
		`,
		userID,
	).Scan(&current, &withdrawn)

	if err != nil {
		s.logger.Error("get user balance error", "err", err)
		return 0, 0, err
	}
	return current, withdrawn, nil
}
