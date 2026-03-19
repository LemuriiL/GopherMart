package storage

import (
	"database/sql"
	"errors"

	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/lib/pq"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrOrderUploadedBySameUser = errors.New("order already uploaded by same user")
var ErrOrderUploadedByAnotherUser = errors.New("order already uploaded by another user")

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Init() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS orders (
			number TEXT PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id),
			status TEXT NOT NULL,
			accrual DOUBLE PRECISION NOT NULL DEFAULT 0,
			uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS withdrawals (
			id BIGSERIAL PRIMARY KEY,
			order_number TEXT NOT NULL,
			user_id BIGINT NOT NULL REFERENCES users(id),
			sum DOUBLE PRECISION NOT NULL,
			processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) CreateUser(login, passwordHash string) (*model.User, error) {
	user := &model.User{}

	err := s.db.QueryRow(
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash`,
		login,
		passwordHash,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUserExists
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByLogin(login string) (*model.User, error) {
	user := &model.User{}

	err := s.db.QueryRow(
		`SELECT id, login, password_hash FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) SaveOrder(number string, userID int64) error {
	var existingUserID int64

	err := s.db.QueryRow(`SELECT user_id FROM orders WHERE number = $1`, number).Scan(&existingUserID)
	if err == nil {
		if existingUserID == userID {
			return ErrOrderUploadedBySameUser
		}
		return ErrOrderUploadedByAnotherUser
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, NOW())`,
		number,
		userID,
		"NEW",
		0,
	)
	return err
}

func (s *PostgresStorage) GetOrdersByUserID(userID int64) ([]model.Order, error) {
	rows, err := s.db.Query(
		`SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *PostgresStorage) GetBalance(userID int64) (float64, float64) {
	var accrual float64
	var withdrawn float64

	_ = s.db.QueryRow(`SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1`, userID).Scan(&accrual)
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`, userID).Scan(&withdrawn)

	return accrual - withdrawn, withdrawn
}

func (s *PostgresStorage) AddWithdrawal(order string, userID int64, sum float64) error {
	_, err := s.db.Exec(
		`INSERT INTO withdrawals (order_number, user_id, sum, processed_at) VALUES ($1, $2, $3, NOW())`,
		order,
		userID,
		sum,
	)
	return err
}

func (s *PostgresStorage) GetWithdrawals(userID int64) []model.Withdrawal {
	rows, err := s.db.Query(
		`SELECT order_number, user_id, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return []model.Withdrawal{}
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var withdrawal model.Withdrawal
		if err := rows.Scan(&withdrawal.Order, &withdrawal.UserID, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return []model.Withdrawal{}
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals
}

func (s *PostgresStorage) GetNewOrders() []model.Order {
	rows, err := s.db.Query(
		`SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE status IN ('NEW', 'PROCESSING') ORDER BY uploaded_at ASC`,
	)
	if err != nil {
		return []model.Order{}
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return []model.Order{}
		}
		orders = append(orders, order)
	}

	return orders
}

func (s *PostgresStorage) UpdateOrder(number string, status string, accrual float64) {
	_, _ = s.db.Exec(
		`UPDATE orders SET status = $2, accrual = $3 WHERE number = $1`,
		number,
		status,
		accrual,
	)
}
