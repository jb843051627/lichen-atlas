package handler

import (
	"net/http"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Server) sites(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := s.app.ListSites(r.Context(), r.URL.Query().Get("region"))
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
	var site model.Site
	if err := codec.DecodeJSON(r, &site); err != nil {
		codec.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.app.CreateSite(r.Context(), site); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusCreated, site)
}

func (s *Server) siteByID(w http.ResponseWriter, r *http.Request, id string) {
	if strings.TrimSpace(id) == "" {
		http.NotFound(w, r)
		return
	}
	site, err := s.app.GetSite(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	codec.WriteJSON(w, http.StatusOK, site)
}
