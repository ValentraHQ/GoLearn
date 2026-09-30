package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestPlaceholderRebindForPostgres(t *testing.T) {
	pg := &Store{postgres: true}
	got := pg.q("SELECT * FROM t WHERE a = ? AND b = ? AND c = ?")
	if got != "SELECT * FROM t WHERE a = $1 AND b = $2 AND c = $3" {
		t.Errorf("got %q", got)
	}
	lite := &Store{}
	if q := lite.q("a = ?"); q != "a = ?" {
		t.Errorf("sqlite queries must be unchanged, got %q", q)
	}
}

func TestOpenRejectsUnknownScheme(t *testing.T) {
	if _, err := Open(context.Background(), "mysql://x"); err == nil {
		t.Error("expected an error for an unsupported URL")
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	s := open(t)
	if err := s.migrate(context.Background()); err != nil {
		t.Fatalf("re-running migrations: %v", err)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n); err != nil || n != 1 {
		t.Errorf("schema_migrations rows = %d, %v", n, err)
	}
}

func TestUsersAndProfiles(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, err := s.CreateUser(ctx, " Ada@Example.com ", "hash", "Ada")
	if err != nil || u.Email != "ada@example.com" {
		t.Fatalf("%+v %v", u, err)
	}
	if _, err := s.CreateUser(ctx, "ADA@example.com", "hash", "Other"); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("duplicate email: %v", err)
	}
	got, err := s.UserByEmail(ctx, "ada@EXAMPLE.com")
	if err != nil || got.ID != u.ID {
		t.Errorf("lookup: %+v %v", got, err)
	}
	if _, err := s.UserByEmail(ctx, "nobody@example.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing user: %v", err)
	}
	p, err := s.Profile(ctx, u.ID)
	if err != nil || p.DisplayName != "Ada" || p.LearningPath != "beginner" || p.Timezone != "UTC" || !p.Gamification {
		t.Errorf("default profile: %+v %v", p, err)
	}
	p.Theme, p.Gamification, p.DailyLessons = "dark", false, 3
	if err := s.UpdateProfile(ctx, u.ID, p); err != nil {
		t.Fatal(err)
	}
	if p2, _ := s.Profile(ctx, u.ID); p2.Theme != "dark" || p2.Gamification || p2.DailyLessons != 3 {
		t.Errorf("updated profile: %+v", p2)
	}
}

