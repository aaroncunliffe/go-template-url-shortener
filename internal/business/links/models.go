package links

// Coded error for low cardinality label support in the obseravility stack
type CodedError struct {
	code    string
	message string
}

func (e CodedError) Error() string { return e.message }
func (e CodedError) Code() string  { return e.code }

// Package specific business errors
var (
	ErrNotFound = CodedError{code: "not_found", message: "link not found"}
	ErrConflict = CodedError{code: "conflict", message: "link already exists with that short_path"}
)

type Link struct {
	ShortPath   string
	OriginalURL string
}
