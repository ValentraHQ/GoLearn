package progress

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/store"
)

const modA = "# Basics\nid: basics\nnumber: 01\nskill: fundamentals\npaths: beginner\nsummary: s\n\n" +
	"## One\nslug: one\nobjectives: a\ntakeaways: b\nchallenges: c1\n\n### Concept\nx\n\n### Example\n```go\npackage main\n\nfunc main() {}\n```\n\n### Check\nQ: q\nT: mcq\n- [x] a\n- [ ] b\nE: e\n\n" +
	"## Two\nslug: two\nobjectives: a\ntakeaways: b\n\n### Concept\nx\n\n### Example\n```go\npackage main\n\nfunc main() {}\n```\n\n### Check\nQ: q\nT: mcq\n- [x] a\n- [ ] b\nE: e\n"

const modB = "# Advanced\nid: advanced\nnumber: 02\nskill: concurrency\nrequires: basics\nsummary: s\n\n" +
	"## Three\nslug: three\nobjectives: a\ntakeaways: b\n\n### Concept\nx\n\n### Example\n```go\npackage main\n\nfunc main() {}\n```\n\n### Check\nQ: q\nT: mcq\n- [x] a\n- [ ] b\nE: e\n\n" +
	"## Planned\nslug: planned\nstatus: planned\n"

func catalog(t *testing.T) *content.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		"d/modules/01.md":   {Data: []byte(modA)},
		"d/modules/02.md":   {Data: []byte(modB)},
		"d/challenges/c.md": {Data: []byte("# C\n\n## c1\ntitle: C1\ndifficulty: advanced\nmodule: basics\nlesson: basics/one\nskill: fundamentals\n\n### Problem\np\n\n### Starter\n```go\npackage main\n```\n\n### Tests\n```go\npackage main\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n```\n\n### Solution\n```go\npackage main\n```\n")},
	}
	c, err := content.LoadFS(fsys, "d")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var utc = time.UTC

func at(day string, hour int) time.Time {
	d, _ := time.ParseInLocation("2006-01-02", day, utc)
	return d.Add(time.Duration(hour) * time.Hour)
}

func done(day string) *time.Time { t := at(day, 12); return &t }

func TestStreaks(t *testing.T) {
	now := at("2025-03-10", 15)
	days := func(ds ...string) map[string]bool {
		m := map[string]bool{}
		for _, d := range ds {
			m[d] = true
		}
		return m
	}
	tests := []struct {
		name         string
		days         map[string]bool
		cur, longest int
	}{
		{"none", nil, 0, 0},
		{"today only", days("2025-03-10"), 1, 1},
		{"yesterday keeps the streak alive", days("2025-03-08", "2025-03-09"), 2, 2},
		{"gap breaks it", days("2025-03-07", "2025-03-10"), 1, 1},
		{"old streak, none recent", days("2025-03-01", "2025-03-02", "2025-03-03"), 0, 3},
		{"three days", days("2025-03-08", "2025-03-09", "2025-03-10"), 3, 3},
		{"month boundary", days("2025-02-28", "2025-03-01"), 0, 2},
	}
	for _, tc := range tests {
		cur, longest := streaks(tc.days, now)
		if cur != tc.cur || longest != tc.longest {
			t.Errorf("%s: got (%d,%d), want (%d,%d)", tc.name, cur, longest, tc.cur, tc.longest)
		}
	}
}

func TestSkillLevelFor(t *testing.T) {
	tests := []struct {
		done, total int
		want        SkillLevel
	}{
		{0, 10, NotStarted}, {1, 10, Learning}, {3, 10, Learning}, {4, 10, Practicing},
		{6, 10, Practicing}, {7, 10, Proficient}, {9, 10, Proficient}, {10, 10, Mastered}, {0, 0, NotStarted},
	}
	for _, tc := range tests {
		if got := SkillLevelFor(tc.done, tc.total); got != tc.want {
			t.Errorf("SkillLevelFor(%d,%d) = %s, want %s", tc.done, tc.total, got, tc.want)
		}
	}
}

