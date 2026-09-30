// Package httpx holds small HTTP helpers: JSON envelopes, errors, middleware,
// rate limiting and Prometheus-style metrics, built on the standard library.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// APIError is returned by handlers to produce a consistent error body.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string { return e.Message }

func Err(status int, code, msg string) *APIError { return &APIError{status, code, msg} }

var (
	ErrUnauthorized = Err(http.StatusUnauthorized, "unauthorized", "Sign in to continue.")
	ErrNotFound     = Err(http.StatusNotFound, "not_found", "Not found.")
)

func BadRequest(msg string) *APIError { return Err(http.StatusBadRequest, "bad_request", msg) }

type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func OK(w http.ResponseWriter, data any) { JSON(w, http.StatusOK, map[string]any{"data": data}) }

func Page(w http.ResponseWriter, data any, meta Meta) {
	JSON(w, http.StatusOK, map[string]any{"data": data, "meta": meta})
}

func WriteError(w http.ResponseWriter, e *APIError) {
	JSON(w, e.Status, map[string]any{"error": map[string]any{"code": e.Code, "message": e.Message}})
}

// Handler is an http.Handler that can return an error.
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		var ae *APIError
		if errors.As(err, &ae) {
			WriteError(w, ae)
			return
		}
		slog.Error("handler error", "err", err, "path", r.URL.Path, "request_id", RequestID(r.Context()))
		WriteError(w, Err(http.StatusInternalServerError, "internal", "Something went wrong on our side."))
	}
}

// Decode reads a JSON body with a size limit and rejects unknown fields.
func Decode(w http.ResponseWriter, r *http.Request, dst any, limit int64) error {
	if limit == 0 {
		limit = 1 << 20
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		switch {
		case errors.As(err, &mbe):
			return Err(http.StatusRequestEntityTooLarge, "too_large", "Request body is too large.")
		case errors.Is(err, io.EOF):
			return BadRequest("Request body is empty.")
		default:
			return BadRequest("Invalid JSON: " + err.Error())
		}
	}
	return nil
}

type ctxKey int

const reqIDKey ctxKey = 1

func RequestID(ctx context.Context) string {
	s, _ := ctx.Value(reqIDKey).(string)
	return s
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusWriter) WriteHeader(c int) {
	s.status = c
	s.ResponseWriter.WriteHeader(c)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = 200
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Middleware applies request IDs, panic recovery, security headers, logging and metrics.
func Middleware(next http.Handler, m *Metrics, allowOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Responses can be per-user; never let a shared cache keep them.
			h.Set("Cache-Control", "no-store")
		}
		if allowOrigin != "" && r.Header.Get("Origin") == allowOrigin {
			h.Set("Access-Control-Allow-Origin", allowOrigin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Headers", "Content-Type, X-GoLearn-CSRF")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		sw := &statusWriter{ResponseWriter: w}
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		r = r.WithContext(ctx)
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "panic", fmt.Sprint(rec), "stack", string(debug.Stack()), "request_id", id)
				if sw.status == 0 {
					WriteError(sw, Err(http.StatusInternalServerError, "internal", "Something went wrong on our side."))
				}
			}
			route := r.Pattern
			if route == "" {
				route = "unmatched"
			}
			d := time.Since(start)
			if m != nil {
				m.Observe(r.Method, route, sw.status, d)
			}
			if !strings.HasPrefix(r.URL.Path, "/assets/") && r.URL.Path != "/healthz" {
				slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", sw.status, "ms", d.Milliseconds(), "request_id", id)
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// RequireCSRFHeader rejects state-changing API calls that lack the custom
// header. Browsers cannot attach it cross-origin without a CORS preflight, so
// together with SameSite cookies this blocks CSRF.
func RequireCSRFHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if strings.HasPrefix(r.URL.Path, "/api/") && r.Header.Get("X-GoLearn-CSRF") != "1" {
				WriteError(w, Err(http.StatusForbidden, "csrf", "Missing CSRF header."))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP returns the caller's IP, honouring X-Forwarded-For only when trusted.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			return strings.TrimSpace(first)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
