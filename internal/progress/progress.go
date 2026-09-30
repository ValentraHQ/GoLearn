// Package progress derives everything a learner sees about their progress
// (module states, skills, streaks, XP, achievements) from raw stored records
// and the curriculum. It is pure: no I/O, so it is straightforward to test.
package progress

import (
	"math"
	"sort"
	"time"

	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/store"
)

// Input is the raw learner state loaded from the store.
type Input struct {
	Lessons    map[string]store.LessonProgress
	Challenges map[string]store.ChallengeStatus
	Quizzes    []store.QuizAttempt
	Projects   map[string]map[string]time.Time
	Seconds    map[string]int // by local day
	Events     []store.Event
	Unlocked   map[string]time.Time
	Path       string // beginner | pro
	Goals      Goals
}

type Goals struct{ Lessons, Challenges, Quizzes int }

type ModuleState string

const (
	Completed  ModuleState = "completed"
	InProgress ModuleState = "in_progress"
	Available  ModuleState = "available"
	Locked     ModuleState = "locked"
	Planned    ModuleState = "planned" // no published lessons yet
)

// PrereqThreshold is the fraction of a required module that must be complete
// before dependent modules stop showing as locked. Locking is advisory.
const PrereqThreshold = 0.7

type ModuleProgress struct {
	ModuleID string      `json:"moduleId"`
	State    ModuleState `json:"state"`
	Total    int         `json:"totalLessons"`
	Done     int         `json:"completedLessons"`
	Percent  int         `json:"percent"`
	Blockers []string    `json:"blockedBy,omitempty"`
	InPath   bool        `json:"inPath"`
}

type SkillLevel string

const (
	NotStarted SkillLevel = "not_started"
	Learning   SkillLevel = "learning"
	Practicing SkillLevel = "practicing"
	Proficient SkillLevel = "proficient"
	Mastered   SkillLevel = "mastered"
)

type SkillProgress struct {
	SkillID string     `json:"skillId"`
	Name    string     `json:"name"`
	Level   SkillLevel `json:"level"`
	Done    int        `json:"done"`
	Total   int        `json:"total"`
	Percent int        `json:"percent"`
}

type DailyGoal struct {
	Lessons    Count `json:"lessons"`
	Challenges Count `json:"challenges"`
	Quizzes    Count `json:"quizzes"`
}

type Count struct {
	Done int `json:"done"`
	Goal int `json:"goal"`
}

type Achievement struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlockedAt,omitempty"`
}

type Summary struct {
	LessonsCompleted    int `json:"lessonsCompleted"`
	LessonsTotal        int `json:"lessonsTotal"`
	ModulesCompleted    int `json:"modulesCompleted"`
	ChallengesCompleted int `json:"challengesCompleted"`
	ChallengesTotal     int `json:"challengesTotal"`
	ProjectsCompleted   int `json:"projectsCompleted"`
	ProjectsTotal       int `json:"projectsTotal"`
	QuizzesPassed       int `json:"quizzesPassed"`
	QuizAverage         int `json:"quizAveragePercent"` // best score per lesson, averaged
	CurrentStreak       int `json:"currentStreak"`
	LongestStreak       int `json:"longestStreak"`
	LearningSeconds     int `json:"learningSeconds"`
	OverallPercent      int `json:"overallPercent"`
	XP                  int `json:"xp"`
	Level               int `json:"level"`
	XPIntoLevel         int `json:"xpIntoLevel"`
	XPForNextLevel      int `json:"xpForNextLevel"`

	Modules      []ModuleProgress `json:"modules"`
	Skills       []SkillProgress  `json:"skills"`
	Daily        DailyGoal        `json:"daily"`
	Achievements []Achievement    `json:"achievements"`
	Continue     *Continue        `json:"continue,omitempty"`

	// NewlyUnlocked lists achievement IDs earned by this state that were not yet persisted.
	NewlyUnlocked []string `json:"-"`
}

type Continue struct {
	ModuleID    string `json:"moduleId"`
	ModuleTitle string `json:"moduleTitle"`
	LessonID    string `json:"lessonId"`
	LessonTitle string `json:"lessonTitle"`
	Started     bool   `json:"started"`
}

func XPForChallenge(difficulty string) int {
	switch difficulty {
	case "beginner":
		return 10
	case "intermediate":
		return 25
	case "advanced":
		return 50
	default:
		return 100
	}
}

const (
	xpLesson   = 20
	xpExercise = 10
	xpQuiz     = 10
	xpProject  = 150
)

func inPath(m *content.Module, path string) bool {
	if path == "" {
		return true
	}
	for _, p := range m.Paths {
		if p == path {
			return true
		}
	}
	return false
}