func TestModuleStatesAndPrerequisites(t *testing.T) {
	c := catalog(t)
	now := at("2025-03-10", 12)

	// Nothing done, beginner path: advanced is locked behind basics.
	sum := Compute(c, Input{Path: "beginner"}, now, utc)
	byID := map[string]ModuleProgress{}
	for _, m := range sum.Modules {
		byID[m.ModuleID] = m
	}
	if byID["basics"].State != Available || byID["advanced"].State != Locked {
		t.Fatalf("states: basics=%s advanced=%s", byID["basics"].State, byID["advanced"].State)
	}
	if got := byID["advanced"].Blockers; len(got) != 1 || got[0] != "basics" {
		t.Errorf("blockers = %v", got)
	}

	// Complete one of two basics lessons (50% < 70%): still locked, basics in progress.
	in := Input{Path: "beginner", Lessons: map[string]store.LessonProgress{"basics/one": {LessonID: "basics/one", CompletedAt: done("2025-03-10")}}}
	sum = Compute(c, in, now, utc)
	if sum.Modules[0].State != InProgress || sum.Modules[0].Percent != 50 || sum.Modules[1].State != Locked {
		t.Errorf("half done: %+v", sum.Modules)
	}

	// Complete both: basics completed, advanced unlocked.
	in.Lessons["basics/two"] = store.LessonProgress{LessonID: "basics/two", CompletedAt: done("2025-03-10")}
	sum = Compute(c, in, now, utc)
	if sum.Modules[0].State != Completed || sum.Modules[1].State != Available || sum.ModulesCompleted != 1 {
		t.Errorf("both done: %+v", sum.Modules)
	}
	if sum.Continue == nil || sum.Continue.LessonID != "advanced/three" {
		t.Errorf("continue = %+v", sum.Continue)
	}
}

func TestPlannedLessonsDoNotCount(t *testing.T) {
	c := catalog(t)
	sum := Compute(c, Input{Path: "pro"}, at("2025-03-10", 12), utc)
	if sum.LessonsTotal != 3 {
		t.Errorf("LessonsTotal = %d, want 3 (planned lessons excluded)", sum.LessonsTotal)
	}
	for _, m := range sum.Modules {
		if m.ModuleID == "advanced" && m.Total != 1 {
			t.Errorf("advanced published lessons = %d, want 1", m.Total)
		}
	}
}

func TestPathFilterIgnoresOutOfPathPrerequisites(t *testing.T) {
	c := catalog(t)
	// "basics" is beginner-only, so a pro learner must not be locked out of "advanced" by it.
	sum := Compute(c, Input{Path: "pro"}, at("2025-03-10", 12), utc)
	for _, m := range sum.Modules {
		if m.ModuleID == "advanced" && m.State != Available {
			t.Errorf("pro path: advanced state = %s, want available", m.State)
		}
	}
	if sum.Continue == nil || sum.Continue.ModuleID != "advanced" {
		t.Errorf("pro path should start at the first module in its path, got %+v", sum.Continue)
	}
}

