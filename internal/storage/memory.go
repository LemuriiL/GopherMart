// Package storage - хранилища данных приложения.
package storage

import (
	"sort"
	"sync"
	"time"

	"github.com/LemuriiL/GopherMart/internal/model"
)

// MemoryStorage - простое in-memory хранилище для разработки и тестов.
type MemoryStorage struct {
	mu          sync.RWMutex
	users       map[string]*model.User
	orders      map[string]*model.Order
	withdrawals []model.Withdrawal
	nextID      int64
}

// NewMemoryStorage - создаёт новое in-memory хранилище.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:       make(map[string]*model.User),
		orders:      make(map[string]*model.Order),
		withdrawals: make([]model.Withdrawal, 0),
		nextID:      1,
	}
}

// CreateUser - сохраняет нового пользователя.
func (s *MemoryStorage) CreateUser(login, passwordHash string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[login]; ok {
		return nil, ErrUserExists
	}

	user := &model.User{
		ID:           s.nextID,
		Login:        login,
		PasswordHash: passwordHash,
	}

	s.users[login] = user
	s.nextID++

	return user, nil
}

// GetUserByLogin - возвращает пользователя по логину.
func (s *MemoryStorage) GetUserByLogin(login string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[login]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// SaveOrder - сохраняет заказ пользователя.
func (s *MemoryStorage) SaveOrder(number string, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[number]
	if ok {
		if order.UserID == userID {
			return ErrOrderUploadedBySameUser
		}
		return ErrOrderUploadedByAnotherUser
	}

	s.orders[number] = &model.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	return nil
}

// GetOrdersByUserID - возвращает заказы пользователя.
func (s *MemoryStorage) GetOrdersByUserID(userID int64) ([]model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]model.Order, 0)

	for _, order := range s.orders {
		if order.UserID == userID {
			orders = append(orders, *order)
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.After(orders[j].UploadedAt)
	})

	return orders, nil
}

// GetBalance - возвращает текущий баланс и сумму списаний пользователя.
func (s *MemoryStorage) GetBalance(userID int64) (float64, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var accrual float64
	var withdrawn float64

	for _, order := range s.orders {
		if order.UserID == userID {
			accrual += order.Accrual
		}
	}

	for _, withdrawal := range s.withdrawals {
		if withdrawal.UserID == userID {
			withdrawn += withdrawal.Sum
		}
	}

	return accrual - withdrawn, withdrawn
}

// AddWithdrawal - сохраняет факт списания средств.
func (s *MemoryStorage) AddWithdrawal(order string, userID int64, sum float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.withdrawals = append(s.withdrawals, model.Withdrawal{
		Order:       order,
		UserID:      userID,
		Sum:         sum,
		ProcessedAt: time.Now(),
	})

	return nil
}

// GetWithdrawals - возвращает историю списаний пользователя.
func (s *MemoryStorage) GetWithdrawals(userID int64) []model.Withdrawal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.Withdrawal, 0)

	for _, withdrawal := range s.withdrawals {
		if withdrawal.UserID == userID {
			result = append(result, withdrawal)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ProcessedAt.After(result[j].ProcessedAt)
	})

	return result
}

// GetNewOrders - возвращает заказы, которые ещё нужно обработать.
func (s *MemoryStorage) GetNewOrders() []model.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.Order, 0)

	for _, order := range s.orders {
		if order.Status == "NEW" || order.Status == "PROCESSING" {
			result = append(result, *order)
		}
	}

	return result
}

// UpdateOrder - обновляет статус и начисление заказа.
func (s *MemoryStorage) UpdateOrder(number string, status string, accrual float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[number]
	if !ok {
		return
	}

	order.Status = status
	order.Accrual = accrual
}
