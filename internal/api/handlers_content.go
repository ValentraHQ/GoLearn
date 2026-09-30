package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/httpx"
	"github.com/valentrahq/golearn/internal/progress"
	"github.com/valentrahq/golearn/internal/search"
	"github.com/valentrahq/golearn/internal/store"
)

type lessonSummary struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Status      content.Status `json:"status"`
	Minutes     int            `json:"minutes"`
	Skill       string         `json:"skill"`
	HasExercise bool           `json:"hasExercise"`
	HasQuiz     bool           `json:"hasQuiz"`
	Completed   bool           `json:"completed"`
}

type moduleSummary struct {
	ID       string                   `json:"id"`
	Number   string                   `json:"number"`
	Title    string                   `json:"title"`
	Summary  string                   `json:"summary"`
	Track    string                   `json:"track"`
	Paths    []string                 `json:"paths"`
	Skill    string                   `json:"skill"`
	Requires []string                 `json:"requires"`
	Project  string                   `json:"project,omitempty"`
	Lessons  []lessonSummary          `json:"lessons"`
	Progress *progress.ModuleProgress `json:"progress,omitempty"`
}

func (s *Server) lessonSummaries(m *content.Module, lp map[string]store.LessonProgress) []lessonSummary {
	out := make([]lessonSummary, 0, len(m.Lessons))
	for _, l := range m.Lessons {
		ls := lessonSummary{ID: l.ID, Slug: l.Slug, Title: l.Title, Status: l.Status, Minutes: l.Minutes, Skill: l.Skill,
			HasExercise: l.Exercise != nil, HasQuiz: len(l.Quiz) > 0}
		if p, ok := lp[l.ID]; ok && p.CompletedAt != nil {
			ls.Completed = true
		}
		out = append(out, ls)
	}
	return out
}

func (s *Server) moduleSummary(m *content.Module, lp map[string]store.LessonProgress, mp *progress.ModuleProgress) moduleSummary {
	return moduleSummary{ID: m.ID, Number: m.Number, Title: m.Title, Summary: m.Summary, Track: m.Track, Paths: m.Paths,
		Skill: m.Skill, Requires: m.Requires, Project: m.Project, Lessons: s.lessonSummaries(m, lp), Progress: mp}
}

func (s *Server) learnerState(r *http.Request, u *store.User) (progress.Summary, map[string]store.LessonProgress, bool, error) {
	if u == nil {
		return progress.Summary{}, nil, false, nil
	}
	sum, _, in, err := s.loadProgress(r.Context(), *u)
	return sum, in.Lessons, true, err
}

func (s *Server) listModules(w http.ResponseWriter, r *http.Request, u *store.User) error {
	sum, lp, authed, err := s.learnerState(r, u)
	if err != nil {
		return err
	}
	out := make([]moduleSummary, 0, len(s.cat.Modules))
	for i, m := range s.cat.Modules {
		var mp *progress.ModuleProgress
		if authed {
			mp = &sum.Modules[i]
		}
		out = append(out, s.moduleSummary(m, lp, mp))
	}
	httpx.OK(w, out)
	return nil
}

func (s *Server) getModule(w http.ResponseWriter, r *http.Request, u *store.User) error {
	m := s.cat.Module(r.PathValue("id"))
	if m == nil {
		return httpx.ErrNotFound
	}
	sum, lp, authed, err := s.learnerState(r, u)
	if err != nil {
		return err
	}
	var mp *progress.ModuleProgress
	if authed {
		for i := range sum.Modules {
			if sum.Modules[i].ModuleID == m.ID {
				mp = &sum.Modules[i]
			}
		}
	}
	var statuses map[string]store.ChallengeStatus
	if authed {
		statuses, _ = s.store.ChallengeStatuses(r.Context(), u.ID)
	}
	var chs []challengeSummary
	for _, c := range s.cat.Challenges {
		if c.ModuleID == m.ID {
			chs = append(chs, challengeSummaryOf(c, statuses))
		}
	}
	var proj *projectSummary
	if p := s.cat.Project(m.Project); p != nil {
		ps := projectSummaryOf(p, nil)
		proj = &ps
	}
	httpx.OK(w, map[string]any{"module": s.moduleSummary(m, lp, mp), "challenges": chs, "project": proj})
	return nil
}

type lessonNav struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Module string `json:"module"`
	Slug   string `json:"slug"`
}

func navOf(l *content.Lesson) *lessonNav {
	if l == nil {
		return nil
	}
	return &lessonNav{ID: l.ID, Title: l.Title, Module: l.ModuleID, Slug: l.Slug}
}

type quizQuestionDTO struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Code    string   `json:"code,omitempty"`
	Options []string `json:"options,omitempty"`
}

