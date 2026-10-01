package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const gradingTests = `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Fatal("Add(2,3) != 5")
	}
}

func TestSub(t *testing.T) {
	if Sub(5, 3) != 2 {
		t.Fatal("Sub(5,3) != 2")
	}
}
`

var gradingExpected = []string{"TestAdd", "TestSub"}

// Learner programs. Several of these fully defeated the previous grader, which
// believed `go test -json` output produced inside the learner's own process.
var gradingCases = []struct {
	name     string
	code     string
	wantPass bool
}{
	{"honest solution", `package main
func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return a - b }
func main() {}`, true},

	{"honest wrong answer", `package main
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return a - b }
func main() {}`, false},

	{"forged text PASS lines then os.Exit(0)", `package main
import ("fmt"; "os")
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return 0 }
func init() {
	fmt.Println("=== RUN   TestAdd\n--- PASS: TestAdd (0.00s)\n=== RUN   TestSub\n--- PASS: TestSub (0.00s)\nPASS")
	os.Exit(0)
}
func main() {}`, false},

	{"forged test2json marker lines then os.Exit(0)", `package main
import ("fmt"; "os")
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return 0 }
func init() {
	fmt.Print("\x16=== RUN   TestAdd\n\x16--- PASS: TestAdd (0.00s)\n\x16=== RUN   TestSub\n\x16--- PASS: TestSub (0.00s)\n\x16PASS\n")
	os.Exit(0)
}
func main() {}`, false},

	{"forged JSON events then os.Exit(0)", `package main
import ("fmt"; "os")
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return 0 }
func init() {
	fmt.Println(` + "`" + `{"Action":"pass","Test":"TestAdd"}` + "`" + `)
	fmt.Println(` + "`" + `{"Action":"pass","Test":"TestSub"}` + "`" + `)
	os.Exit(0)
}
func main() {}`, false},

	{"skip every test via flags, print PASS lines", `package main
import ("flag"; "fmt"; "testing")
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return 0 }
func init() {
	fmt.Println("--- PASS: TestAdd (0.00s)\n--- PASS: TestSub (0.00s)")
	// Register the testing flags early and filter every real test out, so
	// m.Run() runs nothing and reports success.
	testing.Init()
	_ = flag.Set("test.run", "^$")
}
func main() {}`, false},

	{"write a guessed verdict file then os.Exit(0)", `package main
import ("os"; "path/filepath")
func Add(a, b int) int { return 0 }
func Sub(a, b int) int { return 0 }
func init() {
	body := "code=0\npass TestAdd\npass TestSub\nmac=00\n"
	for _, dir := range []string{".", "/work", "/tmp"} {
		for _, name := range []string{".verdict", "verdict", ".grade", "grade.txt"} {
			_ = os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)
		}
	}
	os.Exit(0)
}
func main() {}`, false},

	{"overwrite whatever verdict files exist, then exit 0", `package main
import ("os"; "path/filepath")
func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return 0 }
func init() {
	matches, _ := filepath.Glob("/work/.v*")
	for _, m := range matches { _ = os.WriteFile(m, []byte("code=0\npass TestAdd\npass TestSub\nmac=00\n"), 0o600) }
	matches, _ = filepath.Glob("./.v*")
	for _, m := range matches { _ = os.WriteFile(m, []byte("code=0\npass TestAdd\npass TestSub\nmac=00\n"), 0o600) }
}
func main() {}`, false},
}

// fooledOldGrader records which cases the previous stdout-based grader accepted
// (the genuine solution, and the forgeries that printed recognisable pass lines).
var fooledOldGrader = map[string]bool{
	"honest solution":                               true,
	"forged text PASS lines then os.Exit(0)":        true,
	"forged test2json marker lines then os.Exit(0)": true,
	"skip every test via flags, print PASS lines":   true,
}

