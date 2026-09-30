package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerArgsEnforceIsolation(t *testing.T) {
	d := NewDocker(DockerConfig{Bin: "docker", Image: "img", MemoryMB: 512, CPUs: "1", PIDs: 256, Concurrency: 1})
	args := strings.Join(d.Args("n"), " ")
	for _, want := range []string{
		"--network none", "--read-only", "--cap-drop ALL", "--security-opt no-new-privileges",
		"--memory 512m", "--memory-swap 512m", "--cpus 1", "--pids-limit 256", "--user 65534:65534", "--rm",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("docker args missing %q: %s", want, args)
		}
	}
	for _, bad := range []string{"--privileged", "-v ", "--volume", "--mount", "--network host", "docker.sock"} {
		if strings.Contains(args, bad) {
			t.Errorf("docker args must not contain %q", bad)
		}
	}
}

func TestParseTestJSON(t *testing.T) {
	out := `{"Action":"run","Test":"TestA"}
{"Action":"output","Test":"TestA","Output":"=== RUN   TestA\n"}
{"Action":"output","Test":"TestA","Output":"    a_test.go:9: got 1 want 2\n"}
{"Action":"fail","Test":"TestA"}
{"Action":"run","Test":"TestB"}
{"Action":"run","Test":"TestB/sub"}
{"Action":"fail","Test":"TestB/sub"}
{"Action":"pass","Test":"TestB"}
{"Action":"run","Test":"TestC"}
{"Action":"pass","Test":"TestC"}
not json
`
	got := ParseTestJSON(out)
	if len(got) != 3 {
		t.Fatalf("want 3 results, got %+v", got)
	}
	if got[0].Passed || !strings.Contains(got[0].Output, "got 1 want 2") || strings.Contains(got[0].Output, "=== RUN") {
		t.Errorf("TestA: %+v", got[0])
	}
	if got[1].Passed {
		t.Errorf("TestB should fail because a subtest failed: %+v", got[1])
	}
	if !got[2].Passed {
		t.Errorf("TestC should pass: %+v", got[2])
	}
}

