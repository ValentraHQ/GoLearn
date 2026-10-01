package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/httpx"
	"github.com/valentrahq/golearn/internal/progress"
	"github.com/valentrahq/golearn/internal/runner"
	"github.com/valentrahq/golearn/internal/store"
)

const (
	maxCodeBytes          = 32 << 10
	solutionAfterAttempts = 3
)

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request, u store.User) error {
	sum, p, _, err := s.loadProgress(r.Context(), u)
	if err != nil {
		return err
	}
	httpx.OK(w, map[string]any{"summary": sum, "profile": p})
	return nil
}

func (s *Server) getProgress(w http.ResponseWriter, r *http.Request, u store.User) error {
	sum, _, in, err := s.loadProgress(r.Context(), u)
	if err != nil {
		return err
	}
	lessons := map[string]map[string]bool{}
	for id, lp := range in.Lessons {
		lessons[id] = map[string]bool{"completed": lp.CompletedAt != nil, "exercisePassed": lp.ExercisePassed, "quizPassed": lp.QuizPassed}
	}
	httpx.OK(w, map[string]any{"summary": sum, "lessons": lessons})
	return nil
}

type unlocked struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// newAchievements evaluates achievements after a mutation and returns any just earned.
func (s *Server) newAchievements(ctx context.Context, u store.User) []unlocked {
	before, _ := s.store.UnlockedAchievements(ctx, u.ID)
	sum, _, _, err := s.loadProgress(ctx, u)
	if err != nil {
		return nil
	}
	out := []unlocked{}
	for _, a := range sum.Achievements {
		if a.Unlocked {
			if _, had := before[a.ID]; !had {
				out = append(out, unlocked{a.ID, a.Title})
			}
		}
	}
	return out
}

func (s *Server) runnerFor(r *http.Request, u store.User) error {
	if ok, wait := s.runRate.Allow(u.ID); !ok {
		return httpx.Err(http.StatusTooManyRequests, "rate_limited", "You're running code too quickly. Try again in "+wait.Round(time.Second).String()+".")
	}
	return nil
}

func (s *Server) execute(r *http.Request, req runner.Request) (runner.Result, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	res, err := s.runner.Run(ctx, req)
	switch {
	case errors.Is(err, runner.ErrUnavailable):
		s.metrics.Inc("golearn_runs_total", `result="unavailable"`)
		return res, httpx.Err(http.StatusServiceUnavailable, "runner_unavailable", strings.TrimPrefix(err.Error(), runner.ErrUnavailable.Error()+": "))
	case errors.Is(err, runner.ErrBusy):
		s.metrics.Inc("golearn_runs_total", `result="busy"`)
		return res, httpx.Err(http.StatusServiceUnavailable, "runner_busy", err.Error())
	case err != nil:
		s.metrics.Inc("golearn_runs_total", `result="error"`)
		return res, err
	}
	switch {
	case res.TimedOut:
		s.metrics.Inc("golearn_runs_total", `result="timeout"`)
	case res.BuildFailed:
		s.metrics.Inc("golearn_runs_total", `result="build_error"`)
	default:
		s.metrics.Inc("golearn_runs_total", `result="ok"`)
	}
	return res, nil
}

type codeRequest struct {
	Code string `json:"code"`
}

func readCode(w http.ResponseWriter, r *http.Request) (string, error) {
	var in codeRequest
	if err := httpx.Decode(w, r, &in, maxCodeBytes+1024); err != nil {
		return "", err
	}
	if strings.TrimSpace(in.Code) == "" {
		return "", httpx.BadRequest("Write some code first.")
	}
	if len(in.Code) > maxCodeBytes {
		return "", httpx.Err(http.StatusRequestEntityTooLarge, "too_large", "Code is too large (32 KB max).")
	}
	return in.Code, nil
}

// normalize makes output comparison insensitive to trailing whitespace.
func normalize(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

type runResponse struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exitCode"`
	TimedOut   bool   `json:"timedOut"`
	Truncated  bool   `json:"truncated"`
	BuildError bool   `json:"buildError"`
	DurationMS int64  `json:"durationMs"`
}

