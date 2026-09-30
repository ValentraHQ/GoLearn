// Command sandbox-entrypoint runs inside the GoLearn runner container. It reads
// one JSON runner.Request from stdin, builds and runs the files, and writes one
// JSON runner.Result to stdout. It must only ever run inside the sandbox
// container (no network, read-only rootfs, dropped capabilities).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/valentrahq/golearn/internal/runner"
)

const (
	maxInput   = 128 << 10
	maxOut     = 200 << 10
	buildLimit = 40 * time.Second
)

var fileRe = runner.FileNameRe

func main() {
	res := handle()
	_ = json.NewEncoder(os.Stdout).Encode(res)
}

func fail(msg string) runner.Result {
	return runner.Result{ExitCode: 1, Stderr: msg}
}

func handle() runner.Result {
	var req runner.Request
	if err := json.NewDecoder(io.LimitReader(os.Stdin, maxInput)).Decode(&req); err != nil {
		return fail("invalid request: " + err.Error())
	}
	work := envOr("GOLEARN_WORK_DIR", "/work")
	cache := envOr("GOLEARN_GOCACHE", "/tmp/gocache")
	seed := envOr("GOLEARN_GOCACHE_SEED", "/opt/gocache")
	if _, err := os.Stat(seed); err == nil {
		_ = copyTree(seed, cache)
	}
	for name, body := range req.Files {
		if !fileRe.MatchString(name) {
			return fail("invalid file name: " + name)
		}
		if err := os.WriteFile(filepath.Join(work, name), []byte(body), 0o644); err != nil {
			return fail(err.Error())
		}
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), []byte("module sandbox\n\ngo 1.24\n"), 0o644); err != nil {
		return fail(err.Error())
	}
	env := append(os.Environ(),
		"HOME=/tmp", "GOCACHE="+cache, "GOPATH=/tmp/gopath", "GOPROXY=off", "GOTOOLCHAIN=local",
		"GOFLAGS=-buildvcs=false", "CGO_ENABLED=0", "GO111MODULE=on", "GONOSUMDB=*", "GOTELEMETRY=off")
	timeout := time.Duration(req.TimeoutSec) * time.Second
	if timeout <= 0 || timeout > 30*time.Second {
		timeout = 10 * time.Second
	}

	switch req.Kind {
	case runner.KindRun:
		build := exec.Command("go", "build", "-o", filepath.Join(work, "prog"), ".")
		res := execute(build, work, env, buildLimit)
		if res.ExitCode != 0 || res.TimedOut {
			res.BuildFailed = true
			return res
		}
		return execute(exec.Command(filepath.Join(work, "prog")), work, env, timeout)
	case runner.KindTest:
		cmd := exec.Command("go", "test", "-json", "-count=1", "-timeout", strconv.Itoa(int(timeout.Seconds()))+"s", ".")
		res := execute(cmd, work, env, buildLimit+timeout)
		return res
	}
	return fail("unknown kind")
}

func execute(cmd *exec.Cmd, dir string, env []string, limit time.Duration) runner.Result {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	cmd.Dir, cmd.Env = dir, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	var so, se capped
	so.max, se.max = maxOut, 64<<10
	cmd.Stdout, cmd.Stderr = &so, &se
	start := time.Now()
	err := cmd.Run()
	res := runner.Result{
		Stdout: so.buf.String(), Stderr: se.buf.String(), Truncated: so.over || se.over,
		DurationMS: time.Since(start).Milliseconds(),
	}
	if ctx.Err() != nil {
		res.TimedOut = true
		res.ExitCode = -1
		return res
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				// Typically the kernel OOM killer when the memory limit is hit.
				res.Stderr += fmt.Sprintf("\nProcess was killed by signal %q (often the memory limit).\n", ws.Signal())
			}
		} else {
			res.ExitCode = 1
			res.Stderr += err.Error()
		}
	}
	return res
}

type capped struct {
	buf  bytes.Buffer
	max  int
	over bool
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); room < len(p) {
		if room > 0 {
			c.buf.Write(p[:room])
		}
		c.over = true
		return len(p), nil
	}
	return c.buf.Write(p)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