func TestSessions(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, _ := s.CreateUser(ctx, "a@b.co", "h", "A")
	tok, err := s.CreateSession(ctx, u.ID, time.Hour)
	if err != nil || len(tok) < 32 {
		t.Fatalf("token %q %v", tok, err)
	}
	if got, err := s.UserBySession(ctx, tok); err != nil || got.ID != u.ID {
		t.Errorf("valid session: %+v %v", got, err)
	}
	// Only a hash is stored, never the raw token.
	var n int
	s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = ?`, tok).Scan(&n)
	if n != 0 {
		t.Error("raw token must not be stored")
	}
	expired, _ := s.CreateSession(ctx, u.ID, -time.Minute)
	if _, err := s.UserBySession(ctx, expired); !errors.Is(err, ErrNotFound) {
		t.Errorf("expired session should be rejected: %v", err)
	}
	if err := s.DeleteSession(ctx, tok); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserBySession(ctx, tok); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted session still valid: %v", err)
	}
	if _, err := s.UserBySession(ctx, "garbage"); !errors.Is(err, ErrNotFound) {
		t.Errorf("garbage token: %v", err)
	}
}

func TestProgressUpserts(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, _ := s.CreateUser(ctx, "a@b.co", "h", "A")
	if err := s.MarkExercisePassed(ctx, u.ID, "m/l"); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkQuizPassed(ctx, u.ID, "m/l"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLessonCompleted(ctx, u.ID, "m/l", true); err != nil {
		t.Fatal(err)
	}
	lp, _ := s.LessonProgress(ctx, u.ID)
	got := lp["m/l"]
	if !got.ExercisePassed || !got.QuizPassed || got.CompletedAt == nil {
		t.Fatalf("%+v", got)
	}
	first := *got.CompletedAt
	_ = s.SetLessonCompleted(ctx, u.ID, "m/l", true) // re-completing keeps the timestamp
	lp, _ = s.LessonProgress(ctx, u.ID)
	if !lp["m/l"].CompletedAt.Equal(first) {
		t.Error("completion time changed on re-complete")
	}
	_ = s.SetLessonCompleted(ctx, u.ID, "m/l", false)
	lp, _ = s.LessonProgress(ctx, u.ID)
	if lp["m/l"].CompletedAt != nil || !lp["m/l"].QuizPassed {
		t.Errorf("uncomplete should only clear completion: %+v", lp["m/l"])
	}
}

func TestLearningSecondsAccumulate(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, _ := s.CreateUser(ctx, "a@b.co", "h", "A")
	for _, n := range []int{30, 30, 15} {
		if err := s.AddLearningSeconds(ctx, u.ID, "2025-03-10", n); err != nil {
			t.Fatal(err)
		}
	}
	_ = s.AddLearningSeconds(ctx, u.ID, "2025-03-11", 10)
	got, _ := s.LearningSeconds(ctx, u.ID)
	if got["2025-03-10"] != 75 || got["2025-03-11"] != 10 {
		t.Errorf("%v", got)
	}
}

func TestChallengeStatusesAndProjectsAndAchievements(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, _ := s.CreateUser(ctx, "a@b.co", "h", "A")
	_, _ = s.AddSubmission(ctx, u.ID, "c1", "code1", false, 0, 2)
	_, _ = s.AddSubmission(ctx, u.ID, "c1", "code2", true, 2, 2)
	_, _ = s.AddSubmission(ctx, u.ID, "c2", "x", false, 1, 2)
	st, err := s.ChallengeStatuses(ctx, u.ID)
	if err != nil || st["c1"].Attempts != 2 || !st["c1"].Passed || st["c1"].FirstPassAt == nil || st["c2"].Passed {
		t.Errorf("%+v %v", st, err)
	}
	if code, err := s.LastSubmissionCode(ctx, u.ID, "c1"); err != nil || code != "code2" {
		t.Errorf("last code %q %v", code, err)
	}
	if _, err := s.LastSubmissionCode(ctx, u.ID, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("%v", err)
	}

	_ = s.SetProjectTask(ctx, u.ID, "p", "t1", true)
	_ = s.SetProjectTask(ctx, u.ID, "p", "t1", true) // idempotent
	_ = s.SetProjectTask(ctx, u.ID, "p", "t2", true)
	_ = s.SetProjectTask(ctx, u.ID, "p", "t2", false)
	pt, _ := s.ProjectTasks(ctx, u.ID)
	if len(pt["p"]) != 1 {
		t.Errorf("project tasks: %v", pt)
	}

	_ = s.UnlockAchievement(ctx, u.ID, "a1")
	_ = s.UnlockAchievement(ctx, u.ID, "a1")
	ua, _ := s.UnlockedAchievements(ctx, u.ID)
	if len(ua) != 1 {
		t.Errorf("achievements: %v", ua)
	}
}

func TestDeleteUserCascades(t *testing.T) {
	s, ctx := open(t), context.Background()
	u, _ := s.CreateUser(ctx, "a@b.co", "h", "A")
	tok, _ := s.CreateSession(ctx, u.ID, time.Hour)
	_ = s.MarkQuizPassed(ctx, u.ID, "m/l")
	_ = s.AddLearningSeconds(ctx, u.ID, "2025-01-01", 5)
	if err := s.DeleteUser(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"profiles", "sessions", "lesson_progress", "learning_sessions"} {
		var n int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s rows after delete = %d (%v)", table, n, err)
		}
	}
	if _, err := s.UserBySession(ctx, tok); !errors.Is(err, ErrNotFound) {
		t.Errorf("session survived user deletion: %v", err)
	}
}
