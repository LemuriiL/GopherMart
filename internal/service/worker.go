// Package service - бизнес логика приложения.
package service

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/LemuriiL/GopherMart/internal/accrual"
	"github.com/LemuriiL/GopherMart/internal/model"
)

// WorkerStorage - интерфейс для работы воркера с хранилищем.
type WorkerStorage interface {
	GetNewOrders() []model.Order
	UpdateOrder(number string, status string, accrual float64)
}

// Worker - фоновый обработчик заказов.
type Worker struct {
	store        WorkerStorage
	client       *accrual.Client
	pollInterval time.Duration
}

// NewWorker - создаёт нового воркера для опроса системы начислений.
func NewWorker(store WorkerStorage, client *accrual.Client) *Worker {
	return &Worker{
		store:        store,
		client:       client,
		pollInterval: 2 * time.Second,
	}
}

// Start - запускает воркер в отдельной горутине.
func (w *Worker) Start() {
	go func() {
		for {
			w.process()
			time.Sleep(w.pollInterval)
		}
	}()
}

// process - обрабатывает новые заказы и обновляет их статусы.
func (w *Worker) process() {
	orders := w.store.GetNewOrders()

	for _, order := range orders {
		resp, statusCode, retryAfter, err := w.client.GetOrder(order.Number)
		if err != nil {
			slog.Error("accrual request failed", "order", order.Number, "error", err)
			continue
		}

		switch statusCode {
		case http.StatusNoContent:
			continue
		case http.StatusTooManyRequests:
			if retryAfter > 0 {
				time.Sleep(time.Duration(retryAfter) * time.Second)
			}
			continue
		case http.StatusOK:
			if resp == nil {
				continue
			}

			internalStatus := mapAccrualStatus(resp.Status)
			accrualValue := order.Accrual
			if resp.Accrual != nil {
				accrualValue = *resp.Accrual
			}

			w.store.UpdateOrder(order.Number, internalStatus, accrualValue)
		default:
			continue
		}
	}
}

// mapAccrualStatus - переводит статус внешнего сервиса во внутренний статус заказа.
func mapAccrualStatus(status string) string {
	switch status {
	case "REGISTERED":
		return "NEW"
	case "PROCESSING":
		return "PROCESSING"
	case "INVALID":
		return "INVALID"
	case "PROCESSED":
		return "PROCESSED"
	default:
		return "NEW"
	}
}
