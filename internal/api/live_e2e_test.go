package api_test

// Live end-to-end test against a running GoLearn server (optionally backed by
// PostgreSQL and the real Docker sandbox). It is skipped unless
// GOLEARN_E2E_URL is set:
//
//	GOLEARN_E2E_URL=http://localhost:8080 go test ./internal/api -run TestLiveE2E -v
//
// Register two learners, walk the learning loop with the real curriculum, and
// check idempotency, persistence across sessions, and user isolation.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/valentrahq/golearn/internal/content"
)

type liveClient struct {
	t    *testing.T
	base string
	c    *http.Client
}

func newLive(t *testing.T, base string) *liveClient {
	jar, _ := cookiejar.New(nil)
	return &liveClient{t: t, base: base, c: &http.Client{Jar: jar, Timeout: 90 * time.Second}}
}

func (l *liveClient) do(method, path string, body any) (int, map[string]any) {
	l.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, l.base+path, rd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GoLearn-CSRF", "1")
	resp, err := l.c.Do(req)
	if err != nil {
		l.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func dig(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

func num(v any) int { f, _ := v.(float64); return int(f) }

func (l *liveClient) summary() map[string]any {
	l.t.Helper()
	st, r := l.do("GET", "/api/dashboard", nil)
	if st != 200 {
		l.t.Fatalf("dashboard: %d %v", st, r)
	}
	s, _ := dig(r, "data", "summary").(map[string]any)
	return s
}

func want(t *testing.T, what string, got, exp any) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(exp) {
		t.Fatalf("%s: got %v, want %v", what, got, exp)
	}
}

func TestLiveE2E(t *testing.T) {
	base := os.Getenv("GOLEARN_E2E_URL")
	if base == "" {
		t.Skip("set GOLEARN_E2E_URL to run against a live server")
	}
	cat, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	// The first published lesson that has both an exercise and a quiz.
	var lesson *content.Lesson
	for _, l := range cat.PublishedLessons() {
		if l.Exercise != nil && len(l.Quiz) > 0 {
			lesson = l
			break
		}
	}
	ch := cat.Challenge("fizzbuzz")
	if lesson == nil || ch == nil {
		t.Fatal("expected a lesson with exercise+quiz and the fizzbuzz challenge")
	}
	lp := "/api/lessons/" + lesson.ModuleID + "/" + lesson.Slug
	stamp := time.Now().UnixNano()
	a, b := newLive(t, base), newLive(t, base)
	emailA, emailB := fmt.Sprintf("a%d@e2e.test", stamp), fmt.Sprintf("b%d@e2e.test", stamp)
	pw := "correct horse battery"

	// Register, then unauthenticated access is refused.
	if st, r := a.do("POST", "/api/auth/register", map[string]any{"email": emailA, "password": pw, "displayName": "A"}); st != 201 {
		t.Fatalf("register A: %d %v", st, r)
	}
	if st, r := b.do("POST", "/api/auth/register", map[string]any{"email": emailB, "password": pw, "displayName": "B"}); st != 201 {
		t.Fatalf("register B: %d %v", st, r)
	}
	anon := newLive(t, base)
	if st, _ := anon.do("GET", "/api/dashboard", nil); st != 401 {
		t.Fatalf("anonymous dashboard should be 401, got %d", st)
	}
	if st, _ := anon.do("GET", "/api/progress", nil); st != 401 {
		t.Fatalf("anonymous progress should be 401, got %d", st)
	}

	// Roadmap, module, lesson; answers/solutions must not leak.
	if st, r := a.do("GET", "/api/modules", nil); st != 200 || len(dig(r, "data").([]any)) != 26 {
		t.Fatalf("modules: %d", st)
	}
	st, raw := a.do("GET", lp, nil)
	if st != 200 {
		t.Fatalf("lesson: %d", st)
	}
	blob, _ := json.Marshal(raw)
	if strings.Contains(string(blob), lesson.Exercise.Solution) && len(lesson.Exercise.Solution) > 40 {
		t.Fatal("lesson response leaks the exercise solution")
	}

	// Playground execution (needs a runner).
	_, cfg := a.do("GET", "/api/config", nil)
	runnerUp, _ := dig(cfg, "data", "runner", "available").(bool)
	if runnerUp {
		st, r := a.do("POST", "/api/run", map[string]any{"code": "package main\nimport \"fmt\"\nfunc main(){ fmt.Println(\"Hello, World!\") }"})
		if st != 200 || dig(r, "data", "stdout") != "Hello, World!\n" {
			t.Fatalf("playground run: %d %v", st, r)
		}
		// Exercise: starter fails, solution passes, repeated passes stay idempotent.
		st, r = a.do("POST", lp+"/exercise", map[string]any{"code": lesson.Exercise.Starter})
		if st != 200 || dig(r, "data", "passed") != false {
			t.Fatalf("starter should not pass: %d %v", st, r)
		}
		for i := 0; i < 2; i++ {
			st, r = a.do("POST", lp+"/exercise", map[string]any{"code": lesson.Exercise.Solution})
			if st != 200 || dig(r, "data", "passed") != true {
				t.Fatalf("solution should pass: %d %v", st, r)
			}
		}
	} else {
		t.Log("runner unavailable: skipping execution steps")
	}

	// Completion requires the quiz.
	if st, _ := a.do("POST", "/api/progress/lessons/"+lesson.ModuleID+"/"+lesson.Slug, map[string]any{"completed": true}); st != 409 {
		t.Fatalf("completing before the quiz should be 409, got %d", st)
	}
	wrong := map[string]any{}
	right := map[string]any{}
	for _, q := range lesson.Quiz {
		switch q.Type {
		case "short":
			right[q.ID] = q.Accept[0]
			wrong[q.ID] = "zzz-not-it"
		case "multi":
			right[q.ID] = q.Correct
			wrong[q.ID] = []int{}
		default:
			right[q.ID] = q.Correct[0]
			wrong[q.ID] = (q.Correct[0] + 1) % len(q.Options)
		}
	}
	if st, r := a.do("POST", lp+"/quiz", map[string]any{"answers": wrong}); st != 200 || dig(r, "data", "passed") != false {
		t.Fatalf("wrong quiz should fail: %d %v", st, r)
	}
	for i := 0; i < 2; i++ {
		if st, r := a.do("POST", lp+"/quiz", map[string]any{"answers": right}); st != 200 || dig(r, "data", "passed") != true {
			t.Fatalf("right quiz should pass: %d %v", st, r)
		}
	}
	// Completion is idempotent.
	for i := 0; i < 3; i++ {
		if st, r := a.do("POST", "/api/progress/lessons/"+lesson.ModuleID+"/"+lesson.Slug, map[string]any{"completed": true}); st != 200 {
			t.Fatalf("complete #%d: %d %v", i, st, r)
		}
	}
	sum := a.summary()
	want(t, "lessonsCompleted after repeats", num(sum["lessonsCompleted"]), 1)
	xp1 := num(sum["xp"])
	if xp1 <= 0 {
		t.Fatal("expected XP after completing a lesson")
	}
	if num(sum["currentStreak"]) < 1 {
		t.Fatalf("expected a streak of at least 1, got %v", sum["currentStreak"])
	}

	// Challenge: starter fails, solution passes (needs a runner), repeats do not double count.
	if runnerUp {
		if st, r := a.do("POST", "/api/challenges/fizzbuzz/submit", map[string]any{"code": ch.Starter}); st != 200 || dig(r, "data", "passed") != false {
			t.Fatalf("starter should fail fizzbuzz: %d %v", st, r)
		}
		// A learner who forges "pass" output and exits 0 must not pass.
		var forged strings.Builder
		forged.WriteString("package main\nimport (\"fmt\"; \"os\")\nfunc main() {}\nfunc init() {\n")
		for _, n := range ch.TestNames {
			forged.WriteString(fmt.Sprintf("\tfmt.Println(\"=== RUN   %s\\n--- PASS: %s (0.00s)\")\n", n, n))
		}
		forged.WriteString("\tfmt.Println(\"PASS\")\n\tos.Exit(0)\n}\n")
		if st, r := a.do("POST", "/api/challenges/fizzbuzz/submit", map[string]any{"code": forged.String()}); st != 200 || dig(r, "data", "passed") != false {
			t.Fatalf("forged pass output must not pass: %d %v", st, r)
		}
		want(t, "forgery does not count", num(a.summary()["challengesCompleted"]), 0)
		for i := 0; i < 2; i++ {
			if st, r := a.do("POST", "/api/challenges/fizzbuzz/submit", map[string]any{"code": ch.Solution}); st != 200 || dig(r, "data", "passed") != true {
				t.Fatalf("solution should pass fizzbuzz: %d %v", st, r)
			}
		}
		want(t, "challengesCompleted", num(a.summary()["challengesCompleted"]), 1)
		if st, _ := a.do("GET", "/api/challenges/fizzbuzz/solution", nil); st != 200 {
			t.Fatalf("solution should unlock after passing, got %d", st)
		}
	}

	// Achievements and skills.
	_, ach := a.do("GET", "/api/achievements", nil)
	unlocked := map[string]bool{}
	for _, x := range dig(ach, "data", "achievements").([]any) {
		m := x.(map[string]any)
		unlocked[m["id"].(string)], _ = m["unlocked"].(bool)
	}
	if !unlocked["first-lesson"] {
		t.Fatal("first-lesson achievement should be unlocked")
	}
	if st, _ := a.do("GET", "/api/skills", nil); st != 200 {
		t.Fatalf("skills: %d", st)
	}

	// Persistence: logout/login and a second concurrent session see the same state.
	before := a.summary()
	if st, _ := a.do("POST", "/api/auth/logout", nil); st != 200 && st != 204 {
		t.Fatalf("logout: %d", st)
	}
	if st, _ := a.do("GET", "/api/dashboard", nil); st != 401 {
		t.Fatalf("after logout dashboard should be 401, got %d", st)
	}
	second := newLive(t, base)
	if st, r := second.do("POST", "/api/auth/login", map[string]any{"email": emailA, "password": pw}); st != 200 {
		t.Fatalf("login: %d %v", st, r)
	}
	third := newLive(t, base)
	if st, r := third.do("POST", "/api/auth/login", map[string]any{"email": emailA, "password": pw}); st != 200 {
		t.Fatalf("second device login: %d %v", st, r)
	}
	for name, cl := range map[string]*liveClient{"second": second, "third": third} {
		s := cl.summary()
		want(t, name+" lessonsCompleted", s["lessonsCompleted"], before["lessonsCompleted"])
		want(t, name+" xp", s["xp"], before["xp"])
		want(t, name+" challengesCompleted", s["challengesCompleted"], before["challengesCompleted"])
	}
	// A wrong password is refused with a generic error.
	if st, r := anon.do("POST", "/api/auth/login", map[string]any{"email": emailA, "password": "wrong-password-123"}); st != 401 {
		t.Fatalf("bad login should be 401, got %d %v", st, r)
	}

	// Isolation: B sees none of A's progress and cannot read A's earned solution.
	sb := b.summary()
	want(t, "B lessonsCompleted", num(sb["lessonsCompleted"]), 0)
	want(t, "B challengesCompleted", num(sb["challengesCompleted"]), 0)
	want(t, "B xp", num(sb["xp"]), 0)
	if st, _ := b.do("GET", "/api/challenges/fizzbuzz/solution", nil); st == 200 {
		t.Fatal("B must not get the solution without earning it")
	}
	_, pb := b.do("GET", "/api/progress", nil)
	if lessons, _ := dig(pb, "data", "lessons").(map[string]any); lessons[lesson.ID] != nil {
		if m, _ := lessons[lesson.ID].(map[string]any); m["completed"] == true {
			t.Fatal("B sees A's completed lesson")
		}
	}
	// B changing their own state does not affect A.
	b.do("POST", lp+"/quiz", map[string]any{"answers": right})
	want(t, "A lessonsCompleted unchanged", num(second.summary()["lessonsCompleted"]), 1)
}