func Compute(cat *content.Catalog, in Input, now time.Time, loc *time.Location) Summary {
	if loc == nil {
		loc = time.UTC
	}
	var s Summary
	s.LessonsTotal = len(cat.PublishedLessons())
	s.ChallengesTotal = len(cat.Challenges)
	s.ProjectsTotal = len(cat.Projects)

	done := func(id string) bool {
		lp, ok := in.Lessons[id]
		return ok && lp.CompletedAt != nil
	}

	// Modules.
	pct := map[string]float64{}
	byID := map[string]*ModuleProgress{}
	for _, m := range cat.Modules {
		mp := ModuleProgress{ModuleID: m.ID, InPath: inPath(m, in.Path)}
		for _, l := range m.Lessons {
			if l.Status != content.StatusPublished {
				continue
			}
			mp.Total++
			if done(l.ID) {
				mp.Done++
			}
		}
		if mp.Total > 0 {
			mp.Percent = mp.Done * 100 / mp.Total
			pct[m.ID] = float64(mp.Done) / float64(mp.Total)
		}
		s.Modules = append(s.Modules, mp)
	}
	for i := range s.Modules {
		byID[s.Modules[i].ModuleID] = &s.Modules[i]
	}
	for i := range s.Modules {
		mp := &s.Modules[i]
		m := cat.Module(mp.ModuleID)
		switch {
		case mp.Total == 0:
			mp.State = Planned
		case mp.Done == mp.Total:
			mp.State = Completed
			s.ModulesCompleted++
		case mp.Done > 0 || moduleTouched(m, in):
			mp.State = InProgress
		default:
			for _, r := range m.Requires {
				if rp := byID[r]; rp != nil && rp.Total > 0 && pct[r] < PrereqThreshold {
					mp.Blockers = append(mp.Blockers, r)
				}
			}
			if len(mp.Blockers) > 0 {
				mp.State = Locked
			} else {
				mp.State = Available
			}
		}
	}

	// Lessons, XP.
	xp := 0
	for id, lp := range in.Lessons {
		if cat.Lesson(id) == nil {
			continue
		}
		if lp.CompletedAt != nil {
			s.LessonsCompleted++
			xp += xpLesson
		}
		if lp.ExercisePassed {
			xp += xpExercise
		}
		if lp.QuizPassed {
			xp += xpQuiz
		}
	}
	if s.LessonsTotal > 0 {
		s.OverallPercent = s.LessonsCompleted * 100 / s.LessonsTotal
	}

	// Challenges.
	for id, cs := range in.Challenges {
		ch := cat.Challenge(id)
		if ch != nil && cs.Passed {
			s.ChallengesCompleted++
			xp += XPForChallenge(ch.Difficulty)
		}
	}

	// Quizzes: best attempt per lesson.
	best := map[string]int{}
	passed := map[string]bool{}
	for _, a := range in.Quizzes {
		if a.Total == 0 {
			continue
		}
		p := a.Score * 100 / a.Total
		if p > best[a.LessonID] {
			best[a.LessonID] = p
		}
		if a.Passed {
			passed[a.LessonID] = true
		}
	}
	s.QuizzesPassed = len(passed)
	if len(best) > 0 {
		sum := 0
		for _, p := range best {
			sum += p
		}
		s.QuizAverage = sum / len(best)
	}

	// Projects.
	projectDone := map[string]bool{}
	for _, p := range cat.Projects {
		if len(in.Projects[p.ID]) >= len(p.Tasks) && len(p.Tasks) > 0 {
			projectDone[p.ID] = true
			s.ProjectsCompleted++
			xp += xpProject
		}
	}

	// Time, streaks, daily goals.
	for _, sec := range in.Seconds {
		s.LearningSeconds += sec
	}
	days := map[string]bool{}
	for d, sec := range in.Seconds {
		if sec > 0 {
			days[d] = true
		}
	}
	today := now.In(loc).Format("2006-01-02")
	var todayLessons, todayQuizzes int
	todayChallenges := map[string]bool{}
	for _, e := range in.Events {
		d := e.At.In(loc).Format("2006-01-02")
		days[d] = true
		if d != today {
			continue
		}
		switch e.Kind {
		case "lesson":
			todayLessons++
		case "challenge":
			todayChallenges[e.Ref] = true
		case "quiz":
			todayQuizzes++
		}
	}
	s.CurrentStreak, s.LongestStreak = streaks(days, now.In(loc))
	s.Daily = DailyGoal{
		Lessons:    Count{todayLessons, in.Goals.Lessons},
		Challenges: Count{len(todayChallenges), in.Goals.Challenges},
		Quizzes:    Count{todayQuizzes, in.Goals.Quizzes},
	}

	// Skills.
	s.Skills = computeSkills(cat, in, done)

	// XP/level: level n requires 100*(n-1)^2 XP in total.
	s.XP = xp
	s.Level = int(math.Sqrt(float64(xp)/100)) + 1
	lo := 100 * (s.Level - 1) * (s.Level - 1)
	hi := 100 * s.Level * s.Level
	s.XPIntoLevel, s.XPForNextLevel = xp-lo, hi-lo

	// Achievements.
	ctx := achievementContext{cat: cat, in: in, sum: &s, done: done, projectDone: projectDone, perfectQuizzes: perfectQuizzes(in.Quizzes)}
	for _, def := range achievementDefs {
		a := Achievement{ID: def.ID, Title: def.Title, Description: def.Description, Category: def.Category}
		if at, ok := in.Unlocked[def.ID]; ok {
			a.Unlocked, a.UnlockedAt = true, &at
		} else if def.Check(ctx) {
			a.Unlocked = true
			t := now
			a.UnlockedAt = &t
			s.NewlyUnlocked = append(s.NewlyUnlocked, def.ID)
		}
		s.Achievements = append(s.Achievements, a)
	}

	s.Continue = nextLesson(cat, in, done)
	return s
}

