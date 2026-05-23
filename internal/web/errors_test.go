package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("duplicate key")

func TestHandleError(t *testing.T) {
	tests := []struct {
		name                 string
		handler              Handler
		expectedStatus       int
		expectedErrorMessage string
		expectedLogLevel     slog.Level
	}{
		{
			name: "no error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				JSON(w, http.StatusOK, "any")
				return nil
			},
			expectedStatus:       http.StatusOK,
			expectedErrorMessage: "",
			expectedLogLevel:     slog.LevelInfo, // no error logged
		},
		{
			name: "generic error returns 500 with standard text",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("database connection refused")
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedErrorMessage: http.StatusText(http.StatusInternalServerError),
			expectedLogLevel:     slog.LevelError,
		},
		{
			name: "trusted 4xx exposes custom message",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return NewRequestError(http.StatusNotFound, ErrNotFound, Trusted)
			},
			expectedStatus:       http.StatusNotFound,
			expectedErrorMessage: ErrNotFound.Error(),
			expectedLogLevel:     slog.LevelWarn,
		},
		{
			name: "untrusted 4xx returns generic status text",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return NewRequestError(http.StatusBadRequest, fmt.Errorf("decode failed"), Untrusted)
			},
			expectedStatus:       http.StatusBadRequest,
			expectedErrorMessage: http.StatusText(http.StatusBadRequest),
			expectedLogLevel:     slog.LevelWarn,
		},
		{
			name: "wrapped RequestError is detected by errors.As",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				re := NewRequestError(http.StatusConflict, ErrConflict, Trusted)
				return fmt.Errorf("there was a child error: %w", re)
			},
			expectedStatus:       http.StatusConflict,
			expectedErrorMessage: ErrConflict.Error(),
			expectedLogLevel:     slog.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))

			h := HandleError(logger, tt.handler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			h.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var envelope Response
			if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
				t.Fatalf("decode response body: %v", err)
			}

			if tt.expectedErrorMessage == "" {
				if envelope.Error != nil {
					t.Fatalf("expected no error body, got %+v", envelope.Error)
				}
			} else {
				if envelope.Error == nil {
					t.Fatal("expected error body, got nil")
				}
				if envelope.Error.Message != tt.expectedErrorMessage {
					t.Fatalf("expected error message %q, got %q", tt.expectedErrorMessage, envelope.Error.Message)
				}
			}

			// Check log levels are working correctly
			logOutput := buf.String()
			switch tt.expectedLogLevel {
			case slog.LevelError:
				if !strings.Contains(logOutput, `"level":"ERROR"`) {
					t.Fatal("expected ERROR level log")
				}
			case slog.LevelWarn:
				if !strings.Contains(logOutput, `"level":"WARN"`) {
					t.Fatal("expected WARN level log")
				}
			}
		})
	}
}
