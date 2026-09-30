package content

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoadCatalog(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Modules) == 0 {
		t.Fatal("no modules")
	}
	t.Logf("modules=%d lessons=%d challenges=%d projects=%d glossary=%d",
		len(c.Modules), len(c.PublishedLessons()), len(c.Challenges), len(c.Projects), len(c.Glossary))
}

func TestSeedMinimums(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	quizzes := 0
	for _, l := range c.PublishedLessons() {
		if len(l.Quiz) > 0 {
			quizzes++
		}
	}
	for name, ok := range map[string]bool{
		"20+ modules":    len(c.Modules) >= 20,
		"100+ lessons":   len(c.PublishedLessons()) >= 100,
		"50+ challenges": len(c.Challenges) >= 50,
		"10+ projects":   len(c.Projects) >= 10,
		"50+ glossary":   len(c.Glossary) >= 50,
		"20+ quizzes":    quizzes >= 20,
	} {
		if !ok {
			t.Errorf("seed minimum not met: %s", name)
		}
	}
}

func TestParseDocumentIgnoresHeadingsInFences(t *testing.T) {
	doc, err := parseDocument("# T\nid: x\n\n## A\nslug: a\n\n### Concept\n```text\n## not a heading\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.items) != 1 {
		t.Fatalf("want 1 item, got %d", len(doc.items))
	}
}

func TestSplitSolution(t *testing.T) {
	sol, st, marked := splitSolution("func f() {\n\t// BEGIN\n\tx := 1\n\t// END\n}")
	if !marked || strings.Contains(sol, "BEGIN") || !strings.Contains(sol, "x := 1") {
		t.Fatalf("bad solution: %q", sol)
	}
	if strings.Contains(st, "x := 1") || !strings.Contains(st, "TODO") {
		t.Fatalf("bad starter: %q", st)
	}
}

func TestQuestionParsing(t *testing.T) {
	qs, err := parseQuestions("Q: Pick\nT: multi\n- [x] a\n- [ ] b\n- [x] c\nE: because\n\nQ: name?\nT: short\nA: foo | bar\nE: yes\n\nQ: tf\nT: tf\nA: false\nE: no", "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 3 || len(qs[0].Correct) != 2 || len(qs[1].Accept) != 2 || qs[2].Correct[0] != 1 {
		t.Fatalf("unexpected: %+v", qs)
	}
}

// TestContentRuns executes every runnable snippet with the local Go toolchain.
// This is a developer/CI check of the curriculum itself; the running
// application never executes learner code on its own host.
func TestContentRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping content execution in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not found")
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	type job struct {
		name string
		fn   func() error
	}
	var jobs []job
	for _, l := range c.PublishedLessons() {
		l := l
		for i, ex := range l.Examples {
			if ex.Runnable {
				ex := ex
				jobs = append(jobs, job{fmt.Sprintf("example/%s/%d", l.ID, i), func() error {
					_, err := goRun(map[string]string{"main.go": ex.Code}, "run", ".")
					return err
				}})
			}
		}
		if l.Exercise != nil {
			ex := l.Exercise
			jobs = append(jobs, job{"exercise/" + l.ID, func() error {
				out, err := goRun(map[string]string{"main.go": ex.Solution}, "run", ".")
				if err != nil {
					return err
				}
				if out != ex.Expected {
					return fmt.Errorf("solution output %q, expected %q", out, ex.Expected)
				}
				out, _ = goRun(map[string]string{"main.go": ex.Starter}, "run", ".")
				if out == ex.Expected {
					return fmt.Errorf("starter already produces the expected output")
				}
				return nil
			}})
		}
	}
	for _, ch := range c.Challenges {
		ch := ch
		jobs = append(jobs, job{"challenge/" + ch.ID, func() error {
			if _, err := goRun(map[string]string{"main.go": ch.Solution, "main_test.go": ch.Tests}, "test", "-race", "-count=1", "."); err != nil {
				return fmt.Errorf("solution fails tests: %w", err)
			}
			// Challenge starters must compile so learners see failing tests, not build errors.
			if _, err := goRun(map[string]string{"main.go": ch.Starter, "main_test.go": ch.Tests}, "vet", "."); err != nil {
				return fmt.Errorf("starter does not compile with tests: %w", err)
			}
			if _, err := goRun(map[string]string{"main.go": ch.Starter, "main_test.go": ch.Tests}, "test", "-count=1", "."); err == nil {
				return fmt.Errorf("starter unexpectedly passes all tests")
			}
			return nil
		}})
	}
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := j.fn(); err != nil {
				t.Errorf("%s: %v", j.name, err)
			}
		}()
	}
	wg.Wait()
	t.Logf("checked %d snippets", len(jobs))
}

func goRun(files map[string]string, args ...string) (string, error) {
	dir, err := os.MkdirTemp("", "golearn-content-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	files["go.mod"] = "module sandbox\n\ngo 1.24\n"
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			return "", err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off", "CGO_ENABLED=0")
	if args[0] == "test" {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%v\n%s%s", err, stdout.String(), stderr.String())
	}
	return stdout.String(), nil
}
