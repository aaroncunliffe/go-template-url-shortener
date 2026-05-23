package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

func ValidRedirectURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme %q not allowed", parsed.Scheme)
	}
	return nil
}
