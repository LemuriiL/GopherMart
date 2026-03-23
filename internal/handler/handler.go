// Package handler - HTTP обработчики приложения.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/LemuriiL/GopherMart/internal/middleware"
	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

// AuthService - интерфейс для работы с регистрацией и логином.
type AuthService interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

// OrderService - интерфейс для работы с заказами.
type OrderService interface {
	UploadOrder(number string, userID int64) error
	GetUserOrders(userID int64) ([]model.Order, error)
}

// BalanceService - интерфейс для работы с балансом и списаниями.
type BalanceService interface {
	GetBalance(userID int64) (float64, float64)
	Withdraw(userID int64, order string, sum float64) error
	GetWithdrawals(userID int64) []model.Withdrawal
}

// Handler - объединяет HTTP обработчики приложения.
type Handler struct {
	auth    AuthService
	orders  OrderService
	balance BalanceService
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type orderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// NewHandler - создаёт новый набор HTTP обработчиков.
func NewHandler(auth AuthService, orders OrderService, balance BalanceService) *Handler {
	return &Handler{
		auth:    auth,
		orders:  orders,
		balance: balance,
	}
}

// Register - регистрирует пользователя.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.auth.Register(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login - аутентифицирует пользователя.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// CreateOrder - принимает номер заказа от пользователя.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.orders.UploadOrder(number, userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderNumber) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, storage.ErrOrderUploadedBySameUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, storage.ErrOrderUploadedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders - возвращает список заказов пользователя.
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := h.orders.GetUserOrders(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		item := orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if order.Accrual != 0 {
			accrual := order.Accrual
			item.Accrual = &accrual
		}

		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GetBalance - возвращает текущий баланс пользователя.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	current, withdrawn := h.balance.GetBalance(userID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Withdraw - списывает баллы с баланса пользователя.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isDigits(req.Order) || !isValidLuhn(req.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err := h.balance.Withdraw(userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, service.ErrNotEnoughBalance) || errors.Is(err, storage.ErrNotEnoughBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals - возвращает историю списаний пользователя.
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals := h.balance.GetWithdrawals(userID)
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]withdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		response = append(response, withdrawalResponse{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
