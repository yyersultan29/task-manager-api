package task

import (
	"context"
	"strings"
	"unicode/utf8"
)

type Repository interface {
	List(ctx context.Context) ([]Task, error)
	Create(ctx context.Context, title string) (Task, error)
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

func (s *Service) Create(ctx context.Context, title string) (Task, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return Task{}, ErrTitleRequired
	}

	if utf8.RuneCountInString(title) > maxTitleLength {
		return Task{}, ErrTitleTooLong
	}

	return s.repo.Create(ctx, title)
}

func (s *Service) List(ctx context.Context) ([]Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id int) (Task, error) {
	if err := validateID(id); err != nil {
		return Task{}, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Complete(ctx context.Context, id int) (Task, error) {
	if err := validateID(id); err != nil {
		return Task{}, err
	}
	
	return s.repo.Complete(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func validateID(id int) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}
