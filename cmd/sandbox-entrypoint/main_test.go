package main

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestExecuteReportsSignalKill(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh")
	}
	res := execute(exec.Command(sh, "-c", "kill -9 $$"), t.TempDir(), nil, 5*time.Second)
	if res.ExitCode != -1 || !strings.Contains(res.Stderr, "killed by signal") {
		t.Fatalf("want a kill explanation, got exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
}

func TestExecuteTimeout(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh")
	}
	res := execute(exec.Command(sh, "-c", "sleep 30"), t.TempDir(), nil, 300*time.Millisecond)
	if !res.TimedOut {
		t.Fatalf("expected timeout, got %+v", res)
	}
}
