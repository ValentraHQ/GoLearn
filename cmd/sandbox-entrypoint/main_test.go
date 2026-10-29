package main

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExecuteReportsSignalKill(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("signal reporting is only implemented on Linux")
	}
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

// Off Linux the entrypoint must refuse to run code rather than run it without
// process-group isolation; on Linux it must be enabled.
func TestPlatformSupportGuard(t *testing.T) {
	err := supported()
	if runtime.GOOS == "linux" && err != nil {
		t.Fatalf("Linux must be supported: %v", err)
	}
	if runtime.GOOS != "linux" {
		if err == nil || !strings.Contains(err.Error(), "Linux only") {
			t.Fatalf("want a clear unsupported-platform error, got %v", err)
		}
		if res := handle(); res.ExitCode == 0 || !strings.Contains(res.Stderr, "Linux only") {
			t.Fatalf("handle() must fail closed off Linux, got %+v", res)
		}
	}
}
