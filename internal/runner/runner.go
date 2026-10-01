// Package runner executes learner Go code inside an isolated container.
//
// The GoLearn server never compiles or runs learner code on its own host. The
// Docker driver starts a throwaway container per request with no network, a
// read-only root filesystem, dropped capabilities, and CPU / memory / PID /
// time limits, and removes it afterwards.
package runner

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

type Kind string

const (
	KindRun  Kind = "run"  // go run
	KindTest Kind = "test" // go test
)

var (
	ErrUnavailable = errors.New("code execution is unavailable")
	ErrBusy        = errors.New("the code runner is busy, try again shortly")
)

type Request struct {
	Kind  Kind              `json:"kind"`
	Files map[string]string `json:"files"`
	// TimeoutSec bounds the run phase inside the container.
	TimeoutSec int `json:"timeoutSec"`
}

type TestResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Output string `json:"output,omitempty"`
}

// ProtocolVersion marks output produced by sandbox-entrypoint. The driver
// rejects anything without it, so a wrong or broken image (whose output could
// still be valid JSON) can never be mistaken for a successful empty run.
const ProtocolVersion = 1

type Result struct {
	Protocol    int    `json:"protocol"`
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
	ExitCode    int    `json:"exitCode"`
	TimedOut    bool   `json:"timedOut"`
	Truncated   bool   `json:"truncated"`
	BuildFailed bool   `json:"buildFailed"`
	DurationMS  int64  `json:"durationMs"`
	// Verdict is set for KindTest runs. Only an authenticated verdict (see
	// grader.go) may be used to decide that a challenge passed.
	Verdict *Verdict `json:"verdict,omitempty"`
}

type Status struct {
	Available bool   `json:"available"`
	Backend   string `json:"backend"`
	Reason    string `json:"reason,omitempty"`
}

type Runner interface {
	Status(ctx context.Context) Status
	Run(ctx context.Context, req Request) (Result, error)
}

// Disabled is used when GOLEARN_RUNNER=disabled.
type Disabled struct{}

func (Disabled) Status(context.Context) Status {
	return Status{Backend: "disabled", Reason: "Code execution is disabled on this server (GOLEARN_RUNNER=disabled)."}
}
func (Disabled) Run(context.Context, Request) (Result, error) { return Result{}, ErrUnavailable }

// FileNameRe is the only file name shape the sandbox accepts.
var FileNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,30}\.go$`)

const (
	maxFiles      = 4
	maxTotalBytes = 96 << 10
)

// Validate rejects malformed requests before any container is started. The
// sandbox entrypoint repeats the file name check as defence in depth.
func (r Request) Validate() error {
	if r.Kind != KindRun && r.Kind != KindTest {
		return fmt.Errorf("unknown kind %q", r.Kind)
	}
	if len(r.Files) == 0 || len(r.Files) > maxFiles {
		return fmt.Errorf("expected 1-%d files", maxFiles)
	}
	total := 0
	for name, body := range r.Files {
		if !FileNameRe.MatchString(name) {
			return fmt.Errorf("invalid file name %q", name)
		}
		total += len(body)
	}
	if total > maxTotalBytes {
		return errors.New("request too large")
	}
	return nil
}
