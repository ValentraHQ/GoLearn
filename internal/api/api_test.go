package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/valentrahq/golearn/internal/api"
	"github.com/valentrahq/golearn/internal/config"
	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/runner"
	"github.com/valentrahq/golearn/internal/store"
)

const fixtureModule = "# Basics\nid: basics\nnumber: 01\nskill: fundamentals\nsummary: Start here.\n\n" +
	"## Hello\nslug: hello\nchallenges: add\n\n### Objectives\n" + // ignored section, objectives come from meta below
	"## Real\nslug: real\nobjectives: Say hi; Compile\ntakeaways: It runs\nminutes: 4\nchallenges: add\n\n" +
	"### Concept\nGo programs start in `main`.\n\n" +
	"### Example\n```go\npackage main\n\nfunc main() {}\n```\n\n" +
	"### Exercise\nPrint hi.\n```text expect\nhi\n```\n```go solution\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\t// BEGIN\n\tfmt.Println(\"hi\")\n\t// END\n}\n```\n\n" +
	"### Check\nQ: Entry point?\nT: mcq\n- [ ] init\n- [x] main\nE: main is the entry point.\n\nQ: Pick both\nT: multi\n- [x] a\n- [ ] b\n- [x] c\nE: a and c.\n\nQ: Name the file\nT: short\nA: go.mod | go.mod file\nE: go.mod.\n\n" +
	"## Later\nslug: later\nstatus: planned\n"

func fixtureFS() fstest.MapFS {
	// The first lesson block above is intentionally malformed; strip it.
	mod := strings.Replace(fixtureModule, "## Hello\nslug: hello\nchallenges: add\n\n### Objectives\n", "", 1)
	return fstest.MapFS{
		"data/modules/01-basics.md": {Data: []byte(mod)},
		"data/challenges/c.md":      {Data: []byte("# C\n\n## add\ntitle: Add\ndifficulty: beginner\nmodule: basics\nlesson: basics/real\nskill: functions\n\n### Problem\nAdd two ints.\n\n### Starter\n```go\npackage main\n\nfunc Add(a, b int) int { return 0 }\n\nfunc main() {}\n```\n\n### Tests\n```go\npackage main\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Fatal(\"bad\")\n\t}\n}\n```\n\n### Solution\n```go\npackage main\n\nfunc Add(a, b int) int { return a + b }\n\nfunc main() {}\n```\n")},
		"data/projects.md":          {Data: []byte("# P\n\n## Tiny\nlevel: beginner\nmodule: basics\nskills: fundamentals\nhours: 1\n\n### Summary\nBuild it.\n\n### Tasks\n- One\n- Two\n")},
		"data/glossary.md":          {Data: []byte("# G\n\n## Goroutine\nrelated: basics/real\n\n### Simple\nA light thread.\n\n### Technical\nA runtime-managed coroutine.\n")},
	}
}

type fakeRunner struct {
	available bool
	res       runner.Result
	last      runner.Request
}

func (f *fakeRunner) Status(context.Context) runner.Status {
	return runner.Status{Available: f.available, Backend: "fake", Reason: "fake unavailable"}
}
func (f *fakeRunner) Run(_ context.Context, r runner.Request) (runner.Result, error) {
	if !f.available {
		return runner.Result{}, runner.ErrUnavailable
	}
	f.last = r
	return f.res, nil
}

type env struct {
	t      *testing.T
	srv    *httptest.Server
	runner *fakeRunner
	client *http.Client
}

