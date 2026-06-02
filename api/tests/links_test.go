package links_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"

	apilinks "github.com/aaroncunliffe/go-template-url-shortener/api/links"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"
)

// Currently supported characters
// Explicitly not exported from the package, these are black box (ish) integration tests
// intended to test the contract for the API
var validShortCode = regexp.MustCompile(`^[a-zA-Z0-9_-]{6}$`)

// Expected API response contract
type responseEnvelope[T any] struct {
	Data  T          `json:"data"`
	Error *web.Error `json:"error"`
}

const apiPath = "/api/link"

func TestCreateLinks(t *testing.T) {
	t.Run("CreateLink", func(t *testing.T) {
		env := integrationEnvForTest(t)

		tests := []struct {
			name                   string
			body                   string
			seed                   *database.InsertLinkParams
			expectedStatus         int
			expectedShortPath      string
			expectErrorBody        bool
			expectedErrorBody      string
			expectGenericErrorBody bool
		}{
			{
				name:              "success with provided short_path",
				body:              `{"short_path":"abcd","origin_url":"https://aaroncunliff.dev"}`,
				expectedStatus:    http.StatusCreated,
				expectedShortPath: "abcd",
			},
			{
				name:              "success with auto-generated short_path",
				body:              `{"origin_url":"https://aaroncunliff.dev"}`,
				expectedStatus:    http.StatusCreated,
				expectedShortPath: "", // generated, validated separately
			},
			{
				name:            "validation rejects slash in short_path",
				body:            `{"short_path":"bad/path","origin_url":"https://aaroncunliff.dev"}`,
				expectedStatus:  http.StatusBadRequest,
				expectErrorBody: true,
			},
			{
				name:            "validation rejects special chars in short_path",
				body:            `{"short_path":"hi!@#","origin_url":"https://aaroncunliff.dev"}`,
				expectedStatus:  http.StatusBadRequest,
				expectErrorBody: true,
			},
			{
				name:            "validation rejects invalid redirect",
				body:            `{"short_path":"abcd","origin_url":"notvalidurl"}`,
				expectedStatus:  http.StatusBadRequest,
				expectErrorBody: true,
			},
			{
				name: "link already taken",
				body: `{"short_path":"already-taken","origin_url":"https://aaroncunliffe.dev"}`,
				seed: &database.InsertLinkParams{
					ShortPath:   "already-taken",
					OriginalUrl: "https://aaroncunliffe.dev",
				},
				expectedStatus:  http.StatusConflict,
				expectErrorBody: true,
			},
			{
				name:                   "malformed request gives generic error",
				body:                   `{"short_path":"`,
				expectedStatus:         http.StatusBadRequest,
				expectErrorBody:        true,
				expectGenericErrorBody: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				// Seed specific test data
				if tt.seed != nil {
					if err := env.queries.InsertLink(env.ctx, *tt.seed); err != nil {
						t.Fatalf("seed link: %v", err)
					}
				}

				req, err := http.NewRequestWithContext(env.ctx, http.MethodPost, env.server.URL+apiPath, strings.NewReader(tt.body))
				if err != nil {
					t.Fatalf("build request: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := env.server.Client().Do(req)
				if err != nil {
					t.Fatalf("perform request: %v", err)
				}
				defer resp.Body.Close()

				// Status code check
				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				// Error body check
				if tt.expectErrorBody {
					var body responseEnvelope[any]
					if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
						t.Fatalf("decode error response: %v", err)
					}

					if body.Error == nil || body.Error.Message == "" {
						t.Fatal("expected error response body")
					}

					if tt.expectGenericErrorBody && body.Error.Message != http.StatusText(resp.StatusCode) {
						t.Fatalf("expected generic error message with text %q, got %q",
							http.StatusText(resp.StatusCode),
							body.Error.Message,
						)
					}

					return
				}

				var body responseEnvelope[apilinks.CreateLinkResponse]
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("decode success response: %v", err)
				}

				if body.Error != nil {
					t.Fatalf("expected nil error, got %+v", body.Error)
				}

				if tt.expectedShortPath != "" && body.Data.ShortPath != tt.expectedShortPath {
					t.Fatalf("expected short path %q, got %q", tt.expectedShortPath, body.Data.ShortPath)
				}

				if tt.expectedShortPath == "" && !validShortCode.MatchString(body.Data.ShortPath) {
					t.Fatalf("expected auto-generated short path matching %s, got %q", validShortCode.String(), body.Data.ShortPath)
				}
			})
		}
	})
}

func TestLinksRedirect(t *testing.T) {
	t.Run("LinkRedirect", func(t *testing.T) {
		env := integrationEnvForTest(t)

		tests := []struct {
			name             string
			path             string
			seed             *database.InsertLinkParams
			expectedStatus   int
			expectedLocation string
		}{
			{
				name: "success",
				path: "redirect-success",
				seed: &database.InsertLinkParams{
					ShortPath:   "redirect-success",
					OriginalUrl: "https://aaroncunliffe.dev/redirect",
				},
				expectedStatus:   http.StatusFound,
				expectedLocation: "https://aaroncunliffe.dev/redirect",
			},
			{
				name:           "not found",
				path:           "missing-link",
				expectedStatus: http.StatusNotFound,
			},
		}

		client := env.server.Client()

		// Block redirects to be able to test header value
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				// Seed specific test data
				if tt.seed != nil {
					if err := env.queries.InsertLink(env.ctx, *tt.seed); err != nil {
						t.Fatalf("seed link: %v", err)
					}
				}

				req, err := http.NewRequestWithContext(env.ctx, http.MethodGet, env.server.URL+"/"+tt.path, nil)
				if err != nil {
					t.Fatalf("build request: %v", err)
				}

				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("perform request: %v", err)
				}
				defer resp.Body.Close()

				// Status code check
				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				// If redirect is expected
				if tt.expectedLocation != "" {
					if location := resp.Header.Get("Location"); location != tt.expectedLocation {
						t.Fatalf("expected redirect location %q, got %q", tt.expectedLocation, location)
					}
				}
			})
		}
	})
}
