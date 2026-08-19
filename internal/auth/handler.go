package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"task-manager-api/internal/httputil"
	"task-manager-api/internal/user"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

// it repeats needs to save in folder common ??
const (
	maxJSONBodyBytes = 1 << 20 // 1 MiB
)

func New(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
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

	if request.Email == nil || request.Password == nil {
		httputil.WriteError(w, http.StatusBadRequest, "email and password required")
		return
	}

	createdUser, err := h.service.Register(r.Context(), *request.Email, *request.Password)

	if errors.Is(err, ErrInvalidEmail) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}

	if errors.Is(err, ErrPasswordTooShort) {
		httputil.WriteError(w, http.StatusBadRequest, "password too short")
		return
	}

	if errors.Is(err, user.ErrEmailAlreadyExists) {
		httputil.WriteError(w, http.StatusConflict, "email already exists")
		return
	}

	if err != nil {
		h.logger.Error("register user failed", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// тут по идее токен надо или ок ?

	httputil.WriteJSON(
		w, http.StatusCreated,
		registerResponse{ID: int(createdUser.ID), Email: createdUser.Email})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
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

	if request.Email == nil || request.Password == nil {
		httputil.WriteError(w, http.StatusBadRequest, "email and password required")
		return
	}

	user, err := h.service.Login(r.Context(), *request.Email, *request.Password)

	if errors.Is(err, ErrInvalidEmail) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}

	if errors.Is(err, ErrInvalidCredentials) {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err != nil {
		h.logger.Error("login failed", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "login failed")
		return
	}

	httputil.WriteJSON(
		w, http.StatusOK,
		loginResponse{Email: &user.Email})

}
