package main

import "net/http"

func (app *app) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", app.taskHandler.HandleTasks)
	mux.HandleFunc("GET /tasks/{id}", app.taskHandler.HandleTaskByID)
	mux.HandleFunc("POST /tasks", app.taskHandler.HandleTaskCreate)
	mux.HandleFunc("DELETE /tasks/{id}", app.taskHandler.HandleTaskDelete)
	mux.HandleFunc("POST /tasks/{id}/complete", app.taskHandler.HandleTaskComplete)
	mux.HandleFunc("GET /slow", app.handleSlow)
	mux.HandleFunc("GET /timeout", app.handleTimeout)
	mux.HandleFunc("GET /health", app.handleHealth)

	return requestIDMiddleware(loggingMiddleware(app.logger, mux))
}
