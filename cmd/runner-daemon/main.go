// Command runner-daemon exposes the Docker sandbox runner over an
// authenticated HTTP API. It is the only GoLearn component that needs access
// to the container runtime, so it should run on an isolated network/node and
// never be exposed to the internet.
package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/valentrahq/golearn/internal/runner"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if n, err := strconv.Atoi(os.Getenv(k)); err == nil {
		return n
	}
	return d
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	token := os.Getenv("GOLEARN_RUNNER_TOKEN")
	if len(token) < 16 {
		slog.Error("GOLEARN_RUNNER_TOKEN must be set to a random string of at least 16 characters")
		os.Exit(1)
	}
	d := runner.NewDocker(runner.DockerConfig{
		Bin: env("GOLEARN_DOCKER_BIN", "docker"), Image: env("GOLEARN_RUNNER_IMAGE", "golearn-runner:latest"),
		MemoryMB: envInt("GOLEARN_RUN_MEMORY_MB", 512), CPUs: env("GOLEARN_RUN_CPUS", "1"),
		PIDs: envInt("GOLEARN_RUN_PIDS", 256), Concurrency: envInt("GOLEARN_RUN_CONCURRENCY", 4),
		Runtime: env("GOLEARN_DOCKER_RUNTIME", ""),
	})

	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("GET /v1/status", auth(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(d.Status(r.Context()))
	}))
	mux.HandleFunc("POST /v1/run", auth(func(w http.ResponseWriter, r *http.Request) {
		var req runner.Request
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Kind != runner.KindRun && req.Kind != runner.KindTest {
			http.Error(w, "bad kind", http.StatusBadRequest)
			return
		}
		if req.TimeoutSec < 1 || req.TimeoutSec > 30 {
			req.TimeoutSec = 10
		}
		res, err := d.Run(r.Context(), req)
		switch {
		case errors.Is(err, runner.ErrBusy):
			http.Error(w, "busy", http.StatusServiceUnavailable)
		case err != nil:
			slog.Error("run failed", "err", err)
			http.Error(w, "runner error", http.StatusBadGateway)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
		}
	}))

	srv := &http.Server{
		Addr: env("GOLEARN_RUNNER_ADDR", ":9090"), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 2 * time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	slog.Info("runner daemon listening", "addr", srv.Addr, "image", env("GOLEARN_RUNNER_IMAGE", "golearn-runner:latest"))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