func TestXPLevelsAndDailyGoals(t *testing.T) {
	c := catalog(t)
	now := at("2025-03-10", 15)
	in := Input{
		Path:  "beginner",
		Goals: Goals{Lessons: 1, Challenges: 1, Quizzes: 1},
		Lessons: map[string]store.LessonProgress{
			"basics/one": {LessonID: "basics/one", ExercisePassed: true, QuizPassed: true, CompletedAt: done("2025-03-10")},
		},
		Challenges: map[string]store.ChallengeStatus{"c1": {ChallengeID: "c1", Passed: true, Attempts: 2}},
		Quizzes:    []store.QuizAttempt{{LessonID: "basics/one", Score: 1, Total: 1, Passed: true, CreatedAt: at("2025-03-10", 9)}},
		Events: []store.Event{
			{Kind: "lesson", Ref: "basics/one", At: at("2025-03-10", 12)},
			{Kind: "challenge", Ref: "c1", At: at("2025-03-10", 13)},
			{Kind: "quiz", Ref: "basics/one", At: at("2025-03-10", 9)},
			{Kind: "lesson", Ref: "old", At: at("2025-03-09", 9)},
		},
	}
	sum := Compute(c, in, now, utc)
	// lesson 20 + exercise 10 + quiz 10 + advanced challenge 50
	if sum.XP != 90 {
		t.Errorf("XP = %d, want 90", sum.XP)
	}
	if sum.Level != 1 || sum.XPIntoLevel != 90 || sum.XPForNextLevel != 100 {
		t.Errorf("level=%d into=%d next=%d", sum.Level, sum.XPIntoLevel, sum.XPForNextLevel)
	}
	if sum.Daily.Lessons.Done != 1 || sum.Daily.Challenges.Done != 1 || sum.Daily.Quizzes.Done != 1 {
		t.Errorf("daily = %+v", sum.Daily)
	}
	if sum.CurrentStreak != 2 {
		t.Errorf("streak = %d, want 2 (yesterday + today)", sum.CurrentStreak)
	}
	if sum.QuizAverage != 100 || sum.QuizzesPassed != 1 {
		t.Errorf("quiz avg=%d passed=%d", sum.QuizAverage, sum.QuizzesPassed)
	}
}

func TestAchievementsUnlockOnceAndPersist(t *testing.T) {
	c := catalog(t)
	now := at("2025-03-10", 12)
	in := Input{Path: "beginner", Lessons: map[string]store.LessonProgress{
		"basics/one": {LessonID: "basics/one", ExercisePassed: true, CompletedAt: done("2025-03-10")},
	}}
	sum := Compute(c, in, now, utc)
	got := map[string]bool{}
	for _, id := range sum.NewlyUnlocked {
		got[id] = true
	}
	if !got["first-program"] || !got["first-lesson"] {
		t.Fatalf("newly unlocked = %v", sum.NewlyUnlocked)
	}
	// Once persisted they are not reported as new again, and keep their original time.
	earlier := at("2025-03-01", 8)
	in.Unlocked = map[string]time.Time{"first-lesson": earlier, "first-program": earlier}
	sum = Compute(c, in, now, utc)
	for _, id := range sum.NewlyUnlocked {
		if id == "first-lesson" || id == "first-program" {
			t.Errorf("%s reported as new although already persisted", id)
		}
	}
	for _, a := range sum.Achievements {
		if a.ID == "first-lesson" && (a.UnlockedAt == nil || !a.UnlockedAt.Equal(earlier)) {
			t.Errorf("first-lesson unlock time not preserved: %v", a.UnlockedAt)
		}
	}
}

func TestSkillProgressCountsLessonsAndChallenges(t *testing.T) {
	c := catalog(t)
	in := Input{Path: "beginner",
		Lessons:    map[string]store.LessonProgress{"basics/one": {LessonID: "basics/one", CompletedAt: done("2025-03-10")}},
		Challenges: map[string]store.ChallengeStatus{"c1": {ChallengeID: "c1", Passed: true}},
	}
	sum := Compute(c, in, at("2025-03-10", 12), utc)
	for _, s := range sum.Skills {
		if s.SkillID == "fundamentals" {
			// 2 lessons + 1 challenge tagged fundamentals; 1 lesson + 1 challenge done.
			if s.Total != 3 || s.Done != 2 || s.Level != Practicing {
				t.Errorf("fundamentals = %+v", s)
			}
		}
		if s.SkillID == "kubernetes" && s.Level != NotStarted {
			t.Errorf("kubernetes should be not started: %+v", s)
		}
	}
}
