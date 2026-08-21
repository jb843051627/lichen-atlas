package handler

import (
	"net/http"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) samples(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		siteID := r.URL.Query().Get("site_id")
		if siteID == "" {
			codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "site_id is required"})
			return
		}
		values, err := s.app.ListSamples(r.Context(), siteID)
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
	var sample model.Sample
	if err := codec.DecodeJSON(r, &sample); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.CreateSample(r.Context(), sample); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, sample)
}

func (s *Server) sampleByID(w http.ResponseWriter, r *http.Request, id string) {
	if strings.TrimSpace(id) == "" {
		http.NotFound(w, r)
		return
	}
	sample, err := s.app.GetSample(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	codec.WriteJSON(w, http.StatusOK, sample)
}
