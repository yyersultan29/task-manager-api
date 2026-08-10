package main

import (
	"log/slog"
	"task-manager-api/internal/task"
	taskhandler "task-manager-api/internal/task/handler"
)

type app struct {
	taskHandler *taskhandler.Handler
	logger      *slog.Logger
}

func newApp(logger *slog.Logger, repo task.Repository) *app {
	taskService := task.NewService(repo)

	return &app{
		logger:      logger,
		taskHandler: taskhandler.New(taskService, logger),
	}
}
