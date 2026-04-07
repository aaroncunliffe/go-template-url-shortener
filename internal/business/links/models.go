package links

import "errors"

var (
	ErrNotFound = errors.New("link not found")
	ErrConflict = errors.New("link already exists with that short_path")
)

type Link struct {
	ShortPath   string
	OriginalURL string
}
