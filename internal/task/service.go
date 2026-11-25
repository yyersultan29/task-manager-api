package task

import (
	"context"
	"strings"
	"unicode/utf8"
)

type Repository interface {
	List(ctx context.Context, userID int64, options ListOptions) ([]Task, error)
	Create(ctx context.Context, userID int64, title string) (Task, error)
	FindByID(ctx context.Context, id int) (Task, error)
	Complete(ctx context.Context, id int) (Task, error)
	Delete(ctx context.Context, id int) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID int64, title string) (Task, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return Task{}, ErrTitleRequired
	}

	if utf8.RuneCountInString(title) > maxTitleLength {
		return Task{}, ErrTitleTooLong
	}

	return s.repo.Create(ctx, userID, title)
}

func (s *Service) List(ctx context.Context, userID int64, options ListOptions) ([]Task, error) {
	return s.repo.List(ctx, userID, options)
}

func (s *Service) FindByID(ctx context.Context, id int, userID int64) (Task, error) {
	if err := validateID(id); err != nil {
		return Task{}, err
	}
	foundTask, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return Task{}, err
	}

	if foundTask.UserID != userID {
		return Task{}, ErrForbidden
	}

	return foundTask, err
}

func (s *Service) Complete(ctx context.Context, id int, userID int64) (Task, error) {
	if err := validateID(id); err != nil {
		return Task{}, err
	}

	foundTask, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return Task{}, err
	}

	if foundTask.UserID != userID {
		return Task{}, ErrForbidden
	}

	return s.repo.Complete(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int, userID int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	foundTask, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return err
	}

	if foundTask.UserID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func validateID(id int) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}
