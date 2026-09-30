// Command golearn is the GoLearn API server (and static frontend host).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valentrahq/golearn/internal/api"
	"github.com/valentrahq/golearn/internal/config"
	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/runner"
	"github.com/valentrahq/golearn/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cat, err := content.Load()
	if err != nil {
		return err
	}
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	var rn runner.Runner = runner.Disabled{}
	switch cfg.Runner {
	case "docker":
		rn = runner.NewDocker(runner.DockerConfig{
			Bin: cfg.DockerBin, Image: cfg.RunnerImage, MemoryMB: cfg.RunMemoryMB, CPUs: cfg.RunCPUs,
			PIDs: cfg.RunPIDs, Concurrency: cfg.RunConcurrent, Runtime: cfg.DockerRuntime,
		})
	case "remote":
		rn = runner.NewRemote(cfg.RunnerURL, cfg.RunnerToken)
	}
	if s := rn.Status(ctx); s.Available {
		slog.Info("code runner ready", "backend", s.Backend)
	} else {
		slog.Warn("code runner unavailable; lessons remain readable but code cannot be executed", "reason", s.Reason)
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.New(cfg, st, cat, rn).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      2 * time.Minute, // code runs can take a while
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr, "modules", len(cat.Modules), "lessons", len(cat.PublishedLessons()), "challenges", len(cat.Challenges))
		errc <- srv.ListenAndServe()
	}()
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
		sctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		return srv.Shutdown(sctx)
	}
	return nil
}
