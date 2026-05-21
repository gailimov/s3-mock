package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const logLevel = slog.LevelInfo

const storageDir = "./storage"

const (
	addr              = ":8080"
	readHeaderTimeout = 1 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}))

	if err := os.MkdirAll(storageDir, 0o700); err != nil {
		logger.Error(
			"failed to create storage directory",
			"err", err,
		)

		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /", handleGet(logger, storageDir))
	mux.HandleFunc("PUT /", handlePut(logger, storageDir))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Info("starting S3 mock server")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"failed to start server",
				"err", err,
				"addr", addr,
			)

			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit

	logger.Info("shutting down S3 mock server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(
			"failed to gracefully shut down server",
			"err", err,
		)

		return
	}

	logger.Info("server stopped")
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleGet(logger *slog.Logger, storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := buildPath(storageDir, r.URL.Path)

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			logger.Error(
				"failed to get file info",
				"err", err,
				"path", path,
			)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if info.IsDir() {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		http.ServeFile(w, r, path)
	}
}

func handlePut(logger *slog.Logger, storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Allow", "GET")
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		path := buildPath(storageDir, r.URL.Path)

		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			logger.Error(
				"failed to create parent directory",
				"err", err,
				"path", path,
			)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		//nolint:gosec // path is normalized and scoped to storageDir in buildPath
		f, err := os.Create(path)
		if err != nil {
			logger.Error(
				"failed to create file",
				"err", err,
				"path", path,
			)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				logger.Error(
					"failed to close file",
					"err", closeErr,
					"path", path,
				)
			}
		}()

		_, err = io.Copy(f, r.Body)
		if err != nil {
			logger.Error(
				"failed to write file",
				"err", err,
				"path", path,
			)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func buildPath(storageDir, urlPath string) string {
	return filepath.Join(storageDir, filepath.Clean(urlPath))
}
