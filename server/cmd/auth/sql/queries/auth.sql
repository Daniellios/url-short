-- name: CreateUser :one
INSERT INTO users (email, display_name)
VALUES ($1, $2)
RETURNING id, email, display_name, created_at, updated_at;

-- name: UpdateUserEmailAndDisplay :exec
UPDATE users
SET
    email = $2,
    display_name = $3,
    updated_at = now()
WHERE
    id = $1;

-- name: GetUser :one
SELECT id, email, display_name, created_at, updated_at
FROM users
WHERE
    id = $1;

-- name: GetOAuthIdentity :one
SELECT
    id,
    user_id,
    provider,
    provider_subject,
    created_at
FROM oauth_identities
WHERE
    provider = $1
    AND provider_subject = $2;

-- name: CreateOAuthIdentity :exec
INSERT INTO oauth_identities (user_id, provider, provider_subject)
VALUES ($1, $2, $3);

-- name: CreateSession :one
INSERT INTO
    sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM sessions
WHERE
    token_hash = $1
    AND expires_at > now();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE
    id = $1;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions
WHERE
    expires_at <= now();
