package service

import (
	"testing"
	"time"

	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

type orderTestStore struct {
	orders     map[string]int64
	userOrders map[int64][]model.Order
}

func newOrderTestStore() *orderTestStore {
	return &orderTestStore{
		orders:     make(map[string]int64),
		userOrders: make(map[int64][]model.Order),
	}
}

func (s *orderTestStore) SaveOrder(number string, userID int64) error {
	existingUserID, ok := s.orders[number]
	if ok {
		if existingUserID == userID {
			return storage.ErrOrderUploadedBySameUser
		}
		return storage.ErrOrderUploadedByAnotherUser
	}

	s.orders[number] = userID
	s.userOrders[userID] = append(s.userOrders[userID], model.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	})
	return nil
}

func (s *orderTestStore) GetOrdersByUserID(userID int64) ([]model.Order, error) {
	return s.userOrders[userID], nil
}

func TestOrderServiceUploadOrderSuccess(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrderServiceUploadOrderInvalidNumber(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("12345", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrInvalidOrderNumber {
		t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestOrderServiceUploadOrderNonDigits(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("12ab34", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrInvalidOrderNumber {
		t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestOrderServiceUploadOrderEmpty(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrInvalidOrderNumber {
		t.Fatalf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestOrderServiceUploadOrderSameUser(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = service.UploadOrder("79927398713", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != storage.ErrOrderUploadedBySameUser {
		t.Fatalf("expected ErrOrderUploadedBySameUser, got %v", err)
	}
}

func TestOrderServiceUploadOrderAnotherUser(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = service.UploadOrder("79927398713", 2)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != storage.ErrOrderUploadedByAnotherUser {
		t.Fatalf("expected ErrOrderUploadedByAnotherUser, got %v", err)
	}
}

func TestOrderServiceGetUserOrders(t *testing.T) {
	store := newOrderTestStore()
	service := NewOrderService(store)

	err := service.UploadOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = service.UploadOrder("12345678903", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orders, err := service.GetUserOrders(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
}
