package model

import "time"

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

type Order struct {
	Number     string
	UserID     int64
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type Withdrawal struct {
	Order       string    `json:"order"`
	UserID      int64     `json:"-"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
