package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task-manager-api/internal/task"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	logger := newLogger()
	// stop proccess if load config takes more than 5 sec
	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("could not load config", "error", err)
		os.Exit(1)
	}

	db, err := pgxpool.New(startupCtx, cfg.databaseURL)

	if err != nil {
		logger.Error("could not create database pool", "error", err)
		os.Exit(1)

	}
	defer db.Close()

	if err := db.Ping(startupCtx); err != nil {
		logger.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	logger.Info("database connected")

	application := newApp(logger, task.NewPostgresRepository(db))

	server := &http.Server{
		Addr:         cfg.httpAddr,
		Handler:      application.routes(),
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
		IdleTimeout:  cfg.idleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server started", "addr", cfg.httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped with error", "error", err)
			os.Exit(1)
		}
	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received")
	}

	stop()

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(gracefulCtx); err != nil {
		logger.Error("could not shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
