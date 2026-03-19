package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

type AuthService interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

type Handler struct {
	auth AuthService
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewHandler(auth AuthService) *Handler {
	return &Handler{auth: auth}
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

func (h *Handler) CreateOrderStub(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetOrdersStub(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
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
