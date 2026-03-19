package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/LemuriiL/GopherMart/internal/accrual"
	"github.com/LemuriiL/GopherMart/internal/config"
	"github.com/LemuriiL/GopherMart/internal/handler"
	"github.com/LemuriiL/GopherMart/internal/middleware"
	"github.com/LemuriiL/GopherMart/internal/service"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

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

	r := chi.NewRouter()

	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Post("/api/user/orders", h.CreateOrder)
		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalance)
		r.Post("/api/user/balance/withdraw", h.Withdraw)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})

	return http.ListenAndServe(cfg.RunAddress, r)
}
