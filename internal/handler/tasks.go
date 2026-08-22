package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
)

func (s *Server) runTask(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodPost) {
		return
	}
	if err := s.app.RunNextTask(r.Context()); err != nil {
		codec.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusOK, map[string]string{"status": "done"})
}
