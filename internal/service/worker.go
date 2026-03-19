package service

import (
	"log"
	"time"

	"github.com/LemuriiL/GopherMart/internal/accrual"
	"github.com/LemuriiL/GopherMart/internal/model"
)

type WorkerStorage interface {
	GetNewOrders() []model.Order
	UpdateOrder(number string, status string, accrual float64)
}

type Worker struct {
	store    WorkerStorage
	client   *accrual.Client
	interval time.Duration
}

func NewWorker(store WorkerStorage, client *accrual.Client) *Worker {
	return &Worker{
		store:    store,
		client:   client,
		interval: 2 * time.Second,
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.process()
			time.Sleep(w.interval)
		}
	}()
}

func (w *Worker) process() {
	orders := w.store.GetNewOrders()

	for _, order := range orders {
		resp, statusCode, err := w.client.GetOrder(order.Number)
		if err != nil {
			continue
		}

		if statusCode == 429 {
			log.Println("rate limited")
			time.Sleep(5 * time.Second)
			continue
		}

		if resp == nil {
			continue
		}

		accrual := 0.0
		if resp.Accrual != nil {
			accrual = *resp.Accrual
		}

		w.store.UpdateOrder(order.Number, resp.Status, accrual)
	}
}
