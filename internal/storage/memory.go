package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/LemuriiL/GopherMart/internal/model"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrOrderUploadedBySameUser = errors.New("order already uploaded by same user")
var ErrOrderUploadedByAnotherUser = errors.New("order already uploaded by another user")

type MemoryStorage struct {
	mu     sync.RWMutex
	users  map[string]*model.User
	orders map[string]*model.Order
	nextID int64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:  make(map[string]*model.User),
		orders: make(map[string]*model.Order),
		nextID: 1,
	}
}

func (s *MemoryStorage) CreateUser(login, passwordHash string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[login]; ok {
		return nil, ErrUserExists
	}

	user := &model.User{
		ID:           s.nextID,
		Login:        login,
		PasswordHash: passwordHash,
	}

	s.users[login] = user
	s.nextID++

	return user, nil
}

func (s *MemoryStorage) GetUserByLogin(login string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[login]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *MemoryStorage) SaveOrder(number string, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[number]
	if ok {
		if order.UserID == userID {
			return ErrOrderUploadedBySameUser
		}
		return ErrOrderUploadedByAnotherUser
	}

	s.orders[number] = &model.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	return nil
}
