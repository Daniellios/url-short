-- name: CreatePaste :one
INSERT INTO pastes (title, content, is_public, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, title, content, created_at, updated_at, is_public, expires_at;


-- name: GetPaste :one
SELECT id, title, content, created_at, updated_at, is_public, expires_at
FROM pastes
WHERE id = $1
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);

-- name: GetAllPastes :many
SELECT id, title, content, created_at, updated_at, is_public, expires_at
FROM pastes
WHERE is_public = true
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC
LIMIT 10;
