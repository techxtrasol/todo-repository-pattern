package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/techxtrasol/todo-app/internal/db"
)

type TodoRepository interface {
	Create(ctx context.Context, title string, description *string) (db.Todo, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error)
	List(ctx context.Context) ([]db.Todo, error)
	Update(ctx context.Context, id uuid.UUID, title *string, description *string, completed *bool) (db.Todo, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type todoRepository struct {
	queries *db.Queries
}

func NewTodoRepository(pool *pgxpool.Pool) TodoRepository {
	return &todoRepository{
		queries: db.New(pool),
	}
}

func (r *todoRepository) Create(ctx context.Context, title string, description *string) (db.Todo, error) {
	return r.queries.CreateTodo(ctx, db.CreateTodoParams{
		Title:       title,
		Description: description,
	})
}

func (r *todoRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Todo, error) {
	return r.queries.GetTodo(ctx, id)
}

func (r *todoRepository) List(ctx context.Context) ([]db.Todo, error) {
	return r.queries.ListTodos(ctx)
}

func (r *todoRepository) Update(ctx context.Context, id uuid.UUID, title *string, description *string, completed *bool) (db.Todo, error) {
	params := db.UpdateTodoParams{
		ID:          id,
		Title:       title,
		Description: description,
		Completed:   completed,
	}

	return r.queries.UpdateTodo(ctx, params)
}

var ErrNotFound = errors.New("not found")

func (r *todoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.DeleteTodo(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
