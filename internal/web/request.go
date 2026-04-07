package web

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

type ValidationError struct {
	Field string
	Tag   string
}

func DecodeJSON(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	return nil
}

func ValidateStruct(s any) error {
	return v.Struct(s)
}
