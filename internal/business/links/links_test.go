package links

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
)

type stubStore struct {
	inserts    []string
	insertErrs map[string]error
	link       database.Link
	lookupErr  error
}

func (s *stubStore) GetLinkByPath(_ context.Context, _ string) (database.Link, error) {
	if s.lookupErr != nil {
		return database.Link{}, s.lookupErr
	}
	return s.link, nil
}

func (s *stubStore) InsertLink(_ context.Context, shortPath string, _ string) error {
	s.inserts = append(s.inserts, shortPath)
	if err, ok := s.insertErrs[shortPath]; ok {
		return err
	}
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestResolveLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		storeLink   database.Link
		storeErr    error
		expectedURL string
		expectedErr error
	}{
		{
			name: "returns original url",
			storeLink: database.Link{
				OriginalUrl: "https://aaroncunliffe.dev/page",
			},
			expectedURL: "https://aaroncunliffe.dev/page",
		},
		{
			name:        "returns store error",
			storeErr:    ErrNotFound,
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			core := Core{
				Logger: testLogger(),
				Store: &stubStore{
					link:      tt.storeLink,
					lookupErr: tt.storeErr,
				},
			}

			url, err := core.ResolveLink(context.Background(), "any-path")

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if url != tt.expectedURL {
				t.Fatalf("expected url %q, got %q", tt.expectedURL, url)
			}
		})
	}
}

const (
	customPath = "custom-path"
	pathABC123 = "ABC123"
	pathXYZ789 = "XYZ789"
	pathAAAAAA = "AAAAAA"
	pathBBBBBB = "BBBBBB"
	pathCCCCCC = "CCCCCC"
	pathDDDDDD = "DDDDDD"
	pathEEEEEE = "EEEEEE"
)

func TestCreateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		inputShortPath  string
		generatedPaths  []string
		generateErr     error
		insertErrs      map[string]error
		expectedInserts []string
		expectedPath    string
		expectedErr     error
	}{
		{
			name:           "uses provided short path",
			inputShortPath: customPath,
			expectedInserts: []string{
				customPath,
			},
			expectedPath: customPath,
		},
		{
			name: "generates short path on first attempt",
			generatedPaths: []string{
				pathABC123,
			},
			expectedInserts: []string{
				pathABC123,
			},
			expectedPath: pathABC123,
		},
		{
			name: "retries when generated short path conflicts",
			generatedPaths: []string{
				pathABC123,
				pathXYZ789,
			},
			insertErrs: map[string]error{
				pathABC123: ErrConflict,
			},
			expectedInserts: []string{
				pathABC123,
				pathXYZ789,
			},
			expectedPath: pathXYZ789,
		},
		{
			name: "returns insert error when not conflict",
			generatedPaths: []string{
				pathABC123,
			},
			insertErrs: map[string]error{
				pathABC123: errors.New("database unavailable"),
			},
			expectedInserts: []string{
				pathABC123,
			},
			expectedErr: errors.New("database unavailable"),
		},
		{
			name: "fails after max generation attempts",
			generatedPaths: []string{
				pathAAAAAA,
				pathBBBBBB,
				pathCCCCCC,
				pathDDDDDD,
				pathEEEEEE,
			},
			insertErrs: map[string]error{
				pathAAAAAA: ErrConflict,
				pathBBBBBB: ErrConflict,
				pathCCCCCC: ErrConflict,
				pathDDDDDD: ErrConflict,
				pathEEEEEE: ErrConflict,
			},
			expectedInserts: []string{
				pathAAAAAA,
				pathBBBBBB,
				pathCCCCCC,
				pathDDDDDD,
				pathEEEEEE,
			},
			expectedErr: ErrShortCodeGenerationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := &stubStore{insertErrs: tt.insertErrs}
			generateCall := 0

			core := Core{
				Logger: testLogger(),
				Store:  store,

				// override generate function to use tests
				generate: func() (string, error) {
					if tt.generateErr != nil {
						return "", tt.generateErr
					}
					if generateCall >= len(tt.generatedPaths) {
						t.Fatalf("unexpected generate call %d", generateCall+1)
					}
					path := tt.generatedPaths[generateCall]
					generateCall++
					return path, nil
				},
			}

			path, err := core.CreateLink(context.Background(), tt.inputShortPath, "https://example.com")

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.expectedErr.Error() {
					t.Fatalf("expected error %q, got %q", tt.expectedErr.Error(), err.Error())
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if path != tt.expectedPath {
				t.Fatalf("expected path %q, got %q", tt.expectedPath, path)
			}

			if !reflect.DeepEqual(store.inserts, tt.expectedInserts) {
				t.Fatalf("expected inserts %v, got %v", tt.expectedInserts, store.inserts)
			}
		})
	}
}
