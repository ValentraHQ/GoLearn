package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type LessonProgress struct {
	LessonID       string
	ExercisePassed bool
	QuizPassed     bool
	CompletedAt    *time.Time
}

func (s *Store) LessonProgress(ctx context.Context, userID string) (map[string]LessonProgress, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT lesson_id, exercise_passed, quiz_passed, completed_at FROM lesson_progress WHERE user_id = ?`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]LessonProgress{}
	for rows.Next() {
		var lp LessonProgress
		var ex, qz int
		var c sql.NullInt64
		if err := rows.Scan(&lp.LessonID, &ex, &qz, &c); err != nil {
			return nil, err
		}
		lp.ExercisePassed, lp.QuizPassed = ex != 0, qz != 0
		if c.Valid {
			t := time.Unix(c.Int64, 0)
			lp.CompletedAt = &t
		}
		out[lp.LessonID] = lp
	}
	return out, rows.Err()
}

func (s *Store) ensureLesson(ctx context.Context, userID, lessonID string) error {
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO lesson_progress (user_id, lesson_id, updated_at) VALUES (?, ?, ?) ON CONFLICT (user_id, lesson_id) DO NOTHING`), userID, lessonID, now())
	return err
}

func (s *Store) MarkExercisePassed(ctx context.Context, userID, lessonID string) error {
	if err := s.ensureLesson(ctx, userID, lessonID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, s.q(`UPDATE lesson_progress SET exercise_passed = 1, updated_at = ? WHERE user_id = ? AND lesson_id = ?`), now(), userID, lessonID)
	return err
}

func (s *Store) MarkQuizPassed(ctx context.Context, userID, lessonID string) error {
	if err := s.ensureLesson(ctx, userID, lessonID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, s.q(`UPDATE lesson_progress SET quiz_passed = 1, updated_at = ? WHERE user_id = ? AND lesson_id = ?`), now(), userID, lessonID)
	return err
}

// SetLessonCompleted marks or un-marks completion. Re-completing keeps the original timestamp.
func (s *Store) SetLessonCompleted(ctx context.Context, userID, lessonID string, done bool) error {
	if err := s.ensureLesson(ctx, userID, lessonID); err != nil {
		return err
	}
	if done {
		_, err := s.db.ExecContext(ctx, s.q(`UPDATE lesson_progress SET completed_at = COALESCE(completed_at, ?), updated_at = ? WHERE user_id = ? AND lesson_id = ?`), now(), now(), userID, lessonID)
		return err
	}
	_, err := s.db.ExecContext(ctx, s.q(`UPDATE lesson_progress SET completed_at = NULL, updated_at = ? WHERE user_id = ? AND lesson_id = ?`), now(), userID, lessonID)
	return err
}

type Submission struct {
	ID          string
	ChallengeID string
	Passed      bool
	TestsPassed int
	TestsTotal  int
	CreatedAt   time.Time
}

func (s *Store) AddSubmission(ctx context.Context, userID, challengeID, code string, passed bool, tp, tt int) (Submission, error) {
	sub := Submission{ID: newID(), ChallengeID: challengeID, Passed: passed, TestsPassed: tp, TestsTotal: tt, CreatedAt: time.Unix(now(), 0)}
	p := 0
	if passed {
		p = 1
	}
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO challenge_submissions (id, user_id, challenge_id, code, passed, tests_passed, tests_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
		sub.ID, userID, challengeID, code, p, tp, tt, sub.CreatedAt.Unix())
	return sub, err
}

// ChallengeStatus summarises attempts per challenge for a user.
type ChallengeStatus struct {
	ChallengeID string
	Attempts    int
	Passed      bool
	FirstPassAt *time.Time
}

func (s *Store) ChallengeStatuses(ctx context.Context, userID string) (map[string]ChallengeStatus, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT challenge_id, COUNT(*), MAX(passed), MIN(CASE WHEN passed = 1 THEN created_at END) FROM challenge_submissions WHERE user_id = ? GROUP BY challenge_id`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]ChallengeStatus{}
	for rows.Next() {
		var cs ChallengeStatus
		var p int
		var first sql.NullInt64
		if err := rows.Scan(&cs.ChallengeID, &cs.Attempts, &p, &first); err != nil {
			return nil, err
		}
		cs.Passed = p != 0
		if first.Valid {
			t := time.Unix(first.Int64, 0)
			cs.FirstPassAt = &t
		}
		out[cs.ChallengeID] = cs
	}
	return out, rows.Err()
}

func (s *Store) LastSubmissionCode(ctx context.Context, userID, challengeID string) (string, error) {
	var code string
	err := s.db.QueryRowContext(ctx, s.q(`SELECT code FROM challenge_submissions WHERE user_id = ? AND challenge_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`), userID, challengeID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return code, err
}

type QuizAttempt struct {
	LessonID  string
	Score     int
	Total     int
	Passed    bool
	CreatedAt time.Time
}

func (s *Store) AddQuizAttempt(ctx context.Context, userID, lessonID string, score, total int, passed bool) error {
	p := 0
	if passed {
		p = 1
	}
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO quiz_attempts (id, user_id, lesson_id, score, total, passed, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`),
		newID(), userID, lessonID, score, total, p, now())
	return err
}

