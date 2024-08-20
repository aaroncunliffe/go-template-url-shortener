-- name: InsertLink :exec
INSERT INTO links (short_path, original_url) VALUES ($1, $2);

-- name: GetLink :one
SELECT * FROM links
WHERE short_path = $1;