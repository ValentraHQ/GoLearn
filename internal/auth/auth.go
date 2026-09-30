// Package auth implements local email/password accounts with server-side
// sessions. Authentication is behind the Authenticator interface so an OIDC
// provider (Keycloak, Google, ...) can be added by implementing that interface
// and chaining it in front of SessionAuthenticator.
package auth

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/valentrahq/golearn/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNoSession          = errors.New("no valid session")
)

// Authenticator resolves the current user from a request.
type Authenticator interface {
	Authenticate(r *http.Request) (store.User, error)
}

// Chain tries each authenticator in order.
type Chain []Authenticator

func (c Chain) Authenticate(r *http.Request) (store.User, error) {
	for _, a := range c {
		if u, err := a.Authenticate(r); err == nil {
			return u, nil
		}
	}
	return store.User{}, ErrNoSession
}

type Service struct {
	Store      *store.Store
	CookieName string
	Secure     bool
	TTL        time.Duration
	Cost       int
	dummyHash  []byte
}

func NewService(st *store.Store, cookieName string, secure bool, ttl time.Duration, cost int) *Service {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	h, _ := bcrypt.GenerateFromPassword([]byte("golearn-dummy-password"), cost)
	return &Service{Store: st, CookieName: cookieName, Secure: secure, TTL: ttl, Cost: cost, dummyHash: h}
}

func ValidateEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 254 || !strings.Contains(email[strings.Index(email, "@"):], ".") {
		return "", errors.New("Enter a valid email address.")
	}
	return email, nil
}

func ValidatePassword(pw string) error {
	n := utf8.RuneCountInString(pw)
	switch {
	case n < 10:
		return errors.New("Password must be at least 10 characters.")
	case len(pw) > 72:
		return errors.New("Password must be at most 72 bytes.") // bcrypt limit
	}
	return nil
}

func (s *Service) Register(ctx context.Context, email, password, name string) (store.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.Cost)
	if err != nil {
		return store.User{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}
	return s.Store.CreateUser(ctx, email, string(hash), name)
}

// Login verifies credentials in near-constant time regardless of whether the account exists.
func (s *Service) Login(ctx context.Context, email, password string) (store.User, error) {
	u, err := s.Store.UserByEmail(ctx, email)
	if errors.Is(err, store.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return store.User{}, ErrInvalidCredentials
	}
	if err != nil {
		return store.User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return store.User{}, ErrInvalidCredentials
	}
	return u, nil
}

func (s *Service) CheckPassword(u store.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (s *Service) HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), s.Cost)
	return string(h), err
}

func (s *Service) StartSession(w http.ResponseWriter, r *http.Request, u store.User) error {
	token, err := s.Store.CreateSession(r.Context(), u.ID, s.TTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: s.CookieName, Value: token, Path: "/", HttpOnly: true, Secure: s.Secure,
		SameSite: http.SameSiteLaxMode, MaxAge: int(s.TTL.Seconds()),
	})
	return nil
}

func (s *Service) EndSession(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.CookieName); err == nil {
		_ = s.Store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: s.CookieName, Value: "", Path: "/", HttpOnly: true, Secure: s.Secure,
		SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

// SessionAuthenticator authenticates via the session cookie.
type SessionAuthenticator struct{ *Service }

func (a SessionAuthenticator) Authenticate(r *http.Request) (store.User, error) {
	c, err := r.Cookie(a.CookieName)
	if err != nil || c.Value == "" {
		return store.User{}, ErrNoSession
	}
	u, err := a.Store.UserBySession(r.Context(), c.Value)
	if err != nil {
		return store.User{}, ErrNoSession
	}
	return u, nil
}