func TestDisabledRunner(t *testing.T) {
	var r Runner = Disabled{}
	if r.Status(context.Background()).Available {
		t.Fatal("disabled runner must not be available")
	}
	if _, err := r.Run(context.Background(), Request{}); err != ErrUnavailable {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

func TestUnavailableWhenDockerMissing(t *testing.T) {
	d := NewDocker(DockerConfig{Bin: "/nonexistent/docker", Image: "img", MemoryMB: 64, CPUs: "1", PIDs: 8, Concurrency: 1})
	st := d.Status(context.Background())
	if st.Available || st.Reason == "" {
		t.Fatalf("expected unavailable with reason, got %+v", st)
	}
	if _, err := d.Run(context.Background(), Request{Kind: KindRun}); err == nil {
		t.Fatal("expected error")
	}
}

// TestPipelineWithFakeDocker exercises the host<->entrypoint protocol. A shim
// stands in for the docker CLI and executes the entrypoint directly; this only
// happens in tests on a developer/CI machine, never in the application.
func TestPipelineWithFakeDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "entry")
	build := exec.Command("go", "build", "-o", bin, "github.com/valentrahq/golearn/cmd/sandbox-entrypoint")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build entrypoint: %v\n%s", err, out)
	}
	work := filepath.Join(dir, "work")
	_ = os.MkdirAll(work, 0o755)
	shim := filepath.Join(dir, "docker")
	script := `#!/bin/sh
case "$1" in
  image) exit 0;;
  rm) exit 0;;
  run) GOLEARN_WORK_DIR="` + work + `" GOLEARN_GOCACHE="` + filepath.Join(dir, "cache") + `" GOPROXY=off exec "` + bin + `";;
esac
`
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	d := NewDocker(DockerConfig{Bin: shim, Image: "img", MemoryMB: 64, CPUs: "1", PIDs: 8, Concurrency: 1})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("run", func(t *testing.T) {
		res, err := d.Run(ctx, Request{Kind: KindRun, TimeoutSec: 5, Files: map[string]string{
			"main.go": "package main\nimport \"fmt\"\nfunc main(){ fmt.Println(\"hi\") }\n"}})
		if err != nil || res.ExitCode != 0 || res.Stdout != "hi\n" {
			t.Fatalf("res=%+v err=%v", res, err)
		}
	})
	t.Run("build error", func(t *testing.T) {
		res, err := d.Run(ctx, Request{Kind: KindRun, TimeoutSec: 5, Files: map[string]string{
			"main.go": "package main\nfunc main(){ x := 1 }\n"}})
		if err != nil || !res.BuildFailed || !strings.Contains(res.Stderr, "declared and not used") {
			t.Fatalf("res=%+v err=%v", res, err)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		res, err := d.Run(ctx, Request{Kind: KindRun, TimeoutSec: 1, Files: map[string]string{
			"main.go": "package main\nfunc main(){ for {} }\n"}})
		if err != nil || !res.TimedOut {
			t.Fatalf("res=%+v err=%v", res, err)
		}
	})
	t.Run("output cap", func(t *testing.T) {
		res, err := d.Run(ctx, Request{Kind: KindRun, TimeoutSec: 5, Files: map[string]string{
			"main.go": "package main\nimport (\"fmt\";\"strings\")\nfunc main(){ for i:=0;i<20000;i++ { fmt.Println(strings.Repeat(\"x\", 100)) } }\n"}})
		if err != nil || !res.Truncated || len(res.Stdout) > 256<<10 {
			t.Fatalf("truncated=%v len=%d err=%v", res.Truncated, len(res.Stdout), err)
		}
	})
	t.Run("tests", func(t *testing.T) {
		res, err := d.Run(ctx, Request{Kind: KindTest, TimeoutSec: 10, Files: map[string]string{
			"main.go":      "package main\nfunc Add(a, b int) int { return a + b }\nfunc main() {}\n",
			"main_test.go": "package main\nimport \"testing\"\nfunc TestAdd(t *testing.T){ if Add(1,2)!=3 { t.Fatal(\"bad\") } }\nfunc TestBad(t *testing.T){ t.Fatal(\"nope\") }\n"}})
		tests := ParseTestJSON(res.Stdout)
		if err != nil || len(tests) != 2 || !tests[0].Passed || tests[1].Passed {
			t.Fatalf("res=%+v tests=%+v err=%v", res, tests, err)
		}
	})
	t.Run("bad file name", func(t *testing.T) {
		if _, err := d.Run(ctx, Request{Kind: KindRun, TimeoutSec: 5, Files: map[string]string{"../evil.go": "package main"}}); err == nil {
			t.Fatal("expected the request to be rejected before any container starts")
		}
	})
}

func TestRequestValidate(t *testing.T) {
	ok := Request{Kind: KindRun, Files: map[string]string{"main.go": "package main"}}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	bad := map[string]Request{
		"unknown kind":   {Kind: "sh", Files: ok.Files},
		"no files":       {Kind: KindRun},
		"traversal":      {Kind: KindRun, Files: map[string]string{"../x.go": "x"}},
		"absolute":       {Kind: KindRun, Files: map[string]string{"/etc/x.go": "x"}},
		"not go":         {Kind: KindRun, Files: map[string]string{"run.sh": "x"}},
		"uppercase":      {Kind: KindRun, Files: map[string]string{"Main.go": "x"}},
		"too large":      {Kind: KindRun, Files: map[string]string{"main.go": strings.Repeat("a", maxTotalBytes+1)}},
		"too many files": {Kind: KindRun, Files: map[string]string{"a.go": "", "b.go": "", "c.go": "", "d.go": "", "e.go": ""}},
	}
	for name, r := range bad {
		if err := r.Validate(); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}
