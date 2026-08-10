package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"task-manager-api/internal/httputil"
)

func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {

	options, err := parseListOptions(r)

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid query parameters")
		return
	}

	tasks, err := h.service.List(r.Context(), options)
	if err != nil {
		h.logger.Error("can not get task from bd", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "could not list tasks")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, toTaskResponses(tasks))
}

func (h *Handler) HandleTaskCreate(w http.ResponseWriter, r *http.Request) {
	var request createTaskRequest
	var maxBytesErr *http.MaxBytesError

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

	if err != nil || mediaType != "application/json" {
		httputil.WriteError(w, http.StatusUnsupportedMediaType, "Content-type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&request)

	if errors.As(err, &maxBytesErr) {
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	err = decoder.Decode(&struct{}{})

	if errors.As(err, &maxBytesErr) {
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	if err != io.EOF {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if request.Title == nil {
		httputil.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}

	createdTask, err := h.service.Create(r.Context(), *request.Title)
	if writeTaskDomainError(w, err) {
		return
	}

	if err != nil {
		h.logger.Error("cannot create task in bd", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "could not create task")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, toTaskResponse(createdTask))
}

func (h *Handler) HandleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if writeTaskDomainError(w, err) {
		return
	}
	if err != nil {
		h.logger.Error("cannot delete in bd", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "could not delete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleTaskComplete(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	updatedTask, err := h.service.Complete(r.Context(), id)
	if writeTaskDomainError(w, err) {
		return
	}

	if err != nil {
		h.logger.Error("update task in bd error", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "could not update task")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toTaskResponse(updatedTask))
}

func (h *Handler) HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	foundTask, err := h.service.FindByID(r.Context(), id)
	if writeTaskDomainError(w, err) {
		return
	}

	if err != nil {
		h.logger.Error("get task by id in bd", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "could not get task")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toTaskResponse(foundTask))
}
