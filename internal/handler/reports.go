package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
)

func (s *Server) siteReport(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodGet) {
		return
	}
	report, err := s.app.BuildSiteReport(r.Context(), r.URL.Query().Get("site_id"))
	if err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusOK, report)
}
