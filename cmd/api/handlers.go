package main

import (
	"context"
	"net/http"
	"time"
)

func (app *app) handleSlow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := waitForWork(ctx, 2*time.Second)

	if err != nil {
		app.logger.Info("request cancelled", "error", err)
		return
	}
	w.Write([]byte("done"))
}

func (app *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (app *app) handleTimeout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	err := waitForWork(ctx, 2*time.Second)
	if err != nil {
		writeError(w, http.StatusGatewayTimeout, err.Error())
		return
	}

}

func waitForWork(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
