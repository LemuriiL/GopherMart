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
