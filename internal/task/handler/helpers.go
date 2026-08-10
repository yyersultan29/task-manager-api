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

func parseListOptions(r *http.Request) (task.ListOptions, error) {
	query := r.URL.Query()

	options := task.ListOptions{
		Limit: defaultListLimit,
	}

	if limitText := query.Get("limit"); limitText != "" {
		limit, err := strconv.Atoi(limitText)
		if err != nil || limit < 1 || limit > maxListLimit {
			return task.ListOptions{}, errors.New("invalid limit")
		}
		options.Limit = limit
	}

	if offsetText := query.Get("offset"); offsetText != "" {
		offset, err := strconv.Atoi(offsetText)
		if err != nil || offset < 0 {
			return task.ListOptions{}, errors.New("invalid offset")
		}
		options.Offset = offset
	}

	switch query.Get("done") {
	case "":
		return options, nil
	case "true":
		done := true
		options.Done = &done
	case "false":
		done := false
		options.Done = &done
	default:
		return task.ListOptions{}, errors.New("invalid done")
	}

	return options, nil
}
