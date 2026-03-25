// Package app - точка входа приложения.
package app

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/LemuriiL/GopherMart/internal/accrual"
	"github.com/LemuriiL/GopherMart/internal/config"
	"github.com/LemuriiL/GopherMart/internal/handler"
	"github.com/LemuriiL/GopherMart/internal/middleware"
	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

// Run - инициализирует всё приложение и запускает HTTP сервер.
func Run() error {
	cfg := config.New()

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	store := storage.NewPostgresStorage(db)
	if err := store.Init(); err != nil {
		return err
	}

	authService := service.NewAuthService(store, cfg.JWTSecret)
	orderService := service.NewOrderService(store)
	balanceService := service.NewBalanceService(store)

	accrualClient := accrual.NewClient(cfg.AccrualAddress)
	worker := service.NewWorker(store, accrualClient)
	worker.Start()

	h := handler.NewHandler(authService, orderService, balanceService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/user/register", h.Register)
	mux.HandleFunc("POST /api/user/login", h.Login)

	authMiddleware := middleware.Auth(cfg.JWTSecret)

	mux.Handle("POST /api/user/orders", authMiddleware(http.HandlerFunc(h.CreateOrder)))
	mux.Handle("GET /api/user/orders", authMiddleware(http.HandlerFunc(h.GetOrders)))
	mux.Handle("GET /api/user/balance", authMiddleware(http.HandlerFunc(h.GetBalance)))
	mux.Handle("POST /api/user/balance/withdraw", authMiddleware(http.HandlerFunc(h.Withdraw)))
	mux.Handle("GET /api/user/withdrawals", authMiddleware(http.HandlerFunc(h.GetWithdrawals)))

	return http.ListenAndServe(cfg.RunAddress, mux)
}
