package web

import (
	"errors"
	"log/slog"
	"net/http"
)

// RequestError is an error that is safe to return to the user
// with a specific HTTP status code and message.
type RequestError struct {
	Status    int
	Message   string
	IsTrusted bool
}

const (
	Trusted   = true
	Untrusted = false
)

func (e *RequestError) Error() string { return e.Message }

func NewRequestError(status int, err error, isTrusted bool) error {
	return &RequestError{
		Status:    status,
		Message:   err.Error(),
		IsTrusted: isTrusted,
	}
}

// HandleError adapts a Handler into an http.Handler.
// For centralised Trusted error handling
func HandleError(logger *slog.Logger, h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		// If not a known request error, log full error and return generic error to user
		var re *RequestError
		if !errors.As(err, &re) {
			logger.ErrorContext(r.Context(), "handler error", "error", err)
			ErrorJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		// Request error - log internally
		if re.Status >= http.StatusInternalServerError {
			logger.ErrorContext(r.Context(), "handler error", "status", re.Status, "error", err)
		} else {
			logger.WarnContext(r.Context(), "handler error", "status", re.Status, "error", err)
		}

		// Then decide how to display this to the user
		if re.IsTrusted {
			ErrorJSON(w, re.Status, re.Message)
		} else {
			// Generic response based on status
			ErrorJSON(w, re.Status, http.StatusText(re.Status))
		}
	}
}
