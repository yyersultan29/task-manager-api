package handler

import (
	"context"
	"log/slog"
	"task-manager-api/internal/task"
)

const (
	maxJSONBodyBytes = 1 << 20 // 1 MiB
	defaultListLimit = 20
	maxListLimit     = 100
)

type Service interface {
	List(ctx context.Context, options task.ListOptions) ([]task.Task, error)
	Create(ctx context.Context, title string) (task.Task, error)
	FindByID(ctx context.Context, id int) (task.Task, error)
	Complete(ctx context.Context, id int) (task.Task, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	service Service
	logger  *slog.Logger
}

func New(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
