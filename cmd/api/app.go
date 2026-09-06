package main

import (
	"log/slog"
	"task-manager-api/internal/auth"
	authhandler "task-manager-api/internal/auth"
	"task-manager-api/internal/task"
	taskhandler "task-manager-api/internal/task/handler"
	"task-manager-api/internal/user"
)

type app struct {
	taskHandler  *taskhandler.Handler
	authHandler  *authhandler.Handler
	logger       *slog.Logger
	tokenManager *auth.TokenManager
}

func newApp(
	logger *slog.Logger,
	taskRepo task.Repository,
	userRepo user.Repository,
	tokenManager *auth.TokenManager,
) *app {
	taskService := task.NewService(taskRepo)

	userService := auth.NewService(userRepo, tokenManager)

	return &app{
		logger:       logger,
		taskHandler:  taskhandler.New(taskService, logger),
		authHandler:  authhandler.New(*userService, logger),
		tokenManager: tokenManager,
	}
}
