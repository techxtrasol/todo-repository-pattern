package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	authmw "github.com/techxtrasol/todo-app/internal/middleware"
)

//go:embed docs
var docsFS embed.FS

func NewRouter(todoHandler *TodoHandler, authHandler *AuthHandler, jwtSecret string, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.Group(func(r chi.Router) {
				r.Use(authmw.Auth(jwtSecret))
				r.Get("/me", authHandler.Me)
			})
		})

		r.Route("/todos", func(r chi.Router) {
			r.Use(authmw.Auth(jwtSecret))
			r.Get("/", todoHandler.List)
			r.Post("/", todoHandler.Create)
			r.Get("/{id}", todoHandler.GetByID)
			r.Put("/{id}", todoHandler.Update)
			r.Delete("/{id}", todoHandler.Delete)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if pool != nil {
			if err := pool.Ping(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy"}`))
				return
			}
		}
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
