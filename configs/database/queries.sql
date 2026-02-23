-- name: InsertLink :exec
INSERT INTO links (short_path, original_url) VALUES ($1, $2);

-- name: GetLinkByPath :one
SELECT * FROM links
WHERE short_path = $1;