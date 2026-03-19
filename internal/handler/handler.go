package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LemuriiL/GopherMart/internal/middleware"
	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

type AuthService interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

type OrderService interface {
	UploadOrder(number string, userID int64) error
	GetUserOrders(userID int64) ([]model.Order, error)
}

type Handler struct {
	auth   AuthService
	orders OrderService
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type orderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func NewHandler(auth AuthService, orders OrderService) *Handler {
	return &Handler{
		auth:   auth,
		orders: orders,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "empty login or password", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Register(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "empty login or password", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

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
			UploadedAt: order.UploadedAt,
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

func (h *Handler) GetBalanceStub(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) WithdrawStub(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetWithdrawalsStub(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
