package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) identifications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		value, err := s.app.GetIdentification(r.Context(), r.URL.Query().Get("sample_id"))
		if err != nil {
			codec.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		codec.WriteJSON(w, http.StatusOK, value)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	var value model.Identification
	if err := codec.DecodeJSON(r, &value); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.SubmitIdentification(r.Context(), value); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, value)
}