func newEnv(t *testing.T) *env {
	t.Helper()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	cat, err := content.LoadFS(fixtureFS(), "data")
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(context.Background(), "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	fr := &fakeRunner{available: true}
	cfg := config.Config{CookieName: "s", SessionTTL: 3600e9, BcryptCost: 4, RunsPerMinute: 600}
	srv := httptest.NewServer(api.New(cfg, st, cat, fr).Handler())
	t.Cleanup(srv.Close)
	jar := newJar()
	return &env{t: t, srv: srv, runner: fr, client: &http.Client{Jar: jar}}
}

func (e *env) do(method, path string, body any, csrf bool) (int, map[string]any) {
	e.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, e.srv.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrf {
		req.Header.Set("X-GoLearn-CSRF", "1")
	}
	resp, err := e.client.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (e *env) call(method, path string, body any) (int, map[string]any) {
	return e.do(method, path, body, true)
}

func (e *env) register(email string) {
	e.t.Helper()
	code, out := e.call("POST", "/api/auth/register", map[string]any{"email": email, "password": "correct horse battery", "displayName": "Ada"})
	if code != 201 {
		e.t.Fatalf("register: %d %v", code, out)
	}
}

func data(m map[string]any) map[string]any { d, _ := m["data"].(map[string]any); return d }

func TestAuthFlow(t *testing.T) {
	e := newEnv(t)
	if code, _ := e.do("POST", "/api/auth/register", map[string]any{"email": "a@b.co", "password": "correct horse battery"}, false); code != 403 {
		t.Fatalf("missing CSRF header should be 403, got %d", code)
	}
	if code, _ := e.call("POST", "/api/auth/register", map[string]any{"email": "nope", "password": "correct horse battery"}); code != 400 {
		t.Fatalf("bad email: %d", code)
	}
	if code, _ := e.call("POST", "/api/auth/register", map[string]any{"email": "a@b.co", "password": "short"}); code != 400 {
		t.Fatalf("short password: %d", code)
	}
	e.register("Ada@Example.com")
	if code, _ := e.call("POST", "/api/auth/register", map[string]any{"email": "ada@example.com", "password": "correct horse battery"}); code != 409 {
		t.Fatalf("duplicate: %d", code)
	}
	code, out := e.call("GET", "/api/auth/me", nil)
	if code != 200 || data(out)["email"] != "ada@example.com" {
		t.Fatalf("me: %d %v", code, out)
	}
	e.call("POST", "/api/auth/logout", nil)
	if _, out := e.call("GET", "/api/auth/me", nil); out["data"] != nil {
		t.Fatalf("expected signed out: %v", out)
	}
	if code, _ := e.call("POST", "/api/auth/login", map[string]any{"email": "ada@example.com", "password": "wrong password!!"}); code != 401 {
		t.Fatalf("bad login: %d", code)
	}
	if code, _ := e.call("POST", "/api/auth/login", map[string]any{"email": "ada@example.com", "password": "correct horse battery"}); code != 200 {
		t.Fatalf("login: %d", code)
	}
}

func TestProtectedRoutesNeedAuth(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"/api/dashboard", "/api/progress", "/api/skills", "/api/achievements", "/api/profile"} {
		if code, _ := e.call("GET", p, nil); code != 401 {
			t.Errorf("%s: want 401, got %d", p, code)
		}
	}
	if code, _ := e.call("POST", "/api/run", map[string]any{"code": "package main"}); code != 401 {
		t.Errorf("run without auth: %d", code)
	}
}

func TestLessonDoesNotLeakAnswers(t *testing.T) {
	e := newEnv(t)
	req, _ := http.NewRequest("GET", e.srv.URL+"/api/lessons/basics/real", nil)
	resp, err := e.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	body := buf.String()
	for _, secret := range []string{"fmt.Println(\\\"hi\\\")", "main is the entry point", "go.mod file"} {
		if strings.Contains(body, secret) {
			t.Errorf("lesson response leaks %q", secret)
		}
	}
	if !strings.Contains(body, "TODO: write your code here") {
		t.Errorf("starter code missing: %s", body)
	}
	if code, _ := e.call("GET", "/api/lessons/basics/later", nil); code != 404 {
		t.Errorf("planned lesson must be 404, got %d", code)
	}
}

func TestLessonLoop(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")

	// Cannot complete before passing the quiz.
	if code, out := e.call("POST", "/api/progress/lessons/basics/real", map[string]any{"completed": true}); code != 409 {
		t.Fatalf("expected 409, got %d %v", code, out)
	}

	// Wrong exercise output does not pass.
	e.runner.res = runner.Result{Stdout: "nope\n"}
	_, out := e.call("POST", "/api/lessons/basics/real/exercise", map[string]any{"code": "package main\nfunc main(){}"})
	if data(out)["passed"] != false {
		t.Fatalf("should not pass: %v", out)
	}
	// Right output (trailing whitespace ignored) passes.
	e.runner.res = runner.Result{Stdout: "hi  \n\n"}
	_, out = e.call("POST", "/api/lessons/basics/real/exercise", map[string]any{"code": "package main\nfunc main(){}"})
	if data(out)["passed"] != true {
		t.Fatalf("should pass: %v", out)
	}
	if e.runner.last.Kind != runner.KindRun || e.runner.last.Files["main.go"] == "" {
		t.Fatalf("runner request wrong: %+v", e.runner.last)
	}

	// Quiz: wrong then right.
	_, out = e.call("POST", "/api/lessons/basics/real/quiz", map[string]any{"answers": map[string]any{"q1": 0, "q2": []int{0}, "q3": "nope"}})
	if d := data(out); d["passed"] != false || d["score"].(float64) != 0 {
		t.Fatalf("quiz should fail: %v", out)
	}
	_, out = e.call("POST", "/api/lessons/basics/real/quiz", map[string]any{"answers": map[string]any{"q1": 1, "q2": []int{2, 0}, "q3": " GO.MOD "}})
	if d := data(out); d["passed"] != true || d["score"].(float64) != 3 {
		t.Fatalf("quiz should pass: %v", out)
	}

	// Now completion works and is reflected everywhere.
	if code, out := e.call("POST", "/api/progress/lessons/basics/real", map[string]any{"completed": true}); code != 200 {
		t.Fatalf("complete: %d %v", code, out)
	}
	_, out = e.call("GET", "/api/dashboard", nil)
	sum := data(out)["summary"].(map[string]any)
	if sum["lessonsCompleted"].(float64) != 1 || sum["overallPercent"].(float64) != 100 || sum["xp"].(float64) <= 0 {
		t.Fatalf("dashboard: %v", sum)
	}
	if sum["currentStreak"].(float64) != 1 {
		t.Fatalf("streak: %v", sum["currentStreak"])
	}
	achs := sum["achievements"].([]any)
	got := map[string]bool{}
	for _, a := range achs {
		am := a.(map[string]any)
		got[am["id"].(string)] = am["unlocked"].(bool)
	}
	if !got["first-program"] || !got["first-lesson"] {
		t.Fatalf("achievements not unlocked: %v", got)
	}
	// Uncomplete.
	e.call("POST", "/api/progress/lessons/basics/real", map[string]any{"completed": false})
	_, out = e.call("GET", "/api/progress", nil)
	if data(out)["summary"].(map[string]any)["lessonsCompleted"].(float64) != 0 {
		t.Fatal("expected un-completed")
	}
}

func TestRunnerUnavailableIsExplicit(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	e.runner.available = false
	code, out := e.call("POST", "/api/run", map[string]any{"code": "package main\nfunc main(){}"})
	if code != 503 || out["error"].(map[string]any)["code"] != "runner_unavailable" {
		t.Fatalf("want 503 runner_unavailable, got %d %v", code, out)
	}
	// With no runner the exercise is not enforced, but the quiz still is.
	e.call("POST", "/api/lessons/basics/real/quiz", map[string]any{"answers": map[string]any{"q1": 1, "q2": []int{0, 2}, "q3": "go.mod"}})
	if code, out := e.call("POST", "/api/progress/lessons/basics/real", map[string]any{"completed": true}); code != 200 {
		t.Fatalf("complete without runner: %d %v", code, out)
	}
}

func TestCodeSizeLimit(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	code, _ := e.call("POST", "/api/run", map[string]any{"code": strings.Repeat("x", 40<<10)})
	if code != 413 {
		t.Fatalf("want 413, got %d", code)
	}
	if code, _ := e.call("POST", "/api/run", map[string]any{"code": "  "}); code != 400 {
		t.Fatalf("want 400 for blank, got %d", code)
	}
}

func TestChallengeSubmitAndSolutionGate(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	if code, _ := e.call("GET", "/api/challenges/add/solution", nil); code != 403 {
		t.Fatalf("solution should be locked: %d", code)
	}
	failing := runner.Result{ExitCode: 1, Stdout: `{"Action":"run","Test":"TestAdd"}
{"Action":"output","Test":"TestAdd","Output":"    main_test.go:9: bad\n"}
{"Action":"fail","Test":"TestAdd"}
`}
	e.runner.res = failing
	for i := 0; i < 3; i++ {
		_, out := e.call("POST", "/api/challenges/add/submit", map[string]any{"code": "package main\nfunc main(){}"})
		if d := data(out); d["passed"] != false || d["testsTotal"].(float64) != 1 {
			t.Fatalf("submit: %v", out)
		}
	}
	if e.runner.last.Kind != runner.KindTest || e.runner.last.Files["main_test.go"] == "" {
		t.Fatalf("expected test run with hidden test file: %+v", e.runner.last)
	}
	if code, out := e.call("GET", "/api/challenges/add/solution", nil); code != 200 || !strings.Contains(data(out)["solution"].(string), "a + b") {
		t.Fatalf("solution after 3 attempts: %d %v", code, out)
	}
	// A run where os.Exit(0) hides the tests must not count as a pass.
	e.runner.res = runner.Result{ExitCode: 0, Stdout: ""}
	_, out := e.call("POST", "/api/challenges/add/submit", map[string]any{"code": "package main\nfunc init(){}\nfunc main(){}"})
	if data(out)["passed"] != false {
		t.Fatalf("no test events must not pass: %v", out)
	}
	e.runner.res = runner.Result{ExitCode: 0, Stdout: `{"Action":"run","Test":"TestAdd"}
{"Action":"pass","Test":"TestAdd"}
`}
	_, out = e.call("POST", "/api/challenges/add/submit", map[string]any{"code": "package main\nfunc main(){}"})
	if data(out)["passed"] != true {
		t.Fatalf("should pass: %v", out)
	}
	_, out = e.call("GET", "/api/challenges?status=passed", nil)
	if out["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("filter passed: %v", out)
	}
	// Hidden tests and solution never appear in the challenge payload.
	req, _ := http.NewRequest("GET", e.srv.URL+"/api/challenges/add", nil)
	resp, _ := e.client.Do(req)
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	if strings.Contains(buf.String(), "a + b") || strings.Contains(buf.String(), "t.Fatal") {
		t.Fatalf("challenge payload leaks solution/tests")
	}
}

func TestProjectsSearchGlossary(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	if code, _ := e.call("PUT", "/api/projects/tiny/tasks/t1", map[string]any{"done": true}); code != 200 {
		t.Fatal("task t1")
	}
	if code, _ := e.call("PUT", "/api/projects/tiny/tasks/t9", map[string]any{"done": true}); code != 404 {
		t.Fatal("unknown task should 404")
	}
	_, out := e.call("PUT", "/api/projects/tiny/tasks/t2", map[string]any{"done": true})
	if data(out)["completed"] != true {
		t.Fatalf("project should complete: %v", out)
	}
	_, out = e.call("GET", "/api/search?q=goroutine", nil)
	res := data(out)["results"].([]any)
	if len(res) == 0 || res[0].(map[string]any)["kind"] != "glossary" {
		t.Fatalf("search: %v", out)
	}
	_, out = e.call("GET", "/api/search?q=", nil)
	if len(data(out)["results"].([]any)) != 0 {
		t.Fatal("empty query must return no results")
	}
	_, out = e.call("GET", "/api/glossary?q=thread", nil)
	if len(out["data"].([]any)) != 1 {
		t.Fatalf("glossary: %v", out)
	}
	_, out = e.call("GET", "/api/skills", nil)
	if len(out["data"].([]any)) == 0 {
		t.Fatal("skills empty")
	}
}

func TestProfileValidation(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	good := map[string]any{"displayName": "Ada", "bio": "", "learningPath": "pro", "theme": "dark", "dailyLessons": 2, "dailyChallenges": 1, "dailyQuizzes": 0, "gamification": false, "timezone": "Europe/London"}
	if code, _ := e.call("PUT", "/api/profile", good); code != 200 {
		t.Fatal("valid profile rejected")
	}
	bad := map[string]any{}
	for k, v := range good {
		bad[k] = v
	}
	bad["timezone"] = "Mars/Olympus"
	if code, _ := e.call("PUT", "/api/profile", bad); code != 400 {
		t.Fatal("bad timezone accepted")
	}
	bad["timezone"] = "UTC"
	bad["theme"] = "neon"
	if code, _ := e.call("PUT", "/api/profile", bad); code != 400 {
		t.Fatal("bad theme accepted")
	}
	// Unknown fields are rejected.
	bad["theme"] = "dark"
	bad["role"] = "admin"
	if code, _ := e.call("PUT", "/api/profile", bad); code != 400 {
		t.Fatal("unknown field accepted")
	}
}

func TestPasswordChangeAndDelete(t *testing.T) {
	e := newEnv(t)
	e.register("ada@example.com")
	if code, _ := e.call("POST", "/api/auth/password", map[string]any{"current": "wrong wrong wrong", "new": "another long password"}); code != 403 {
		t.Fatalf("wrong current password: %d", code)
	}
	if code, _ := e.call("POST", "/api/auth/password", map[string]any{"current": "correct horse battery", "new": "another long password"}); code != 200 {
		t.Fatal("change password")
	}
	if code, _ := e.call("GET", "/api/auth/me", nil); code != 200 {
		t.Fatal("still signed in after password change")
	}
	if code, _ := e.call("DELETE", "/api/account", map[string]any{"password": "nope nope nope"}); code != 403 {
		t.Fatal("delete with wrong password")
	}
	if code, _ := e.call("DELETE", "/api/account", map[string]any{"password": "another long password"}); code != 200 {
		t.Fatal("delete account")
	}
	if code, _ := e.call("POST", "/api/auth/login", map[string]any{"email": "ada@example.com", "password": "another long password"}); code != 401 {
		t.Fatal("deleted account can still log in")
	}
}

func TestSecurityHeadersAndUnknownAPI(t *testing.T) {
	e := newEnv(t)
	resp, err := http.Get(e.srv.URL + "/api/nope")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 || resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("Content-Security-Policy") == "" {
		t.Fatalf("status=%d headers=%v", resp.StatusCode, resp.Header)
	}
}