func moduleTouched(m *content.Module, in Input) bool {
	for _, l := range m.Lessons {
		if lp, ok := in.Lessons[l.ID]; ok && (lp.ExercisePassed || lp.QuizPassed) {
			return true
		}
	}
	return false
}

// streaks returns the current and longest run of consecutive active days. The
// current streak stays alive if the most recent activity was yesterday.
func streaks(days map[string]bool, now time.Time) (current, longest int) {
	if len(days) == 0 {
		return 0, 0
	}
	keys := make([]string, 0, len(days))
	for d := range days {
		keys = append(keys, d)
	}
	sort.Strings(keys)
	run := 0
	var prev time.Time
	for _, k := range keys {
		t, err := time.ParseInLocation("2006-01-02", k, time.UTC)
		if err != nil {
			continue
		}
		if !prev.IsZero() && t.Sub(prev) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
		prev = t
	}
	todayKey := now.Format("2006-01-02")
	yesterdayKey := now.AddDate(0, 0, -1).Format("2006-01-02")
	start := ""
	switch {
	case days[todayKey]:
		start = todayKey
	case days[yesterdayKey]:
		start = yesterdayKey
	default:
		return 0, longest
	}
	t, _ := time.ParseInLocation("2006-01-02", start, time.UTC)
	for days[t.Format("2006-01-02")] {
		current++
		t = t.AddDate(0, 0, -1)
	}
	return current, longest
}

func computeSkills(cat *content.Catalog, in Input, done func(string) bool) []SkillProgress {
	total := map[string]int{}
	got := map[string]int{}
	for _, l := range cat.PublishedLessons() {
		total[l.Skill]++
		if done(l.ID) {
			got[l.Skill]++
		}
	}
	for _, ch := range cat.Challenges {
		total[ch.Skill]++
		if cs, ok := in.Challenges[ch.ID]; ok && cs.Passed {
			got[ch.Skill]++
		}
	}
	out := make([]SkillProgress, 0, len(cat.Skills))
	for _, sk := range cat.Skills {
		sp := SkillProgress{SkillID: sk.ID, Name: sk.Name, Done: got[sk.ID], Total: total[sk.ID]}
		if sp.Total > 0 {
			sp.Percent = sp.Done * 100 / sp.Total
		}
		sp.Level = SkillLevelFor(sp.Done, sp.Total)
		out = append(out, sp)
	}
	return out
}

// SkillLevelFor maps completed/total practice items to a level.
func SkillLevelFor(done, total int) SkillLevel {
	switch {
	case done == 0 || total == 0:
		return NotStarted
	case done >= total:
		return Mastered
	case done*100/total >= 67:
		return Proficient
	case done*100/total >= 34:
		return Practicing
	default:
		return Learning
	}
}

func perfectQuizzes(as []store.QuizAttempt) int {
	seen := map[string]bool{}
	for _, a := range as {
		if a.Total > 0 && a.Score == a.Total {
			seen[a.LessonID] = true
		}
	}
	return len(seen)
}

// nextLesson picks the first incomplete lesson on the learner's path.
func nextLesson(cat *content.Catalog, in Input, done func(string) bool) *Continue {
	pick := func(pathOnly bool) *Continue {
		for _, l := range cat.PublishedLessons() {
			m := cat.Module(l.ModuleID)
			if pathOnly && !inPath(m, in.Path) {
				continue
			}
			if !done(l.ID) {
				lp := in.Lessons[l.ID]
				return &Continue{ModuleID: m.ID, ModuleTitle: m.Title, LessonID: l.ID, LessonTitle: l.Title,
					Started: lp.ExercisePassed || lp.QuizPassed}
			}
		}
		return nil
	}
	if c := pick(true); c != nil {
		return c
	}
	return pick(false)
}