func toRunResponse(res runner.Result) runResponse {
	return runResponse{Stdout: res.Stdout, Stderr: res.Stderr, ExitCode: res.ExitCode, TimedOut: res.TimedOut,
		Truncated: res.Truncated, BuildError: res.BuildFailed, DurationMS: res.DurationMS}
}

func (s *Server) runnerStatus(w http.ResponseWriter, r *http.Request) error {
	httpx.OK(w, s.runner.Status(r.Context()))
	return nil
}

func (s *Server) playgroundRun(w http.ResponseWriter, r *http.Request, u store.User) error {
	code, err := readCode(w, r)
	if err != nil {
		return err
	}
	if err := s.runnerFor(r, u); err != nil {
		return err
	}
	res, err := s.execute(r, runner.Request{Kind: runner.KindRun, TimeoutSec: 10, Files: map[string]string{"main.go": code}})
	if err != nil {
		return err
	}
	httpx.OK(w, toRunResponse(res))
	return nil
}

func (s *Server) runExercise(w http.ResponseWriter, r *http.Request, u store.User) error {
	l, err := s.lessonOr404(r)
	if err != nil {
		return err
	}
	if l.Exercise == nil {
		return httpx.ErrNotFound
	}
	code, err := readCode(w, r)
	if err != nil {
		return err
	}
	if err := s.runnerFor(r, u); err != nil {
		return err
	}
	res, err := s.execute(r, runner.Request{Kind: runner.KindRun, TimeoutSec: 10, Files: map[string]string{"main.go": code}})
	if err != nil {
		return err
	}
	passed := res.ExitCode == 0 && !res.TimedOut && !res.BuildFailed && normalize(res.Stdout) == normalize(l.Exercise.Expected)
	var awarded []unlocked
	if passed {
		if err := s.store.MarkExercisePassed(r.Context(), u.ID, l.ID); err != nil {
			return err
		}
		awarded = s.newAchievements(r.Context(), u)
	}
	httpx.OK(w, map[string]any{"run": toRunResponse(res), "passed": passed, "expected": l.Exercise.Expected, "achievements": awarded})
	return nil
}

