package storage

import (
	"errors"
	"sync"

	"github.com/LemuriiL/GopherMart/internal/model"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

type MemoryStorage struct {
	mu     sync.RWMutex
	users  map[string]*model.User
	nextID int64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:  make(map[string]*model.User),
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
