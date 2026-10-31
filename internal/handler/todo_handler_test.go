package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/techxtrasol/todo-app/internal/db"
	"github.com/techxtrasol/todo-app/internal/middleware"
	"github.com/techxtrasol/todo-app/internal/repository"
	"github.com/techxtrasol/todo-app/internal/service"
)

func testToken() string {
	claims := &middleware.Claims{
		UserID: uuid.New(),
		Email:  "t@t.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte("secret"))
	return s
}

type fakeRepo2 struct {
	todos map[uuid.UUID]db.Todo
}

func newFakeRepo2() *fakeRepo2 { return &fakeRepo2{todos: map[uuid.UUID]db.Todo{}} }

func (f *fakeRepo2) Create(ctx context.Context, title string, description *string) (db.Todo, error) {
	t := db.Todo{ID: uuid.New(), Title: title, Description: description, Completed: false}
	f.todos[t.ID] = t
	return t, nil
}
func (f *fakeRepo2) GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return db.Todo{}, pgx.ErrNoRows
	}
	return t, nil
}
func (f *fakeRepo2) List(ctx context.Context) ([]db.Todo, error) {
	out := make([]db.Todo, 0, len(f.todos))
	for _, t := range f.todos {
		out = append(out, t)
	}
	return out, nil
}
func (f *fakeRepo2) Update(ctx context.Context, id uuid.UUID, title, description *string, completed *bool) (db.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return db.Todo{}, pgx.ErrNoRows
	}
	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = description
	}
	if completed != nil {
		t.Completed = *completed
	}
	f.todos[id] = t
	return t, nil
}
func (f *fakeRepo2) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.todos[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.todos, id)
	return nil
}

func TestHandlerCreate(t *testing.T) {
	repo := newFakeRepo2()
	svc := service.NewTodoService(repo)
	h := NewTodoHandler(svc)
	r := NewRouter(h, nil, "secret", nil)

	body, _ := json.Marshal(map[string]string{"title": "t"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+testToken())
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestHandlerGetNotFound(t *testing.T) {
	repo := newFakeRepo2()
	svc := service.NewTodoService(repo)
	h := NewTodoHandler(svc)
	r := NewRouter(h, nil, "secret", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken())
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandlerDeleteNotFound(t *testing.T) {
	repo := newFakeRepo2()
	svc := service.NewTodoService(repo)
	h := NewTodoHandler(svc)
	r := NewRouter(h, nil, "secret", nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+testToken())
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