func (s *Server) getLesson(w http.ResponseWriter, r *http.Request, u *store.User) error {
	l, err := s.lessonOr404(r)
	if err != nil {
		return err
	}
	m := s.cat.Module(l.ModuleID)
	_, lp, authed, err := s.learnerState(r, u)
	if err != nil {
		return err
	}
	var statuses map[string]store.ChallengeStatus
	var prog map[string]any
	if authed {
		statuses, _ = s.store.ChallengeStatuses(r.Context(), u.ID)
		p := lp[l.ID]
		prog = map[string]any{"completed": p.CompletedAt != nil, "exercisePassed": p.ExercisePassed, "quizPassed": p.QuizPassed}
	}
	var quiz []quizQuestionDTO
	for _, q := range l.Quiz {
		quiz = append(quiz, quizQuestionDTO{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Code: q.Code, Options: q.Options})
	}
	var chs []challengeSummary
	for _, id := range l.Challenges {
		if c := s.cat.Challenge(id); c != nil {
			chs = append(chs, challengeSummaryOf(c, statuses))
		}
	}
	var terms []map[string]string
	for _, t := range s.cat.Glossary {
		for _, rel := range t.Related {
			if rel == l.ID {
				terms = append(terms, map[string]string{"id": t.ID, "term": t.Term})
			}
		}
	}
	var exercise any
	if l.Exercise != nil {
		exercise = map[string]any{"prompt": l.Exercise.Prompt, "starter": l.Exercise.Starter, "expected": l.Exercise.Expected}
	}
	httpx.OK(w, map[string]any{
		"lesson": map[string]any{
			"id": l.ID, "slug": l.Slug, "moduleId": l.ModuleID, "title": l.Title, "minutes": l.Minutes, "skill": l.Skill,
			"objectives": l.Objectives, "concept": l.Concept, "examples": l.Examples, "exercise": exercise,
			"quiz": quiz, "takeaways": l.Takeaways,
		},
		"module":       map[string]string{"id": m.ID, "number": m.Number, "title": m.Title},
		"curriculum":   s.lessonSummaries(m, lp),
		"prev":         navOf(s.cat.PrevLesson(l.ID)),
		"next":         navOf(s.cat.NextLesson(l.ID)),
		"progress":     prog,
		"challenges":   chs,
		"glossary":     terms,
		"requirements": map[string]bool{"quiz": len(l.Quiz) > 0, "exercise": l.Exercise != nil},
	})
	return nil
}

func (s *Server) lessonSolution(w http.ResponseWriter, r *http.Request) error {
	l, err := s.lessonOr404(r)
	if err != nil {
		return err
	}
	if l.Exercise == nil {
		return httpx.ErrNotFound
	}
	httpx.OK(w, map[string]string{"solution": l.Exercise.Solution})
	return nil
}

type challengeSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	ModuleID   string `json:"moduleId"`
	Lesson     string `json:"lesson,omitempty"`
	Skill      string `json:"skill"`
	Attempts   int    `json:"attempts"`
	Passed     bool   `json:"passed"`
}

func challengeSummaryOf(c *content.Challenge, st map[string]store.ChallengeStatus) challengeSummary {
	cs := challengeSummary{ID: c.ID, Title: c.Title, Difficulty: c.Difficulty, ModuleID: c.ModuleID, Lesson: c.Lesson, Skill: c.Skill}
	if s, ok := st[c.ID]; ok {
		cs.Attempts, cs.Passed = s.Attempts, s.Passed
	}
	return cs
}

var difficultyRank = map[string]int{"beginner": 0, "intermediate": 1, "advanced": 2, "expert": 3}

func pageParams(r *http.Request, def, maxSize int) (page, size int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	size, _ = strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = def
	}
	if size > maxSize {
		size = maxSize
	}
	return
}

func (s *Server) listChallenges(w http.ResponseWriter, r *http.Request, u *store.User) error {
	q := r.URL.Query()
	var statuses map[string]store.ChallengeStatus
	if u != nil {
		statuses, _ = s.store.ChallengeStatuses(r.Context(), u.ID)
	}
	text := strings.ToLower(strings.TrimSpace(q.Get("q")))
	var list []challengeSummary
	for _, c := range s.cat.Challenges {
		if d := q.Get("difficulty"); d != "" && c.Difficulty != d {
			continue
		}
		if m := q.Get("module"); m != "" && c.ModuleID != m {
			continue
		}
		if sk := q.Get("skill"); sk != "" && c.Skill != sk {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(c.Title+" "+c.Problem), text) {
			continue
		}
		cs := challengeSummaryOf(c, statuses)
		switch q.Get("status") {
		case "passed":
			if !cs.Passed {
				continue
			}
		case "todo":
			if cs.Passed {
				continue
			}
		}
		list = append(list, cs)
	}
	if q.Get("sort") == "difficulty" {
		sort.SliceStable(list, func(i, j int) bool { return difficultyRank[list[i].Difficulty] < difficultyRank[list[j].Difficulty] })
	}
	page, size := pageParams(r, 20, 100)
	total := len(list)
	lo := min((page-1)*size, total)
	hi := min(lo+size, total)
	httpx.Page(w, list[lo:hi], httpx.Meta{Page: page, PageSize: size, Total: total})
	return nil
}

