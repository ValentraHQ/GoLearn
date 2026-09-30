package api

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/valentrahq/golearn/internal/auth"
	"github.com/valentrahq/golearn/internal/httpx"
	"github.com/valentrahq/golearn/internal/store"
)

type userDTO struct {
	ID      string        `json:"id"`
	Email   string        `json:"email"`
	Role    string        `json:"role"`
	Profile store.Profile `json:"profile"`
}

func (s *Server) userDTO(r *http.Request, u store.User) userDTO {
	p, _ := s.profileAndLoc(r.Context(), u)
	return userDTO{ID: u.ID, Email: u.Email, Role: u.Role, Profile: p}
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) error {
	httpx.OK(w, map[string]any{
		"auth":   map[string]any{"local": true, "oidc": false},
		"runner": s.runner.Status(r.Context()),
		"limits": map[string]any{"maxCodeBytes": maxCodeBytes},
		"content": map[string]any{
			"modules": len(s.cat.Modules), "lessons": len(s.cat.PublishedLessons()),
			"challenges": len(s.cat.Challenges), "projects": len(s.cat.Projects), "glossary": len(s.cat.Glossary),
		},
	})
	return nil
}

func (s *Server) limitAuth(r *http.Request) error {
	if ok, _ := s.authRate.Allow(httpx.ClientIP(r, s.cfg.TrustProxy)); !ok {
		return httpx.Err(http.StatusTooManyRequests, "rate_limited", "Too many attempts. Wait a minute and try again.")
	}
	return nil
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) error {
	if err := s.limitAuth(r); err != nil {
		return err
	}
	var in struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
		Timezone    string `json:"timezone"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	email, err := auth.ValidateEmail(in.Email)
	if err != nil {
		return httpx.BadRequest(err.Error())
	}
	if err := auth.ValidatePassword(in.Password); err != nil {
		return httpx.BadRequest(err.Error())
	}
	name := strings.TrimSpace(in.DisplayName)
	if utf8.RuneCountInString(name) > 60 {
		return httpx.BadRequest("Display name must be at most 60 characters.")
	}
	u, err := s.auth.Register(r.Context(), email, in.Password, name)
	if err == store.ErrEmailTaken {
		return httpx.Err(http.StatusConflict, "email_taken", "An account with this email already exists.")
	}
	if err != nil {
		return err
	}
	if _, err := time.LoadLocation(in.Timezone); err == nil && in.Timezone != "" && in.Timezone != "Local" {
		p, _ := s.store.Profile(r.Context(), u.ID)
		p.Timezone = in.Timezone
		_ = s.store.UpdateProfile(r.Context(), u.ID, p)
	}
	if err := s.auth.StartSession(w, r, u); err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"data": s.userDTO(r, u)})
	return nil
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) error {
	if err := s.limitAuth(r); err != nil {
		return err
	}
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	u, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err == auth.ErrInvalidCredentials {
		return httpx.Err(http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
	}
	if err != nil {
		return err
	}
	if err := s.auth.StartSession(w, r, u); err != nil {
		return err
	}
	httpx.OK(w, s.userDTO(r, u))
	return nil
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) error {
	s.auth.EndSession(w, r)
	httpx.OK(w, map[string]bool{"ok": true})
	return nil
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, u *store.User) error {
	if u == nil {
		httpx.OK(w, nil)
		return nil
	}
	httpx.OK(w, s.userDTO(r, *u))
	return nil
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request, u store.User) error {
	if err := s.limitAuth(r); err != nil {
		return err
	}
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	if !s.auth.CheckPassword(u, in.Current) {
		return httpx.Err(http.StatusForbidden, "invalid_credentials", "Current password is incorrect.")
	}
	if err := auth.ValidatePassword(in.New); err != nil {
		return httpx.BadRequest(err.Error())
	}
	hash, err := s.auth.HashPassword(in.New)
	if err != nil {
		return err
	}
	if err := s.store.UpdatePassword(r.Context(), u.ID, hash); err != nil {
		return err
	}
	// Sign out everywhere, then keep this browser signed in.
	if err := s.store.DeleteUserSessions(r.Context(), u.ID); err != nil {
		return err
	}
	if err := s.auth.StartSession(w, r, u); err != nil {
		return err
	}
	httpx.OK(w, map[string]bool{"ok": true})
	return nil
}

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request, u store.User) error {
	if err := s.limitAuth(r); err != nil {
		return err
	}
	var in struct {
		Password string `json:"password"`
	}
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	if !s.auth.CheckPassword(u, in.Password) {
		return httpx.Err(http.StatusForbidden, "invalid_credentials", "Password is incorrect.")
	}
	if err := s.store.DeleteUser(r.Context(), u.ID); err != nil {
		return err
	}
	s.auth.EndSession(w, r)
	httpx.OK(w, map[string]bool{"ok": true})
	return nil
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request, u store.User) error {
	httpx.OK(w, s.userDTO(r, u).Profile)
	return nil
}

func (s *Server) putProfile(w http.ResponseWriter, r *http.Request, u store.User) error {
	var in store.Profile
	if err := httpx.Decode(w, r, &in, 0); err != nil {
		return err
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Bio = strings.TrimSpace(in.Bio)
	switch {
	case in.DisplayName == "" || utf8.RuneCountInString(in.DisplayName) > 60:
		return httpx.BadRequest("Display name is required (max 60 characters).")
	case utf8.RuneCountInString(in.Bio) > 300:
		return httpx.BadRequest("Bio must be at most 300 characters.")
	case in.LearningPath != "beginner" && in.LearningPath != "pro":
		return httpx.BadRequest("Learning path must be beginner or pro.")
	case in.Theme != "system" && in.Theme != "light" && in.Theme != "dark":
		return httpx.BadRequest("Theme must be system, light or dark.")
	case !inRange(in.DailyLessons, 0, 20) || !inRange(in.DailyChallenges, 0, 20) || !inRange(in.DailyQuizzes, 0, 20):
		return httpx.BadRequest("Daily goals must be between 0 and 20.")
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil || in.Timezone == "" || in.Timezone == "Local" {
		return httpx.BadRequest("Unknown timezone.")
	}
	if err := s.store.UpdateProfile(r.Context(), u.ID, in); err != nil {
		return err
	}
	httpx.OK(w, s.userDTO(r, u).Profile)
	return nil
}

func inRange(v, lo, hi int) bool { return v >= lo && v <= hi }
