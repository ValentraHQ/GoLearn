package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email already registered")
)

func now() int64 { return time.Now().Unix() }

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Profile struct {
	UserID          string `json:"-"`
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	LearningPath    string `json:"learningPath"`
	Theme           string `json:"theme"`
	DailyLessons    int    `json:"dailyLessons"`
	DailyChallenges int    `json:"dailyChallenges"`
	DailyQuizzes    int    `json:"dailyQuizzes"`
	Gamification    bool   `json:"gamification"`
	Timezone        string `json:"timezone"`
}

func (s *Store) CreateUser(ctx context.Context, email, hash, displayName string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u := User{ID: newID(), Email: email, PasswordHash: hash, Role: "learner", CreatedAt: time.Unix(now(), 0)}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRowContext(ctx, s.q(`SELECT COUNT(*) FROM users WHERE email = ?`), email).Scan(&n); err != nil {
		return User{}, err
	}
	if n > 0 {
		return User{}, ErrEmailTaken
	}
	if _, err := tx.ExecContext(ctx, s.q(`INSERT INTO users (id, email, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?)`),
		u.ID, u.Email, u.PasswordHash, u.Role, u.CreatedAt.Unix()); err != nil {
		return User{}, err
	}
	if _, err := tx.ExecContext(ctx, s.q(`INSERT INTO profiles (user_id, display_name, updated_at) VALUES (?, ?, ?)`),
		u.ID, displayName, now()); err != nil {
		return User{}, err
	}
	return u, tx.Commit()
}

func (s *Store) UserByEmail(ctx context.Context, email string) (User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx, s.q(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = ?`), strings.ToLower(strings.TrimSpace(email))))
}

func (s *Store) UserByID(ctx context.Context, id string) (User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx, s.q(`SELECT id, email, password_hash, role, created_at FROM users WHERE id = ?`), id))
}

func (s *Store) scanUser(row *sql.Row) (User, error) {
	var u User
	var ts int64
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &ts)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	u.CreatedAt = time.Unix(ts, 0)
	return u, err
}

func (s *Store) Profile(ctx context.Context, userID string) (Profile, error) {
	var p Profile
	var g int
	err := s.db.QueryRowContext(ctx, s.q(`SELECT user_id, display_name, bio, learning_path, theme, daily_lessons, daily_challenges, daily_quizzes, gamification, timezone FROM profiles WHERE user_id = ?`), userID).
		Scan(&p.UserID, &p.DisplayName, &p.Bio, &p.LearningPath, &p.Theme, &p.DailyLessons, &p.DailyChallenges, &p.DailyQuizzes, &g, &p.Timezone)
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	p.Gamification = g != 0
	return p, err
}

func (s *Store) UpdateProfile(ctx context.Context, userID string, p Profile) error {
	g := 0
	if p.Gamification {
		g = 1
	}
	_, err := s.db.ExecContext(ctx, s.q(`UPDATE profiles SET display_name = ?, bio = ?, learning_path = ?, theme = ?, daily_lessons = ?, daily_challenges = ?, daily_quizzes = ?, gamification = ?, timezone = ?, updated_at = ? WHERE user_id = ?`),
		p.DisplayName, p.Bio, p.LearningPath, p.Theme, p.DailyLessons, p.DailyChallenges, p.DailyQuizzes, g, p.Timezone, now(), userID)
	return err
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

// CreateSession returns an opaque token; only its hash is stored.
func (s *Store) CreateSession(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	_, err := s.db.ExecContext(ctx, s.q(`INSERT INTO sessions (token_hash, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`),
		hashToken(token), userID, now(), time.Now().Add(ttl).Unix())
	return token, err
}

func (s *Store) UserBySession(ctx context.Context, token string) (User, error) {
	var id string
	var exp int64
	err := s.db.QueryRowContext(ctx, s.q(`SELECT user_id, expires_at FROM sessions WHERE token_hash = ?`), hashToken(token)).Scan(&id, &exp)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if exp < now() {
		_ = s.DeleteSession(ctx, token)
		return User{}, ErrNotFound
	}
	return s.UserByID(ctx, id)
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, s.q(`DELETE FROM sessions WHERE token_hash = ?`), hashToken(token))
	return err
}

func (s *Store) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, s.q(`DELETE FROM sessions WHERE user_id = ?`), userID)
	return err
}

func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, s.q(`DELETE FROM users WHERE id = ?`), userID)
	return err
}

func (s *Store) UpdatePassword(ctx context.Context, userID, hash string) error {
	_, err := s.db.ExecContext(ctx, s.q(`UPDATE users SET password_hash = ? WHERE id = ?`), hash, userID)
	return err
}
