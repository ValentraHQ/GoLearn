package runner

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxOutput = 256 << 10 // bytes read back from a container

type DockerConfig struct {
	Bin         string
	Image       string
	MemoryMB    int
	CPUs        string
	PIDs        int
	Concurrency int
	QueueWait   time.Duration
	// Runtime selects an alternative OCI runtime such as gVisor's runsc.
	Runtime string
}

type Docker struct {
	cfg  DockerConfig
	sem  chan struct{}
	mu   sync.Mutex
	last Status
	at   time.Time
}

func NewDocker(cfg DockerConfig) *Docker {
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.QueueWait == 0 {
		cfg.QueueWait = 10 * time.Second
	}
	return &Docker{cfg: cfg, sem: make(chan struct{}, cfg.Concurrency)}
}

// Args builds the docker run arguments. Exposed for tests: these flags are the
// security boundary of the code runner.
func (d *Docker) Args(name string) []string {
	mem := strconv.Itoa(d.cfg.MemoryMB) + "m"
	args := []string{
		"run", "--rm", "-i",
		"--name", name,
		"--network", "none",
		"--read-only",
		"--tmpfs", "/work:rw,exec,nosuid,nodev,size=64m,uid=65534,gid=65534",
		"--tmpfs", "/tmp:rw,exec,nosuid,nodev,size=384m,uid=65534,gid=65534",
		"--memory", mem, "--memory-swap", mem,
		"--cpus", d.cfg.CPUs,
		"--pids-limit", strconv.Itoa(d.cfg.PIDs),
		"--ulimit", "nofile=256:256",
		"--ulimit", "core=0",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--user", "65534:65534",
		"--stop-timeout", "1",
	}
	if d.cfg.Runtime != "" {
		args = append(args, "--runtime", d.cfg.Runtime)
	}
	return append(args, d.cfg.Image)
}

func (d *Docker) Status(ctx context.Context) Status {
	d.mu.Lock()
	defer d.mu.Unlock()
	if time.Since(d.at) < 15*time.Second {
		return d.last
	}
	st := Status{Backend: "docker"}
	cctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(cctx, d.cfg.Bin, "image", "inspect", "--format", "{{.Id}}", d.cfg.Image).CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		switch {
		case strings.Contains(msg, "No such image"):
			st.Reason = fmt.Sprintf("Runner image %q is not built. Run `make runner-image`.", d.cfg.Image)
		case msg == "" && cctx.Err() != nil:
			st.Reason = "Docker did not respond in time."
		default:
			st.Reason = "Docker is not reachable by the server."
		}
	} else {
		st.Available = true
	}
	d.last, d.at = st, time.Now()
	return st
}

func (d *Docker) Run(ctx context.Context, req Request) (Result, error) {
	if err := req.Validate(); err != nil {
		return Result{}, fmt.Errorf("invalid request: %w", err)
	}
	if st := d.Status(ctx); !st.Available {
		return Result{}, fmt.Errorf("%w: %s", ErrUnavailable, st.Reason)
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 10
	}
	wait := time.NewTimer(d.cfg.QueueWait)
	defer wait.Stop()
	select {
	case d.sem <- struct{}{}:
		defer func() { <-d.sem }()
	case <-wait.C:
		return Result{}, ErrBusy
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return Result{}, err
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	name := "golearn-run-" + hex.EncodeToString(b)

	// Build (cold cache copy + compile) gets its own allowance on top of the run limit.
	deadline := time.Duration(req.TimeoutSec)*time.Second + 45*time.Second
	rctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	cmd := exec.CommandContext(rctx, d.cfg.Bin, d.Args(name)...)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	cmd.Stdout = &limitedWriter{w: &stdout, n: maxOutput}
	cmd.Stderr = &limitedWriter{w: &bytes.Buffer{}, n: 4096}
	start := time.Now()
	runErr := cmd.Run()

	// Always make sure the container is gone, even after a timeout or crash.
	killCtx, kcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer kcancel()
	_ = exec.CommandContext(killCtx, d.cfg.Bin, "rm", "-f", name).Run()

	if rctx.Err() != nil {
		return Result{TimedOut: true, ExitCode: -1, DurationMS: time.Since(start).Milliseconds(),
			Stderr: "Execution exceeded the time limit and was terminated."}, nil
	}
	var res Result
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		if runErr != nil {
			return Result{}, fmt.Errorf("runner failed: %w", runErr)
		}
		return Result{}, fmt.Errorf("runner returned malformed output: %w", err)
	}
	res.DurationMS = time.Since(start).Milliseconds()
	return res, nil
}

type limitedWriter struct {
	w io.Writer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return len(p), nil // discard, keep the process from blocking
	}
	if len(p) > l.n {
		_, _ = l.w.Write(p[:l.n])
		l.n = 0
		return len(p), nil
	}
	l.n -= len(p)
	return l.w.Write(p)
}