func newShimDockerForTest(t *testing.T) *Docker {
	t.Helper()
	if testing.Short() {
		t.Skip("short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "entry")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/valentrahq/golearn/cmd/sandbox-entrypoint").CombinedOutput(); err != nil {
		t.Fatalf("build entrypoint: %v\n%s", err, out)
	}
	work := filepath.Join(dir, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	// The shim stands in for the docker CLI and runs the entrypoint directly,
	// with a fresh work directory for each run.
	shim := filepath.Join(dir, "docker")
	script := `#!/bin/sh
case "$1" in
  image) exit 0;;
  rm) exit 0;;
  run) rm -rf "` + work + `"/* "` + work + `"/.v* ; GOLEARN_WORK_DIR="` + work + `" GOLEARN_GOCACHE="` + filepath.Join(dir, "cache") + `" GOPROXY=off exec "` + bin + `";;
esac
`
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return NewDocker(DockerConfig{Bin: shim, Image: "img", MemoryMB: 64, CPUs: "1", PIDs: 8, Concurrency: 1})
}

// TestChallengeGradingResistsForgedResults runs real `go test` through the
// sandbox entrypoint and checks that only a genuine pass yields an
// authenticated all-pass verdict.
func TestChallengeGradingResistsForgedResults(t *testing.T) {
	d := newShimDockerForTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, c := range gradingCases {
		t.Run(c.name, func(t *testing.T) {
			res, err := d.Run(ctx, Request{Kind: KindTest, TimeoutSec: 20, Files: map[string]string{"main.go": c.code, "main_test.go": gradingTests}})
			if err != nil {
				t.Fatal(err)
			}
			if res.BuildFailed {
				t.Fatalf("test program must compile, otherwise this case proves nothing:\n%s", res.Stdout+res.Stderr)
			}
			// What the previous grader (stdout events + exit code) would have decided.
			byName := map[string]bool{}
			for _, tr := range ParseTestJSON(res.Stdout) {
				byName[tr.Name] = tr.Passed
			}
			oldWouldPass := res.ExitCode == 0 && byName["TestAdd"] && byName["TestSub"]
			if oldWouldPass != fooledOldGrader[c.name] {
				t.Fatalf("old-grader decision = %v, want %v (the case no longer demonstrates the intended attack)\nstdout=%.500s", oldWouldPass, fooledOldGrader[c.name], res.Stdout)
			}
			got := res.Verdict.AllPassed(gradingExpected) && res.ExitCode == 0 && !res.TimedOut && !res.BuildFailed
			if got != c.wantPass {
				t.Fatalf("passed=%v, want %v\nexit=%d verdict=%+v\nstdout=%.600s\nstderr=%.300s", got, c.wantPass, res.ExitCode, res.Verdict, res.Stdout, res.Stderr)
			}
			if c.name == "honest wrong answer" {
				if res.Verdict == nil || !res.Verdict.Valid || res.Verdict.Tests["TestAdd"] || !res.Verdict.Tests["TestSub"] {
					t.Fatalf("an honest failing run should still carry a valid verdict with per-test results: %+v", res.Verdict)
				}
			}
		})
	}
}

func TestInstrumentRejectsTestMainAndEmptyFiles(t *testing.T) {
	g, err := NewGrader(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"own TestMain":  "package main\nimport \"testing\"\nfunc TestMain(m *testing.M) { m.Run() }\nfunc TestA(t *testing.T) {}\n",
		"no tests":      "package main\nfunc helper() {}\n",
		"wrong package": "package other\nimport \"testing\"\nfunc TestA(t *testing.T) {}\n",
		"syntax error":  "package main\nfunc (",
	}
	for name, src := range cases {
		if _, _, err := g.Instrument(src); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestVerdictMACIsEnforced(t *testing.T) {
	dir := t.TempDir()
	g, _ := NewGrader(dir)
	write := func(s string) { _ = os.WriteFile(g.VerdictPath, []byte(s), 0o600) }

	if v := g.ReadVerdict(); v.Valid {
		t.Fatal("a missing verdict must not be valid")
	}
	write("code=0\npass TestAdd\nmac=deadbeef\n")
	if v := g.ReadVerdict(); v.Valid {
		t.Fatal("a wrong MAC must not be valid")
	}
	write("garbage")
	if v := g.ReadVerdict(); v.Valid {
		t.Fatal("garbage must not be valid")
	}
	if (&Verdict{Valid: true, Tests: map[string]bool{"TestAdd": true}}).AllPassed(nil) {
		t.Fatal("an empty expectation list must never pass")
	}
	var nilV *Verdict
	if nilV.Passed("x") || nilV.AllPassed([]string{"x"}) {
		t.Fatal("a nil verdict must never pass")
	}
}
