// Package api wires the REST handlers.
package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/valentrahq/golearn/internal/auth"
	"github.com/valentrahq/golearn/internal/config"
	"github.com/valentrahq/golearn/internal/content"
	"github.com/valentrahq/golearn/internal/httpx"
	"github.com/valentrahq/golearn/internal/progress"
	"github.com/valentrahq/golearn/internal/runner"
	"github.com/valentrahq/golearn/internal/search"
	"github.com/valentrahq/golearn/internal/store"
)

type Server struct {
	cfg      config.Config
	store    *store.Store
	cat      *content.Catalog
	index    *search.Index
	auth     *auth.Service
	authn    auth.Authenticator
	runner   runner.Runner
	metrics  *httpx.Metrics
	authRate *httpx.Limiter
	runRate  *httpx.Limiter
	pingRate *httpx.Limiter
}

func New(cfg config.Config, st *store.Store, cat *content.Catalog, rn runner.Runner) *Server {
	svc := auth.NewService(st, cfg.CookieName, cfg.SecureCookie, cfg.SessionTTL, cfg.BcryptCost)
	return &Server{
		cfg: cfg, store: st, cat: cat, index: search.Build(cat), auth: svc,
		authn:    auth.Chain{auth.SessionAuthenticator{Service: svc}},
		runner:   rn,
		metrics:  httpx.NewMetrics(),
		authRate: httpx.NewLimiter(20, 10),
		runRate:  httpx.NewLimiter(max(cfg.RunsPerMinute, 1), max(cfg.RunsPerMinute/3, 3)),
		pingRate: httpx.NewLimiter(6, 3),
	}
}

type ctxKey int

const userKey ctxKey = 1

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	h := func(f httpx.Handler) http.Handler { return f }
	authed := func(f func(w http.ResponseWriter, r *http.Request, u store.User) error) http.Handler {
		return httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
			u, err := s.authn.Authenticate(r)
			if err != nil {
				return authError(err)
			}
			return f(w, r, u)
		})
	}
	// optional serves public content, personalised when a valid session exists.
	// If the session backend fails the request degrades to anonymous, except
	// where strict is set (e.g. /api/auth/me, which must not claim "signed out").
	optionalWith := func(strict bool, f func(w http.ResponseWriter, r *http.Request, u *store.User) error) http.Handler {
		return httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
			var up *store.User
			u, err := s.authn.Authenticate(r)
			switch {
			case err == nil:
				up = &u
			case strict && !errors.Is(err, auth.ErrNoSession):
				return authError(err)
			}
			return f(w, r, up)
		})
	}
	optional := func(f func(w http.ResponseWriter, r *http.Request, u *store.User) error) http.Handler {
		return optionalWith(false, f)
	}

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := s.store.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ready"))
	})
	mux.Handle("GET /metrics", s.metrics.Handler())

	// Config & auth.
	mux.Handle("GET /api/config", h(s.getConfig))
	mux.Handle("POST /api/auth/register", h(s.register))
	mux.Handle("POST /api/auth/login", h(s.login))
	mux.Handle("POST /api/auth/logout", h(s.logout))
	mux.Handle("GET /api/auth/me", optionalWith(true, s.me))
	mux.Handle("POST /api/auth/password", authed(s.changePassword))
	mux.Handle("DELETE /api/account", authed(s.deleteAccount))
	mux.Handle("GET /api/profile", authed(s.getProfile))
	mux.Handle("PUT /api/profile", authed(s.putProfile))

	// Curriculum.
	mux.Handle("GET /api/modules", optional(s.listModules))
	mux.Handle("GET /api/modules/{id}", optional(s.getModule))
	mux.Handle("GET /api/lessons/{module}/{slug}", optional(s.getLesson))
	mux.Handle("GET /api/lessons/{module}/{slug}/solution", h(s.lessonSolution))
	mux.Handle("GET /api/challenges", optional(s.listChallenges))
	mux.Handle("GET /api/challenges/{id}", optional(s.getChallenge))
	mux.Handle("GET /api/projects", optional(s.listProjects))
	mux.Handle("GET /api/projects/{id}", optional(s.getProject))
	mux.Handle("GET /api/glossary", h(s.listGlossary))
	mux.Handle("GET /api/search", h(s.search))

	// Learner state.
	mux.Handle("GET /api/dashboard", authed(s.dashboard))
	mux.Handle("GET /api/progress", authed(s.getProgress))
	mux.Handle("POST /api/progress/lessons/{module}/{slug}", authed(s.setLessonProgress))
	mux.Handle("POST /api/lessons/{module}/{slug}/exercise", authed(s.runExercise))
	mux.Handle("POST /api/lessons/{module}/{slug}/quiz", authed(s.submitQuiz))
	mux.Handle("POST /api/challenges/{id}/submit", authed(s.submitChallenge))
	mux.Handle("GET /api/challenges/{id}/solution", authed(s.challengeSolution))
	mux.Handle("PUT /api/projects/{id}/tasks/{task}", authed(s.setProjectTask))
	mux.Handle("GET /api/skills", authed(s.getSkills))
	mux.Handle("GET /api/achievements", authed(s.getAchievements))
	mux.Handle("POST /api/sessions/ping", authed(s.ping))

	// Code execution.
	mux.Handle("GET /api/runner/status", h(s.runnerStatus))
	mux.Handle("POST /api/run", authed(s.playgroundRun))

	mux.Handle("/api/", h(func(w http.ResponseWriter, r *http.Request) error { return httpx.ErrNotFound }))
	if s.cfg.StaticDir != "" {
		mux.Handle("/", s.spa(s.cfg.StaticDir))
	}
	return httpx.Middleware(httpx.RequireCSRFHeader(mux), s.metrics, s.cfg.AllowOrigin)
}

