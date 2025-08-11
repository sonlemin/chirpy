-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    (gen_random_uuid(), NEW(), NEW(), $1)
)
RETURNING *;
