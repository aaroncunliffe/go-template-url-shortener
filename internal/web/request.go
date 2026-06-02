package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

var shortPathChars = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func init() {
	if err := v.RegisterValidation("shortpath", shortPath); err != nil {
		panic(err)
	}
}

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

// Custom validators
func shortPath(fl validator.FieldLevel) bool {
	return shortPathChars.MatchString(fl.Field().String())
}
