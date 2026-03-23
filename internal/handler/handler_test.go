package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LemuriiL/GopherMart/internal/middleware"
	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

type authServiceMock struct {
	registerToken string
	loginToken    string
	registerErr   error
	loginErr      error
}

func (m *authServiceMock) Register(login, password string) (string, error) {
	return m.registerToken, m.registerErr
}

func (m *authServiceMock) Login(login, password string) (string, error) {
	return m.loginToken, m.loginErr
}

type orderServiceMock struct {
	uploadErr error
	orders    []model.Order
	getErr    error
}

func (m *orderServiceMock) UploadOrder(number string, userID int64) error {
	return m.uploadErr
}

func (m *orderServiceMock) GetUserOrders(userID int64) ([]model.Order, error) {
	return m.orders, m.getErr
}

type balanceServiceMock struct {
	current     float64
	withdrawn   float64
	withdrawErr error
	withdrawals []model.Withdrawal
}

func (m *balanceServiceMock) GetBalance(userID int64) (float64, float64) {
	return m.current, m.withdrawn
}

func (m *balanceServiceMock) Withdraw(userID int64, order string, sum float64) error {
	return m.withdrawErr
}

func (m *balanceServiceMock) GetWithdrawals(userID int64) []model.Withdrawal {
	return m.withdrawals
}

func newTestHandler() *Handler {
	return NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{},
	)
}

func withUserID(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestRegisterSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{registerToken: "token123"},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	body := []byte(`{"login":"test","password":"test123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if rec.Header().Get("Authorization") != "Bearer token123" {
		t.Fatalf("expected Authorization header with token")
	}
}

func TestRegisterBadRequest(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(`{bad json`)))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterConflict(t *testing.T) {
	h := NewHandler(
		&authServiceMock{registerErr: storage.ErrUserExists},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	body := []byte(`{"login":"test","password":"test123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{loginToken: "token123"},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	body := []byte(`{"login":"test","password":"test123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if rec.Header().Get("Authorization") != "Bearer token123" {
		t.Fatalf("expected Authorization header with token")
	}
}

func TestLoginUnauthorized(t *testing.T) {
	h := NewHandler(
		&authServiceMock{loginErr: service.ErrInvalidCredentials},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	body := []byte(`{"login":"test","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateOrderSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("79927398713")))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestCreateOrderUnauthorized(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("79927398713")))
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateOrderInvalidNumber(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{uploadErr: service.ErrInvalidOrderNumber},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345")))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestCreateOrderConflict(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{uploadErr: storage.ErrOrderUploadedByAnotherUser},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("79927398713")))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestGetOrdersNoContent(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{orders: []model.Order{}},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetOrders(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestGetOrdersSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{
			orders: []model.Order{
				{
					Number:     "79927398713",
					Status:     "PROCESSED",
					Accrual:    500,
					UploadedAt: mustParseTime("2026-03-23T21:10:00+03:00"),
				},
			},
		},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetOrders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 order, got %d", len(resp))
	}

	if resp[0]["number"] != "79927398713" {
		t.Fatalf("unexpected number: %v", resp[0]["number"])
	}

	if resp[0]["status"] != "PROCESSED" {
		t.Fatalf("unexpected status: %v", resp[0]["status"])
	}
}

func TestGetBalanceSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{current: 100, withdrawn: 25},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]float64
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if resp["current"] != 100 {
		t.Fatalf("expected current 100, got %v", resp["current"])
	}

	if resp["withdrawn"] != 25 {
		t.Fatalf("expected withdrawn 25, got %v", resp["withdrawn"])
	}
}

func TestWithdrawSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	body := []byte(`{"order":"12345678903","sum":50}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Withdraw(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithdrawPaymentRequired(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{withdrawErr: service.ErrNotEnoughBalance},
	)

	body := []byte(`{"order":"12345678903","sum":50}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Withdraw(rec, req)

	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rec.Code)
	}
}

func TestWithdrawInvalidOrder(t *testing.T) {
	h := newTestHandler()

	body := []byte(`{"order":"12345","sum":50}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Withdraw(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestGetWithdrawalsNoContent(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{withdrawals: []model.Withdrawal{}},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetWithdrawals(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestGetWithdrawalsSuccess(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{
			withdrawals: []model.Withdrawal{
				{
					Order:       "12345678903",
					UserID:      1,
					Sum:         50,
					ProcessedAt: mustParseTime("2026-03-23T21:15:00+03:00"),
				},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetWithdrawals(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 withdrawal, got %d", len(resp))
	}

	if resp[0]["order"] != "12345678903" {
		t.Fatalf("unexpected order: %v", resp[0]["order"])
	}
}

func TestRegisterEmptyFields(t *testing.T) {
	h := newTestHandler()

	body := []byte(`{"login":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLoginBadRequest(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte(`{bad json`)))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateOrderEmptyBody(t *testing.T) {
	h := NewHandler(
		&authServiceMock{},
		&orderServiceMock{},
		&balanceServiceMock{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("")))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetBalanceUnauthorized(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWithdrawBadJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte(`{bad json`)))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Withdraw(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetWithdrawalsUnauthorized(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rec := httptest.NewRecorder()

	h.GetWithdrawals(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
