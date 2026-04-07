package web

import (
	"encoding/json"
	"net/http"
)

// Standardised API response format
type Response struct {
	Data  any    `json:"data"`
	Error *Error `json:"error"`
}

type Error struct {
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: data, Error: nil})
}

func ErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: nil, Error: &Error{Message: message}})
}
