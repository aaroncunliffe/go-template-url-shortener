package web

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
)

// Handler is an HTTP handler func that returns an error.
// Return a *RequestError to control the status and message sent to the user.
// Any other error is logged internally and returned as a generic 500.
type Handler func(http.ResponseWriter, *http.Request) error

// Router wraps chi.Mux to provide Get/Post/etc. methods that accept
// the Handler signature, automatically wrapping each with HandleError.
type Router struct {
	Mux    *chi.Mux
	Logger *slog.Logger
}

func (r *Router) Get(pattern string, h Handler) {
	r.Mux.Get(pattern, HandleError(r.Logger, h))
}

func (r *Router) Post(pattern string, h Handler) {
	r.Mux.Post(pattern, HandleError(r.Logger, h))
}
