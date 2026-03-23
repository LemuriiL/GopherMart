package storage

import (
	"testing"

	"github.com/LemuriiL/GopherMart/internal/model"
)

func TestMemoryStorageCreateUser(t *testing.T) {
	s := NewMemoryStorage()

	user, err := s.CreateUser("user1", "hash1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Login != "user1" {
		t.Fatalf("expected login user1, got %s", user.Login)
	}

	if user.ID == 0 {
		t.Fatal("expected non-zero user id")
	}
}

func TestMemoryStorageCreateUserDuplicate(t *testing.T) {
	s := NewMemoryStorage()

	_, err := s.CreateUser("user1", "hash1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.CreateUser("user1", "hash2")
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestMemoryStorageGetUserByLogin(t *testing.T) {
	s := NewMemoryStorage()

	_, err := s.CreateUser("user1", "hash1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, err := s.GetUserByLogin("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Login != "user1" {
		t.Fatalf("expected login user1, got %s", user.Login)
	}
}

func TestMemoryStorageGetUserByLoginNotFound(t *testing.T) {
	s := NewMemoryStorage()

	_, err := s.GetUserByLogin("missing")
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryStorageSaveOrder(t *testing.T) {
	s := NewMemoryStorage()

	err := s.SaveOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orders, err := s.GetOrdersByUserID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}

	if orders[0].Number != "79927398713" {
		t.Fatalf("expected number 79927398713, got %s", orders[0].Number)
	}
}

func TestMemoryStorageSaveOrderSameUser(t *testing.T) {
	s := NewMemoryStorage()

	err := s.SaveOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.SaveOrder("79927398713", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrOrderUploadedBySameUser {
		t.Fatalf("expected ErrOrderUploadedBySameUser, got %v", err)
	}
}

func TestMemoryStorageSaveOrderAnotherUser(t *testing.T) {
	s := NewMemoryStorage()

	err := s.SaveOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.SaveOrder("79927398713", 2)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrOrderUploadedByAnotherUser {
		t.Fatalf("expected ErrOrderUploadedByAnotherUser, got %v", err)
	}
}

func TestMemoryStorageGetOrdersByUserID(t *testing.T) {
	s := NewMemoryStorage()

	err := s.SaveOrder("79927398713", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.SaveOrder("12345678903", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.SaveOrder("4242424242424242", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orders, err := s.GetOrdersByUserID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
}

func TestMemoryStorageGetBalance(t *testing.T) {
	s := NewMemoryStorage()

	s.orders["1"] = &model.Order{Number: "1", UserID: 1, Accrual: 100}
	s.orders["2"] = &model.Order{Number: "2", UserID: 1, Accrual: 25}
	s.withdrawals = append(s.withdrawals,
		model.Withdrawal{Order: "w1", UserID: 1, Sum: 40},
		model.Withdrawal{Order: "w2", UserID: 2, Sum: 999},
	)

	current, withdrawn := s.GetBalance(1)

	if current != 85 {
		t.Fatalf("expected current 85, got %v", current)
	}

	if withdrawn != 40 {
		t.Fatalf("expected withdrawn 40, got %v", withdrawn)
	}
}

func TestMemoryStorageAddWithdrawalAndGetWithdrawals(t *testing.T) {
	s := NewMemoryStorage()

	err := s.AddWithdrawal("12345678903", 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddWithdrawal("79927398713", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddWithdrawal("4242424242424242", 2, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	withdrawals := s.GetWithdrawals(1)

	if len(withdrawals) != 2 {
		t.Fatalf("expected 2 withdrawals, got %d", len(withdrawals))
	}
}

func TestMemoryStorageGetNewOrders(t *testing.T) {
	s := NewMemoryStorage()

	s.orders["1"] = &model.Order{Number: "1", UserID: 1, Status: "NEW"}
	s.orders["2"] = &model.Order{Number: "2", UserID: 1, Status: "PROCESSING"}
	s.orders["3"] = &model.Order{Number: "3", UserID: 1, Status: "PROCESSED"}
	s.orders["4"] = &model.Order{Number: "4", UserID: 1, Status: "INVALID"}

	orders := s.GetNewOrders()

	if len(orders) != 2 {
		t.Fatalf("expected 2 new/process orders, got %d", len(orders))
	}
}

func TestMemoryStorageUpdateOrder(t *testing.T) {
	s := NewMemoryStorage()

	s.orders["79927398713"] = &model.Order{
		Number:  "79927398713",
		UserID:  1,
		Status:  "NEW",
		Accrual: 0,
	}

	s.UpdateOrder("79927398713", "PROCESSED", 500)

	order := s.orders["79927398713"]
	if order.Status != "PROCESSED" {
		t.Fatalf("expected status PROCESSED, got %s", order.Status)
	}

	if order.Accrual != 500 {
		t.Fatalf("expected accrual 500, got %v", order.Accrual)
	}
}

func TestMemoryStorageUpdateOrderMissing(t *testing.T) {
	s := NewMemoryStorage()

	s.UpdateOrder("missing", "PROCESSED", 500)

	if len(s.orders) != 0 {
		t.Fatalf("expected 0 orders, got %d", len(s.orders))
	}
}
