package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/LemuriiL/GopherMart/internal/model"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserReaderWriter interface {
	CreateUser(login, passwordHash string) (*model.User, error)
	GetUserByLogin(login string) (*model.User, error)
}

type AuthService struct {
	store     UserReaderWriter
	jwtSecret string
}

func NewAuthService(store UserReaderWriter, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(login, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user, err := s.store.CreateUser(login, string(hash))
	if err != nil {
		return "", err
	}

	return s.generateToken(user.ID)
}

func (s *AuthService) Login(login, password string) (string, error) {
	user, err := s.store.GetUserByLogin(login)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

func (s *AuthService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.jwtSecret))
}
