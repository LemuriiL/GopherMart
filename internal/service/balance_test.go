package service

import (
	"testing"
	"time"

	"github.com/LemuriiL/GopherMart/internal/model"
)

type balanceTestStore struct {
	current     float64
	withdrawn   float64
	withdrawals []model.Withdrawal
}

func newBalanceTestStore(current, withdrawn float64) *balanceTestStore {
	return &balanceTestStore{
		current:     current,
		withdrawn:   withdrawn,
		withdrawals: make([]model.Withdrawal, 0),
	}
}

func (s *balanceTestStore) GetBalance(userID int64) (float64, float64) {
	return s.current, s.withdrawn
}

func (s *balanceTestStore) AddWithdrawal(order string, userID int64, sum float64) error {
	s.current -= sum
	s.withdrawn += sum
	s.withdrawals = append(s.withdrawals, model.Withdrawal{
		Order:       order,
		UserID:      userID,
		Sum:         sum,
		ProcessedAt: time.Now(),
	})
	return nil
}

func (s *balanceTestStore) GetWithdrawals(userID int64) []model.Withdrawal {
	result := make([]model.Withdrawal, 0)
	for _, w := range s.withdrawals {
		if w.UserID == userID {
			result = append(result, w)
		}
	}
	return result
}

func TestBalanceServiceGetBalance(t *testing.T) {
	store := newBalanceTestStore(100, 25)
	service := NewBalanceService(store)

	current, withdrawn := service.GetBalance(1)

	if current != 100 {
		t.Fatalf("expected current 100, got %v", current)
	}

	if withdrawn != 25 {
		t.Fatalf("expected withdrawn 25, got %v", withdrawn)
	}
}

func TestBalanceServiceWithdrawSuccess(t *testing.T) {
	store := newBalanceTestStore(100, 0)
	service := NewBalanceService(store)

	err := service.Withdraw(1, "12345678903", 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	current, withdrawn := service.GetBalance(1)

	if current != 60 {
		t.Fatalf("expected current 60, got %v", current)
	}

	if withdrawn != 40 {
		t.Fatalf("expected withdrawn 40, got %v", withdrawn)
	}

	withdrawals := service.GetWithdrawals(1)
	if len(withdrawals) != 1 {
		t.Fatalf("expected 1 withdrawal, got %d", len(withdrawals))
	}

	if withdrawals[0].Order != "12345678903" {
		t.Fatalf("expected order 12345678903, got %s", withdrawals[0].Order)
	}
}

func TestBalanceServiceWithdrawExactBalance(t *testing.T) {
	store := newBalanceTestStore(50, 0)
	service := NewBalanceService(store)

	err := service.Withdraw(1, "79927398713", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	current, withdrawn := service.GetBalance(1)

	if current != 0 {
		t.Fatalf("expected current 0, got %v", current)
	}

	if withdrawn != 50 {
		t.Fatalf("expected withdrawn 50, got %v", withdrawn)
	}
}

func TestBalanceServiceWithdrawNotEnoughBalance(t *testing.T) {
	store := newBalanceTestStore(30, 0)
	service := NewBalanceService(store)

	err := service.Withdraw(1, "12345678903", 40)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrNotEnoughBalance {
		t.Fatalf("expected ErrNotEnoughBalance, got %v", err)
	}
}

func TestBalanceServiceGetWithdrawals(t *testing.T) {
	store := newBalanceTestStore(100, 0)
	service := NewBalanceService(store)

	err := service.Withdraw(1, "12345678903", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = service.Withdraw(1, "79927398713", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	withdrawals := service.GetWithdrawals(1)
	if len(withdrawals) != 2 {
		t.Fatalf("expected 2 withdrawals, got %d", len(withdrawals))
	}
}

func TestBalanceServiceGetWithdrawalsOtherUser(t *testing.T) {
	store := newBalanceTestStore(100, 0)
	service := NewBalanceService(store)

	err := service.Withdraw(1, "12345678903", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	withdrawals := service.GetWithdrawals(2)
	if len(withdrawals) != 0 {
		t.Fatalf("expected 0 withdrawals, got %d", len(withdrawals))
	}
}
