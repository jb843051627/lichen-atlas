package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) locations(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodPost) {
		return
	}
	var location struct {
		SampleID string `json:"sample_id"`
		model.Location
	}
	if err := codec.DecodeJSON(r, &location); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.Store().SaveLocation(r.Context(), location.SampleID, location.Location); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, location)
}
