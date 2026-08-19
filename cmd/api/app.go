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
	taskHandler *taskhandler.Handler
	authHandler *authhandler.Handler
	logger      *slog.Logger
}

func newApp(
	logger *slog.Logger,
	taskRepo task.Repository,
	userRepo user.Repository) *app {
	taskService := task.NewService(taskRepo)

	userService := auth.NewService(userRepo)

	return &app{
		logger:      logger,
		taskHandler: taskhandler.New(taskService, logger),
		authHandler: authhandler.New(*userService, logger),
	}
}