func (s *Server) getChallenge(w http.ResponseWriter, r *http.Request, u *store.User) error {
	c := s.cat.Challenge(r.PathValue("id"))
	if c == nil {
		return httpx.ErrNotFound
	}
	var statuses map[string]store.ChallengeStatus
	var lastCode string
	if u != nil {
		statuses, _ = s.store.ChallengeStatuses(r.Context(), u.ID)
		lastCode, _ = s.store.LastSubmissionCode(r.Context(), u.ID, c.ID)
	}
	sum := challengeSummaryOf(c, statuses)
	var lesson *lessonNav
	if l := s.cat.Lesson(c.Lesson); l != nil {
		lesson = navOf(l)
	}
	httpx.OK(w, map[string]any{
		"challenge": map[string]any{
			"id": c.ID, "title": c.Title, "difficulty": c.Difficulty, "moduleId": c.ModuleID, "skill": c.Skill,
			"problem": c.Problem, "input": c.Input, "expectedOutput": c.Expected, "constraints": c.Constraints,
			"starter": c.Starter, "hints": c.Hints, "testNames": c.TestNames, "explanation": c.Explanation,
		},
		"status":           sum,
		"lastCode":         lastCode,
		"lesson":           lesson,
		"solutionUnlocked": sum.Passed || sum.Attempts >= solutionAfterAttempts,
		"solutionAfter":    solutionAfterAttempts,
	})
	return nil
}

type projectSummary struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Level     string   `json:"level"`
	ModuleID  string   `json:"moduleId,omitempty"`
	Skills    []string `json:"skills"`
	Hours     int      `json:"hours"`
	Summary   string   `json:"summary"`
	TaskCount int      `json:"taskCount"`
	TasksDone int      `json:"tasksDone"`
}

func projectSummaryOf(p *content.Project, _ map[string]bool) projectSummary {
	return projectSummary{ID: p.ID, Title: p.Title, Level: p.Level, ModuleID: p.ModuleID, Skills: p.Skills, Hours: p.Hours,
		Summary: p.Summary, TaskCount: len(p.Tasks)}
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request, u *store.User) error {
	var done map[string]map[string]time.Time
	if u != nil {
		d, err := s.store.ProjectTasks(r.Context(), u.ID)
		if err != nil {
			return err
		}
		done = d
	}
	out := make([]projectSummary, 0, len(s.cat.Projects))
	for _, p := range s.cat.Projects {
		ps := projectSummaryOf(p, nil)
		ps.TasksDone = len(done[p.ID])
		out = append(out, ps)
	}
	httpx.OK(w, out)
	return nil
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request, u *store.User) error {
	p := s.cat.Project(r.PathValue("id"))
	if p == nil {
		return httpx.ErrNotFound
	}
	done := map[string]bool{}
	if u != nil {
		d, err := s.store.ProjectTasks(r.Context(), u.ID)
		if err != nil {
			return err
		}
		for id := range d[p.ID] {
			done[id] = true
		}
	}
	var module *lessonNav
	if m := s.cat.Module(p.ModuleID); m != nil {
		module = &lessonNav{ID: m.ID, Title: m.Number + " " + m.Title, Module: m.ID}
	}
	httpx.OK(w, map[string]any{"project": p, "done": done, "module": module})
	return nil
}

func (s *Server) listGlossary(w http.ResponseWriter, r *http.Request) error {
	text := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	type related struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Module string `json:"module"`
		Slug   string `json:"slug"`
	}
	type term struct {
		*content.GlossaryTerm
		RelatedLessons []related `json:"relatedLessons"`
	}
	out := []term{}
	for _, t := range s.cat.Glossary {
		if text != "" && !strings.Contains(strings.ToLower(t.Term+" "+t.Simple+" "+t.Technical), text) {
			continue
		}
		tm := term{GlossaryTerm: t, RelatedLessons: []related{}}
		for _, id := range t.Related {
			if l := s.cat.Lesson(id); l != nil {
				tm.RelatedLessons = append(tm.RelatedLessons, related{l.ID, l.Title, l.ModuleID, l.Slug})
			}
		}
		out = append(out, tm)
	}
	httpx.OK(w, out)
	return nil
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) error {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 100 {
		return httpx.BadRequest("Query is too long.")
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 30
	}
	res := s.index.Query(q, limit)
	if res == nil {
		res = []search.Result{}
	}
	httpx.OK(w, map[string]any{"query": q, "results": res})
	return nil
}