// spa serves the built frontend with index.html fallback for client routes.
func (s *Server) spa(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
			http.Error(w, "frontend not built (run `make web`)", http.StatusNotFound)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

// --- shared helpers ---

func (s *Server) profileAndLoc(ctx context.Context, u store.User) (store.Profile, *time.Location) {
	p, err := s.store.Profile(ctx, u.ID)
	if err != nil {
		p = store.Profile{LearningPath: "beginner", DailyLessons: 1, DailyChallenges: 2, DailyQuizzes: 1, Timezone: "UTC"}
	}
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return p, loc
}

// loadProgress loads state, computes the summary, and persists newly earned achievements.
func (s *Server) loadProgress(ctx context.Context, u store.User) (progress.Summary, store.Profile, progress.Input, error) {
	p, loc := s.profileAndLoc(ctx, u)
	in := progress.Input{Path: p.LearningPath, Goals: progress.Goals{Lessons: p.DailyLessons, Challenges: p.DailyChallenges, Quizzes: p.DailyQuizzes}}
	var err error
	if in.Lessons, err = s.store.LessonProgress(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Challenges, err = s.store.ChallengeStatuses(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Quizzes, err = s.store.QuizAttempts(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Projects, err = s.store.ProjectTasks(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Seconds, err = s.store.LearningSeconds(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Events, err = s.store.Events(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	if in.Unlocked, err = s.store.UnlockedAchievements(ctx, u.ID); err != nil {
		return progress.Summary{}, p, in, err
	}
	sum := progress.Compute(s.cat, in, time.Now(), loc)
	for _, id := range sum.NewlyUnlocked {
		if err := s.store.UnlockAchievement(ctx, u.ID, id); err != nil {
			return sum, p, in, err
		}
	}
	return sum, p, in, nil
}

func pathParam(r *http.Request, name string) string { return r.PathValue(name) }

func lessonID(r *http.Request) string { return r.PathValue("module") + "/" + r.PathValue("slug") }

func (s *Server) lessonOr404(r *http.Request) (*content.Lesson, error) {
	l := s.cat.Lesson(lessonID(r))
	if l == nil || l.Status != content.StatusPublished {
		return nil, httpx.ErrNotFound
	}
	return l, nil
}

func isNotFound(err error) bool { return errors.Is(err, store.ErrNotFound) }

// authError maps an authentication failure to an API error: a missing or
// expired session is a 401, anything else (database down) is a 503 so clients
// keep the user signed in and retry.
func authError(err error) error {
	if errors.Is(err, auth.ErrNoSession) {
		return httpx.ErrUnauthorized
	}
	slog.Error("authentication backend failure", "err", err)
	return httpx.Err(http.StatusServiceUnavailable, "unavailable", "The service is temporarily unavailable. Please try again.")
}