func (s *Store) QuizAttempts(ctx context.Context, userID string) ([]QuizAttempt, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT lesson_id, score, total, passed, created_at FROM quiz_attempts WHERE user_id = ? ORDER BY created_at`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []QuizAttempt
	for rows.Next() {
		var a QuizAttempt
		var p int
		var ts int64
		if err := rows.Scan(&a.LessonID, &a.Score, &a.Total, &p, &ts); err != nil {
			return nil, err
		}
		a.Passed = p != 0
		a.CreatedAt = time.Unix(ts, 0)
		out = append(out, a)
	}
	return out, rows.Err()
}

// ProjectTasks returns done task IDs per project.
func (s *Store) ProjectTasks(ctx context.Context, userID string) (map[string]map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT project_id, task_id, done_at FROM project_progress WHERE user_id = ?`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]map[string]time.Time{}
	for rows.Next() {
		var p, t string
		var ts int64
		if err := rows.Scan(&p, &t, &ts); err != nil {
			return nil, err
		}
		if out[p] == nil {
			out[p] = map[string]time.Time{}
		}
		out[p][t] = time.Unix(ts, 0)
	}
	return out, rows.Err()
}

func (s *Store) SetProjectTask(ctx context.Context, userID, projectID, taskID string, done bool) error {
	if done {
		_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO project_progress (user_id, project_id, task_id, done_at) VALUES (?, ?, ?, ?) ON CONFLICT (user_id, project_id, task_id) DO NOTHING`), userID, projectID, taskID, now())
		return err
	}
	_, err := s.db.ExecContext(ctx, s.q(`DELETE FROM project_progress WHERE user_id = ? AND project_id = ? AND task_id = ?`), userID, projectID, taskID)
	return err
}

func (s *Store) UnlockedAchievements(ctx context.Context, userID string) (map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT achievement_id, unlocked_at FROM user_achievements WHERE user_id = ?`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var id string
		var ts int64
		if err := rows.Scan(&id, &ts); err != nil {
			return nil, err
		}
		out[id] = time.Unix(ts, 0)
	}
	return out, rows.Err()
}

func (s *Store) UnlockAchievement(ctx context.Context, userID, id string) error {
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO user_achievements (user_id, achievement_id, unlocked_at) VALUES (?, ?, ?) ON CONFLICT (user_id, achievement_id) DO NOTHING`), userID, id, now())
	return err
}

// AddLearningSeconds accrues time to the given UTC day (YYYY-MM-DD).
func (s *Store) AddLearningSeconds(ctx context.Context, userID, day string, seconds int) error {
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO learning_sessions (user_id, day, seconds) VALUES (?, ?, ?) ON CONFLICT (user_id, day) DO UPDATE SET seconds = learning_sessions.seconds + excluded.seconds`), userID, day, seconds)
	return err
}

// LearningSeconds returns total seconds and per-day seconds.
func (s *Store) LearningSeconds(ctx context.Context, userID string) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, s.q(`SELECT day, seconds FROM learning_sessions WHERE user_id = ?`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var d string
		var n int
		if err := rows.Scan(&d, &n); err != nil {
			return nil, err
		}
		out[d] = n
	}
	return out, rows.Err()
}

// Event is a dated learner action used for streaks and daily goals.
type Event struct {
	Kind string // lesson, challenge, attempt, quiz, quiz_attempt
	Ref  string
	At   time.Time
}

func (s *Store) Events(ctx context.Context, userID string) ([]Event, error) {
	var out []Event
	collect := func(kind, query string) error {
		rows, err := s.db.QueryContext(ctx, s.q(query), userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ref string
			var ts int64
			if err := rows.Scan(&ref, &ts); err != nil {
				return err
			}
			out = append(out, Event{Kind: kind, Ref: ref, At: time.Unix(ts, 0)})
		}
		return rows.Err()
	}
	for _, q := range []struct{ kind, sql string }{
		{"lesson", `SELECT lesson_id, completed_at FROM lesson_progress WHERE user_id = ? AND completed_at IS NOT NULL`},
		{"challenge", `SELECT challenge_id, created_at FROM challenge_submissions WHERE user_id = ? AND passed = 1`},
		{"attempt", `SELECT challenge_id, created_at FROM challenge_submissions WHERE user_id = ?`},
		{"quiz", `SELECT lesson_id, created_at FROM quiz_attempts WHERE user_id = ? AND passed = 1`},
		{"quiz_attempt", `SELECT lesson_id, created_at FROM quiz_attempts WHERE user_id = ?`},
	} {
		if err := collect(q.kind, q.sql); err != nil {
			return nil, err
		}
	}
	return out, nil
}
