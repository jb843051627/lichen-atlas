package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) taxa(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := s.app.ListTaxa(r.Context(), r.URL.Query().Get("rank"))
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
	var taxon model.Taxon
	if err := codec.DecodeJSON(r, &taxon); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.CreateTaxon(r.Context(), taxon); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, taxon)
}
