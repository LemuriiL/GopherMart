// Package service - бизнес логика приложения.
package service

import (
	"errors"

	"github.com/LemuriiL/GopherMart/internal/luhn"
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
	if !luhn.IsDigits(number) {
		return ErrInvalidOrderNumber
	}

	if !luhn.IsValid(number) {
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
