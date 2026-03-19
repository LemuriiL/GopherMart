package service

import (
	"errors"

	"github.com/LemuriiL/GopherMart/internal/model"
)

var ErrNotEnoughBalance = errors.New("not enough balance")

type BalanceStorage interface {
	GetBalance(userID int64) (float64, float64)
	AddWithdrawal(order string, userID int64, sum float64) error
	GetWithdrawals(userID int64) []model.Withdrawal
}

type BalanceService struct {
	store BalanceStorage
}

func NewBalanceService(store BalanceStorage) *BalanceService {
	return &BalanceService{store: store}
}

func (s *BalanceService) GetBalance(userID int64) (float64, float64) {
	return s.store.GetBalance(userID)
}

func (s *BalanceService) Withdraw(userID int64, order string, sum float64) error {
	current, _ := s.store.GetBalance(userID)

	if sum > current {
		return ErrNotEnoughBalance
	}

	return s.store.AddWithdrawal(order, userID, sum)
}

func (s *BalanceService) GetWithdrawals(userID int64) []model.Withdrawal {
	return s.store.GetWithdrawals(userID)
}
