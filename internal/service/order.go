package service

import (
	"errors"

	"github.com/LemuriiL/GopherMart/internal/storage"
)

var ErrInvalidOrderNumber = errors.New("invalid order number")

type OrderSaver interface {
	SaveOrder(number string, userID int64) error
}

type OrderService struct {
	store OrderSaver
}

func NewOrderService(store OrderSaver) *OrderService {
	return &OrderService{store: store}
}

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
