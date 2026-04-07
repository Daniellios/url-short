-- name: CreatePaste :one
INSERT INTO pastes (content) VALUES ($1) RETURNING id, content, created_at, updated_at;


-- name: GetPaste :one
SELECT id, content, created_at, updated_at FROM pastes WHERE id = $1;

-- name: GetAllPastes :many
SELECT id, content, created_at, updated_at FROM pastes;