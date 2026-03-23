package service

import (
	"testing"

	"github.com/LemuriiL/GopherMart/internal/model"
	"github.com/LemuriiL/GopherMart/internal/storage"
)

type authTestStore struct {
	users  map[string]*model.User
	nextID int64
}

func newAuthTestStore() *authTestStore {
	return &authTestStore{
		users:  make(map[string]*model.User),
		nextID: 1,
	}
}

func (s *authTestStore) CreateUser(login, passwordHash string) (*model.User, error) {
	if _, ok := s.users[login]; ok {
		return nil, storage.ErrUserExists
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

func (s *authTestStore) GetUserByLogin(login string) (*model.User, error) {
	user, ok := s.users[login]
	if !ok {
		return nil, storage.ErrUserNotFound
	}

	return user, nil
}

func TestAuthServiceRegister(t *testing.T) {
	store := newAuthTestStore()
	service := NewAuthService(store, "secret")

	token, err := service.Register("user1", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestAuthServiceRegisterDuplicate(t *testing.T) {
	store := newAuthTestStore()
	service := NewAuthService(store, "secret")

	_, err := service.Register("user1", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = service.Register("user1", "password2")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	store := newAuthTestStore()
	service := NewAuthService(store, "secret")

	_, err := service.Register("user1", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := service.Login("user1", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestAuthServiceLoginWrongPassword(t *testing.T) {
	store := newAuthTestStore()
	service := NewAuthService(store, "secret")

	_, err := service.Register("user1", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = service.Login("user1", "wrong")
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceLoginUnknownUser(t *testing.T) {
	store := newAuthTestStore()
	service := NewAuthService(store, "secret")

	_, err := service.Login("unknown", "password1")
	if err == nil {
		t.Fatal("expected error")
	}

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
