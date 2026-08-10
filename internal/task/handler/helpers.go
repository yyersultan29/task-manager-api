package handler

import (
	"errors"
	"net/http"
	"strconv"
	"task-manager-api/internal/httputil"
	"task-manager-api/internal/task"
)

func getID(r *http.Request) (int, error) {
	idText := r.PathValue("id")
	id, err := strconv.Atoi(idText)

	return id, err
}

func writeTaskDomainError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, task.ErrInvalidID) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return true
	}
	if errors.Is(err, task.ErrTitleRequired) {
		httputil.WriteError(w, http.StatusBadRequest, "title is required")
		return true
	}
	if errors.Is(err, task.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "task not found")
		return true
	}
	if errors.Is(err, task.ErrTitleTooLong) {
		httputil.WriteError(w, http.StatusBadRequest, "title is too long")
		return true
	}

	return false
}
