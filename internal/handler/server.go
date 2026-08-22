package handler

import (
	"net/http"

	"github.com/jb843051627/lichen-atlas/internal/service"
	"github.com/jb843051627/lichen-atlas/internal/web"
)

type Server struct {
	app *service.Application
	mux *http.ServeMux
}

func NewServer(app *service.Application) *Server {
	s := &Server{app: app, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler            { return requestLog(s.mux) }
func (s *Server) ListenAndServe(addr string) error { return http.ListenAndServe(addr, s.Handler()) }

func (s *Server) routes() {
	s.mux.HandleFunc("/", web.ServeIndex)
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/api/sites", s.sites)
	s.mux.HandleFunc("/api/samples", s.samples)
	s.mux.HandleFunc("/api/readings", s.readings)
	s.mux.HandleFunc("/api/locations", s.locations)
	s.mux.HandleFunc("/api/taxa", s.taxa)
	s.mux.HandleFunc("/api/identifications", s.identifications)
	s.mux.HandleFunc("/api/reviews", s.reviews)
	s.mux.HandleFunc("/api/archives", s.archives)
	s.mux.HandleFunc("/api/reports/site", s.siteReport)
	s.mux.HandleFunc("/api/tasks/run", s.runTask)
}
