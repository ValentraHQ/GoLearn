// Package config loads environment-based configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr         string
	DatabaseURL  string // "sqlite:path" or "postgres://..."
	StaticDir    string // built frontend; empty disables static serving
	CookieName   string
	SecureCookie bool
	SessionTTL   time.Duration
	AllowOrigin  string // optional CORS origin for split dev setups
	TrustProxy   bool   // honour X-Forwarded-For for client IPs (only behind a trusted proxy)
	BcryptCost   int

	// Code execution sandbox.
	Runner        string // "docker", "remote" or "disabled"
	RunnerURL     string // remote runner daemon base URL
	RunnerToken   string // bearer token shared with the runner daemon
	DockerRuntime string // optional OCI runtime, e.g. "runsc" (gVisor)
	DockerBin     string
	RunnerImage   string
	RunTimeout    time.Duration
	RunMemoryMB   int
	RunCPUs       string
	RunPIDs       int
	RunConcurrent int
	RunsPerMinute int
}

func Load() (Config, error) {
	c := Config{
		Addr:          env("GOLEARN_ADDR", ":8080"),
		DatabaseURL:   env("DATABASE_URL", "sqlite:golearn.db"),
		StaticDir:     env("GOLEARN_STATIC_DIR", "web/dist"),
		CookieName:    env("GOLEARN_COOKIE_NAME", "golearn_session"),
		SecureCookie:  envBool("GOLEARN_COOKIE_SECURE", false),
		SessionTTL:    envDur("GOLEARN_SESSION_TTL", 30*24*time.Hour),
		AllowOrigin:   env("GOLEARN_ALLOW_ORIGIN", ""),
		TrustProxy:    envBool("GOLEARN_TRUST_PROXY", false),
		BcryptCost:    envInt("GOLEARN_BCRYPT_COST", 12),
		Runner:        env("GOLEARN_RUNNER", "docker"),
		RunnerURL:     env("GOLEARN_RUNNER_URL", ""),
		RunnerToken:   env("GOLEARN_RUNNER_TOKEN", ""),
		DockerRuntime: env("GOLEARN_DOCKER_RUNTIME", ""),
		DockerBin:     env("GOLEARN_DOCKER_BIN", "docker"),
		RunnerImage:   env("GOLEARN_RUNNER_IMAGE", "golearn-runner:latest"),
		RunTimeout:    envDur("GOLEARN_RUN_TIMEOUT", 15*time.Second),
		RunMemoryMB:   envInt("GOLEARN_RUN_MEMORY_MB", 512),
		RunCPUs:       env("GOLEARN_RUN_CPUS", "1"),
		RunPIDs:       envInt("GOLEARN_RUN_PIDS", 256),
		RunConcurrent: envInt("GOLEARN_RUN_CONCURRENCY", 4),
		RunsPerMinute: envInt("GOLEARN_RUNS_PER_MINUTE", 30),
	}
	switch c.Runner {
	case "docker", "disabled":
	case "remote":
		if c.RunnerURL == "" || c.RunnerToken == "" {
			return c, fmt.Errorf("GOLEARN_RUNNER=remote requires GOLEARN_RUNNER_URL and GOLEARN_RUNNER_TOKEN")
		}
	default:
		return c, fmt.Errorf("GOLEARN_RUNNER must be docker, remote or disabled, got %q", c.Runner)
	}
	return c, nil
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func envBool(k string, d bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return d
}

func envDur(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if x, err := time.ParseDuration(v); err == nil {
			return x
		}
	}
	return d
}
