package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/techxtrasol/todo-app/internal/db"
	"github.com/techxtrasol/todo-app/internal/repository"
)

type TodoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) Create(ctx context.Context, title string, description *string) (db.Todo, error) {
	return s.repo.Create(ctx, title, description)
}

func (s *TodoService) GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TodoService) List(ctx context.Context) ([]db.Todo, error) {
	return s.repo.List(ctx)
}

func (s *TodoService) Update(ctx context.Context, id uuid.UUID, title *string, description *string, completed *bool) (db.Todo, error) {
	return s.repo.Update(ctx, id, title, description, completed)
}

func (s *TodoService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
