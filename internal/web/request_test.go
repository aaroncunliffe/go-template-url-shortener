package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type decodePayload struct {
	Name string `json:"name"`
}

// Basic isolated unit tests
func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expectError  bool
		expectedName string
	}{
		{
			name:         "accepts single object",
			body:         `{"name":"test"}`,
			expectedName: "test",
		},
		{
			name:        "rejects unknown fields",
			body:        `{"name":"test","extra":true}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(tt.body))
			var payload decodePayload

			err := DecodeJSON(req, &payload)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected decode error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected request to decode, got error %v", err)
			}

			if payload.Name != tt.expectedName {
				t.Fatalf("expected decoded name %q, got %q", tt.expectedName, payload.Name)
			}
		})
	}
}