func (s *Server) submitQuiz(w http.ResponseWriter, r *http.Request, u store.User) error {
	l, err := s.lessonOr404(r)
	if err != nil {
		return err
	}
	if len(l.Quiz) == 0 {
		return httpx.ErrNotFound
	}
	var in struct {
		Answers map[string]json.RawMessage `json:"answers"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	type result struct {
		ID          string `json:"id"`
		Correct     bool   `json:"correct"`
		Explanation string `json:"explanation"`
	}
	results := make([]result, 0, len(l.Quiz))
	score := 0
	for _, q := range l.Quiz {
		ok := gradeAnswer(q, in.Answers[q.ID])
		if ok {
			score++
		}
		results = append(results, result{ID: q.ID, Correct: ok, Explanation: q.Explanation})
	}
	total := len(l.Quiz)
	passed := score*3 >= total*2 // at least two thirds
	if err := s.store.AddQuizAttempt(r.Context(), u.ID, l.ID, score, total, passed); err != nil {
		return err
	}
	var awarded []unlocked
	if passed {
		if err := s.store.MarkQuizPassed(r.Context(), u.ID, l.ID); err != nil {
			return err
		}
	}
	awarded = s.newAchievements(r.Context(), u)
	httpx.OK(w, map[string]any{"score": score, "total": total, "passed": passed, "results": results, "achievements": awarded})
	return nil
}

func gradeAnswer(q content.Question, raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	switch q.Type {
	case "short":
		var s string
		if json.Unmarshal(raw, &s) != nil {
			return false
		}
		s = strings.ToLower(strings.TrimSpace(s))
		for _, a := range q.Accept {
			if s == strings.ToLower(a) {
				return true
			}
		}
		return false
	case "multi":
		var got []int
		if json.Unmarshal(raw, &got) != nil {
			return false
		}
		return sameSet(got, q.Correct, len(q.Options))
	default:
		var got int
		if json.Unmarshal(raw, &got) != nil {
			return false
		}
		return len(q.Correct) == 1 && got == q.Correct[0]
	}
}

func sameSet(a, b []int, n int) bool {
	seen := map[int]bool{}
	for _, x := range a {
		if x < 0 || x >= n {
			return false
		}
		seen[x] = true
	}
	if len(seen) != len(b) {
		return false
	}
	for _, x := range b {
		if !seen[x] {
			return false
		}
	}
	return true
}

func (s *Server) setLessonProgress(w http.ResponseWriter, r *http.Request, u store.User) error {
	l, err := s.lessonOr404(r)
	if err != nil {
		return err
	}
	var in struct {
		Completed bool `json:"completed"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	lp, err := s.store.LessonProgress(r.Context(), u.ID)
	if err != nil {
		return err
	}
	cur := lp[l.ID]
	if in.Completed {
		if len(l.Quiz) > 0 && !cur.QuizPassed {
			return httpx.Err(http.StatusConflict, "requirements_not_met", "Pass the knowledge check to complete this lesson.")
		}
		if l.Exercise != nil && !cur.ExercisePassed && s.runner.Status(r.Context()).Available {
			return httpx.Err(http.StatusConflict, "requirements_not_met", "Complete the exercise to finish this lesson.")
		}
	}
	if err := s.store.SetLessonCompleted(r.Context(), u.ID, l.ID, in.Completed); err != nil {
		return err
	}
	var awarded []unlocked
	if in.Completed {
		awarded = s.newAchievements(r.Context(), u)
	}
	httpx.OK(w, map[string]any{"lessonId": l.ID, "completed": in.Completed, "next": navOf(s.cat.NextLesson(l.ID)), "achievements": awarded})
	return nil
}

func (s *Server) submitChallenge(w http.ResponseWriter, r *http.Request, u store.User) error {
	ch := s.cat.Challenge(r.PathValue("id"))
	if ch == nil {
		return httpx.ErrNotFound
	}
	code, err := readCode(w, r)
	if err != nil {
		return err
	}
	if err := s.runnerFor(r, u); err != nil {
		return err
	}
	res, err := s.execute(r, runner.Request{Kind: runner.KindTest, TimeoutSec: 10,
		Files: map[string]string{"main.go": code, "main_test.go": ch.Tests}})
	if err != nil {
		return err
	}
	byName := map[string]runner.TestResult{}
	for _, t := range runner.ParseTestJSON(res.Stdout) {
		byName[t.Name] = t
	}
	type testOut struct {
		Name    string `json:"name"`
		Passed  bool   `json:"passed"`
		Message string `json:"message,omitempty"`
	}
	tests := make([]testOut, 0, len(ch.TestNames))
	passedN := 0
	// Pass/fail comes only from the authenticated verdict the sandbox
	// entrypoint produced. Events parsed from stdout are untrusted (learner code
	// can print anything) and are used only to explain failures.
	for _, name := range ch.TestNames {
		t, ok := byName[name]
		to := testOut{Name: name, Passed: res.Verdict.Passed(name)}
		switch {
		case to.Passed:
			passedN++
		case ok && t.Passed:
			to.Message = "This result could not be verified."
		case ok:
			to.Message = clip(t.Output, 800)
		default:
			to.Message = "Test did not run."
		}
		tests = append(tests, to)
	}
	passed := res.Verdict.AllPassed(ch.TestNames) && res.ExitCode == 0 && !res.TimedOut && !res.BuildFailed
	sub, err := s.store.AddSubmission(r.Context(), u.ID, ch.ID, code, passed, passedN, len(ch.TestNames))
	if err != nil {
		return err
	}
	awarded := s.newAchievements(r.Context(), u)
	httpx.OK(w, map[string]any{
		"submissionId": sub.ID, "passed": passed, "testsPassed": passedN, "testsTotal": len(ch.TestNames),
		"tests": tests, "buildOutput": clip(buildOutput(res), 4000), "timedOut": res.TimedOut,
		"truncated": res.Truncated, "durationMs": res.DurationMS, "achievements": awarded,
	})
	return nil
}

// buildOutput extracts compiler output from `go test -json` results.
func buildOutput(res runner.Result) string {
	var b strings.Builder
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "{") {
			b.WriteString(line + "\n")
			continue
		}
		var ev struct {
			Test       string
			Output     string
			ImportPath string
		}
		if json.Unmarshal([]byte(line), &ev) == nil && ev.Test == "" && ev.Output != "" {
			out := strings.TrimSpace(ev.Output)
			if out == "" || out == "PASS" || out == "FAIL" || strings.HasPrefix(out, "ok ") || strings.HasPrefix(out, "FAIL\t") || strings.HasPrefix(out, "--- ") {
				continue
			}
			b.WriteString(ev.Output)
		}
	}
	if res.Stderr != "" {
		b.WriteString(res.Stderr)
	}
	return strings.TrimSpace(b.String())
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n…"
}

