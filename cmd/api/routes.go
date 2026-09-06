package main

import "net/http"

func (app *app) routes() http.Handler {

	mux := http.NewServeMux()

	mux.Handle("GET /tasks", app.authMiddleware(http.HandlerFunc(app.taskHandler.HandleTasks)))

	mux.Handle("GET /tasks/{id}", app.authMiddleware(http.HandlerFunc(app.taskHandler.HandleTaskByID)))
	mux.Handle("POST /tasks", app.authMiddleware(http.HandlerFunc(app.taskHandler.HandleTaskCreate)))
	mux.Handle("DELETE /tasks/{id}", app.authMiddleware(http.HandlerFunc(app.taskHandler.HandleTaskDelete)))
	mux.Handle("POST /tasks/{id}/complete", app.authMiddleware(http.HandlerFunc(app.taskHandler.HandleTaskComplete)))

	mux.HandleFunc("POST /auth/register", app.authHandler.HandleRegister)
	mux.HandleFunc("POST /auth/login", app.authHandler.HandleLogin)

	mux.HandleFunc("GET /slow", app.handleSlow)
	mux.HandleFunc("GET /timeout", app.handleTimeout)
	mux.HandleFunc("GET /health", app.handleHealth)

	return requestIDMiddleware(loggingMiddleware(app.logger, mux))
}
