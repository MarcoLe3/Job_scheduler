package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"scheduler/internal/db"
)

type Handler struct {
	service *Service
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", h.submit)
	mux.HandleFunc("GET /jobs/{id}", h.get)
	mux.HandleFunc("POST /jobs/{id}/cancel", h.cancel)
	return mux
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var req job.JobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.service.Submit(r.Context(), req)
	switch err {
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	job, err := h.service.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {

}
