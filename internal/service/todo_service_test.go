package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/techxtrasol/todo-app/internal/db"
	"github.com/techxtrasol/todo-app/internal/repository"
)

type fakeRepo struct {
	todos map[uuid.UUID]db.Todo
}

func newFakeRepo() *fakeRepo { return &fakeRepo{todos: map[uuid.UUID]db.Todo{}} }

func (f *fakeRepo) Create(ctx context.Context, title string, description *string) (db.Todo, error) {
	t := db.Todo{ID: uuid.New(), Title: title, Description: description, Completed: false}
	f.todos[t.ID] = t
	return t, nil
}
func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return db.Todo{}, errors.New("not found")
	}
	return t, nil
}
func (f *fakeRepo) List(ctx context.Context) ([]db.Todo, error) {
	out := make([]db.Todo, 0, len(f.todos))
	for _, t := range f.todos {
		out = append(out, t)
	}
	return out, nil
}
func (f *fakeRepo) Update(ctx context.Context, id uuid.UUID, title, description *string, completed *bool) (db.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return db.Todo{}, errors.New("not found")
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
func (f *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.todos[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.todos, id)
	return nil
}

func TestServiceCreateAndGet(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTodoService(repo)
	ctx := context.Background()

	desc := "desc"
	todo, err := svc.Create(ctx, "title", &desc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "title" {
		t.Fatalf("expected title, got %s", got.Title)
	}
}

func TestServiceUpdatePartial(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTodoService(repo)
	ctx := context.Background()

	todo, _ := svc.Create(ctx, "title", nil)
	completed := true
	updated, err := svc.Update(ctx, todo.ID, nil, nil, &completed)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Completed || updated.Title != "title" {
		t.Fatalf("partial update failed: %+v", updated)
	}
}

func TestServiceDeleteNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTodoService(repo)
	ctx := context.Background()
	err := svc.Delete(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
