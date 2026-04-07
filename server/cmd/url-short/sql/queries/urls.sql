-- name: CreateShortURL :one
INSERT INTO urls (original_url, short_url_code) VALUES ($1, $2) RETURNING *;


-- name: GetShortURL :one
SELECT * FROM urls WHERE short_url_code = $1;