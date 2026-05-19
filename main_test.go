package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testEnv struct {
	logger     *slog.Logger
	storageDir string
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()

	return testEnv{
		logger:     slog.New(slog.DiscardHandler),
		storageDir: t.TempDir(),
	}
}

func TestHandleGet(t *testing.T) {
	const (
		filename    = "test.txt"
		content     = "test"
		contentType = "text/plain; charset=utf-8"
	)

	env := newTestEnv(t)

	if err := os.WriteFile(filepath.Join(env.storageDir, filename), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/"+filename, nil)
	w := httptest.NewRecorder()

	handleGet(env.logger, env.storageDir).ServeHTTP(w, r)

	if got := w.Code; got != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, got)
	}

	if got := w.Header().Get("Content-Type"); got != contentType {
		t.Fatalf("expected %q, got %q", contentType, got)
	}

	if got := w.Body.String(); got != content {
		t.Fatalf("expected %q, got %q", content, got)
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		setup func(t *testing.T, env testEnv)
	}{
		{
			name: "missing file",
			path: "/missing.txt",
		},
		{
			name: "directory",
			path: "/test",
			setup: func(t *testing.T, env testEnv) {
				t.Helper()

				if err := os.Mkdir(filepath.Join(env.storageDir, "test"), 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	env := newTestEnv(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, env)
			}

			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handleGet(env.logger, env.storageDir).ServeHTTP(w, r)

			if got := w.Code; got != http.StatusNotFound {
				t.Fatalf("expected %d, got %d", http.StatusNotFound, got)
			}
		})
	}
}

func TestHandlePut(t *testing.T) {
	const content = "test"

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "root",
			filename: "test.txt",
		},
		{
			name:     "nested path",
			filename: "test/test.txt",
		},
	}

	env := newTestEnv(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPut, "/"+tt.filename, strings.NewReader(content))
			w := httptest.NewRecorder()

			handlePut(env.logger, env.storageDir).ServeHTTP(w, r)

			if got := w.Code; got != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, got)
			}

			data, err := os.ReadFile(filepath.Join(env.storageDir, tt.filename))
			if err != nil {
				t.Fatal(err)
			}

			if got := string(data); got != content {
				t.Fatalf("expected %q, got %q", content, got)
			}
		})
	}
}

func TestHandlePut_Overwrite(t *testing.T) {
	const (
		filename       = "test.txt"
		initialContent = "test"
		content        = "test2"
	)

	env := newTestEnv(t)

	assertContent := func(expected string) {
		t.Helper()

		data, err := os.ReadFile(filepath.Join(env.storageDir, filename))
		if err != nil {
			t.Fatal(err)
		}

		if got := string(data); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}

	r1 := httptest.NewRequest(http.MethodPut, "/"+filename, strings.NewReader(initialContent))
	w1 := httptest.NewRecorder()

	handlePut(env.logger, env.storageDir).ServeHTTP(w1, r1)

	if got := w1.Code; got != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, got)
	}

	assertContent(initialContent)

	r2 := httptest.NewRequest(http.MethodPut, "/"+filename, strings.NewReader(content))
	w2 := httptest.NewRecorder()

	handlePut(env.logger, env.storageDir).ServeHTTP(w2, r2)

	if got := w2.Code; got != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, got)
	}

	assertContent(content)
}
