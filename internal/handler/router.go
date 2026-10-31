package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed docs
var docsFS embed.FS

func NewRouter(todoHandler *TodoHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/todos", func(r chi.Router) {
			r.Get("/", todoHandler.List)
			r.Post("/", todoHandler.Create)
			r.Get("/{id}", todoHandler.GetByID)
			r.Put("/{id}", todoHandler.Update)
			r.Delete("/{id}", todoHandler.Delete)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	docsSub, _ := fs.Sub(docsFS, "docs")
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(docsSub, "openapi.yaml")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(data)
	})
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(docsSub, "swagger.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	r.Get("/redoc", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(docsSub, "index.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	return r
}
