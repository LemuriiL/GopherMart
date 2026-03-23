// Package service - бизнес логика приложения.
package service

import (
	"errors"

	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

// ErrInvalidOrderNumber - неверный номер заказа.
var ErrInvalidOrderNumber = errors.New("invalid order number")

// OrderSaver - интерфейс для сохранения заказа.
type OrderSaver interface {
	SaveOrder(number string, userID int64) error
}

// OrderReader - интерфейс для получения заказов пользователя.
type OrderReader interface {
	GetOrdersByUserID(userID int64) ([]model.Order, error)
}

// OrderService - отвечает за работу с заказами.
type OrderService struct {
	store interface {
		OrderSaver
		OrderReader
	}
}

// NewOrderService - создаёт новый OrderService.
func NewOrderService(store interface {
	OrderSaver
	OrderReader
}) *OrderService {
	return &OrderService{store: store}
}

// UploadOrder - принимает и валидирует номер заказа.
func (s *OrderService) UploadOrder(number string, userID int64) error {
	if !isDigits(number) {
		return ErrInvalidOrderNumber
	}

	if !isValidLuhn(number) {
		return ErrInvalidOrderNumber
	}

	err := s.store.SaveOrder(number, userID)
	if err != nil {
		if errors.Is(err, storage.ErrOrderUploadedBySameUser) {
			return storage.ErrOrderUploadedBySameUser
		}
		if errors.Is(err, storage.ErrOrderUploadedByAnotherUser) {
			return storage.ErrOrderUploadedByAnotherUser
		}
		return err
	}

	return nil
}

// GetUserOrders - возвращает заказы пользователя.
func (s *OrderService) GetUserOrders(userID int64) ([]model.Order, error) {
	return s.store.GetOrdersByUserID(userID)
}

// isDigits - проверяет, что строка состоит только из цифр.
func isDigits(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// isValidLuhn - проверяет номер по алгоритму Луна.
func isValidLuhn(number string) bool {
	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
