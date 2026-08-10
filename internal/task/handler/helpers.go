package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task-manager-api/internal/task"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Error: message,
	})
}

func getID(r *http.Request) (int, error) {
	idText := r.PathValue("id")
	id, err := strconv.Atoi(idText)

	return id, err
}

func writeTaskDomainError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, task.ErrInvalidID) {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return true
	}
	if errors.Is(err, task.ErrTitleRequired) {
		writeError(w, http.StatusBadRequest, "title is required")
		return true
	}
	if errors.Is(err, task.ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return true
	}
	if errors.Is(err, task.ErrTitleTooLong) {
		writeError(w, http.StatusBadRequest, "title is too long")
		return true
	}

	return false
}
