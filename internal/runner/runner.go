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

type Result struct {
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
	ExitCode    int    `json:"exitCode"`
	TimedOut    bool   `json:"timedOut"`
	Truncated   bool   `json:"truncated"`
	BuildFailed bool   `json:"buildFailed"`
	DurationMS  int64  `json:"durationMs"`
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