func (s *Server) challengeSolution(w http.ResponseWriter, r *http.Request, u store.User) error {
	ch := s.cat.Challenge(r.PathValue("id"))
	if ch == nil {
		return httpx.ErrNotFound
	}
	st, err := s.store.ChallengeStatuses(r.Context(), u.ID)
	if err != nil {
		return err
	}
	c := st[ch.ID]
	if !c.Passed && c.Attempts < solutionAfterAttempts {
		return httpx.Err(http.StatusForbidden, "solution_locked", "Make at least 3 attempts (or pass the tests) to unlock the solution.")
	}
	httpx.OK(w, map[string]string{"solution": ch.Solution, "explanation": ch.Explanation})
	return nil
}

func (s *Server) setProjectTask(w http.ResponseWriter, r *http.Request, u store.User) error {
	p := s.cat.Project(r.PathValue("id"))
	if p == nil {
		return httpx.ErrNotFound
	}
	taskID := r.PathValue("task")
	found := false
	for _, t := range p.Tasks {
		if t.ID == taskID {
			found = true
		}
	}
	if !found {
		return httpx.ErrNotFound
	}
	var in struct {
		Done bool `json:"done"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	if err := s.store.SetProjectTask(r.Context(), u.ID, p.ID, taskID, in.Done); err != nil {
		return err
	}
	d, err := s.store.ProjectTasks(r.Context(), u.ID)
	if err != nil {
		return err
	}
	done := map[string]bool{}
	for id := range d[p.ID] {
		done[id] = true
	}
	httpx.OK(w, map[string]any{"done": done, "completed": len(done) >= len(p.Tasks), "achievements": s.newAchievements(r.Context(), u)})
	return nil
}

func (s *Server) getSkills(w http.ResponseWriter, r *http.Request, u store.User) error {
	sum, _, _, err := s.loadProgress(r.Context(), u)
	if err != nil {
		return err
	}
	modulesBySkill := map[string][]string{}
	for _, m := range s.cat.Modules {
		modulesBySkill[m.Skill] = append(modulesBySkill[m.Skill], m.ID)
	}
	type skill struct {
		progress.SkillProgress
		Modules []string `json:"modules"`
	}
	out := make([]skill, 0, len(sum.Skills))
	for _, sp := range sum.Skills {
		ms := modulesBySkill[sp.SkillID]
		sort.Strings(ms)
		out = append(out, skill{sp, ms})
	}
	httpx.OK(w, out)
	return nil
}

func (s *Server) getAchievements(w http.ResponseWriter, r *http.Request, u store.User) error {
	sum, p, _, err := s.loadProgress(r.Context(), u)
	if err != nil {
		return err
	}
	httpx.OK(w, map[string]any{
		"achievements": sum.Achievements, "xp": sum.XP, "level": sum.Level,
		"xpIntoLevel": sum.XPIntoLevel, "xpForNextLevel": sum.XPForNextLevel, "gamification": p.Gamification,
	})
	return nil
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request, u store.User) error {
	var in struct {
		Seconds int `json:"seconds"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	if in.Seconds < 1 || in.Seconds > 60 {
		return httpx.BadRequest("seconds must be between 1 and 60.")
	}
	if ok, _ := s.pingRate.Allow(u.ID); !ok {
		httpx.OK(w, map[string]bool{"recorded": false})
		return nil
	}
	_, loc := s.profileAndLoc(r.Context(), u)
	day := time.Now().In(loc).Format("2006-01-02")
	if err := s.store.AddLearningSeconds(r.Context(), u.ID, day, in.Seconds); err != nil {
		return err
	}
	httpx.OK(w, map[string]bool{"recorded": true})
	return nil
}
