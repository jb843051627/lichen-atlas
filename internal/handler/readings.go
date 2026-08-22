package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) readings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Query().Get("sample_id") == "" {
			http.Error(w, "missing sample", http.StatusBadRequest)
			return
		}
		values, err := s.app.ListReadings(r.Context(), r.URL.Query().Get("sample_id"))
		if err != nil {
			codec.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		codec.WriteJSON(w, http.StatusOK, values)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	var reading model.Reading
	if err := codec.DecodeJSON(r, &reading); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.AddReading(r.Context(), reading); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, reading)
}
