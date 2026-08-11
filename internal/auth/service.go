package auth

import (
	"context"
	"errors"
	"strings"
	"task-manager-api/internal/user"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo user.Repository
}

var (
	ErrInvalidEmail     = errors.New("email is invalid")
	ErrPasswordTooShort = errors.New("password length is too short")
)

func NewService(repo user.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Register(ctx context.Context, email string, password string) (user.User, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if !isValidEmail(email) {
		return user.User{}, ErrInvalidEmail
	}

	if len(password) < 8 {
		return user.User{}, ErrPasswordTooShort
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return user.User{}, err
	}

	return s.repo.Create(ctx, email, string(passwordHash))

}
