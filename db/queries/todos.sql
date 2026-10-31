-- name: CreateTodo :one
INSERT INTO todos (title, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetTodo :one
SELECT * FROM todos
WHERE id = $1;

-- name: ListTodos :many
SELECT * FROM todos
ORDER BY created_at DESC;

-- name: UpdateTodo :one
UPDATE todos
SET 
    title = COALESCE(sqlc.narg('title')::text, title),
    description = COALESCE(sqlc.narg('description')::text, description),
    completed = COALESCE(sqlc.narg('completed')::bool, completed),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTodo :execrows
DELETE FROM todos
WHERE id = $1;
