package handler

import (
	"context"
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/codec"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodGet) {
		return
	}
	if err := s.app.Health(context.Background()); err != nil {
		codec.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "down", "error": err.Error()})
		return
	}
	codec.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
