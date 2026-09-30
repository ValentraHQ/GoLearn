# HTTP and REST Challenges

## middleware-chain
title: Middleware Chain
difficulty: intermediate
module: networking
lesson: networking/middleware
skill: http

### Problem
Write `Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler` that wraps `h` with the middlewares so that the **first** middleware in the list is the **outermost** (runs first on the way in). With no middlewares it returns `h` behaving unchanged.

### Constraints
- First listed = outermost
- A middleware may short-circuit by not calling `next`

### Starter
```go
package main

import "net/http"

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	return h
}

func main() {}
```

### Tests
```go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mw(name string, trace *[]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*trace = append(*trace, name+":in")
			next.ServeHTTP(w, r)
			*trace = append(*trace, name+":out")
		})
	}
}

func TestChainOrder(t *testing.T) {
	var trace []string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace = append(trace, "handler")
	}), mw("a", &trace), mw("b", &trace), mw("c", &trace))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	got := strings.Join(trace, " ")
	want := "a:in b:in c:in handler c:out b:out a:out"
	if got != want {
		t.Errorf("trace = %q\nwant    %q", got, want)
	}
}

func TestChainEmptyAndShortCircuit(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	rec := httptest.NewRecorder()
	Chain(base).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 204 {
		t.Errorf("empty chain changed behaviour: %d", rec.Code)
	}
	deny := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "no", http.StatusForbidden)
		})
	}
	rec = httptest.NewRecorder()
	Chain(base, deny).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 403 {
		t.Errorf("short-circuit failed: %d", rec.Code)
	}
}
```

### Hints
- Wrap from the last middleware to the first: `for i := len(mws)-1; i >= 0; i-- { h = mws[i](h) }`

### Solution
```go
package main

import "net/http"

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func main() {}
```

### Explanation
Wrapping in reverse means the first middleware ends up outermost, so it sees the request first and the response last.

## json-user-handler
title: JSON Create Endpoint
difficulty: intermediate
module: networking
lesson: networking/http-servers
skill: http

### Problem
Write `CreateUser(w http.ResponseWriter, r *http.Request)` for `POST /users` with a JSON body `{"name":"...","email":"..."}`.

- Wrong method → `405`
- Invalid JSON or unknown fields → `400` with body `{"error":"invalid json"}`
- Empty `name` or an `email` without `@` → `422` with body `{"error":"<message>"}` where the message is `name is required` or `email is invalid` (name is checked first)
- Success → `201`, `Content-Type: application/json`, body `{"id":1,"name":"...","email":"..."}`

Ids start at 1 and increase by one per created user (use a package-level counter guarded by a mutex or atomic).

### Constraints
- Reject unknown JSON fields
- Set the JSON content type on every JSON response

### Starter
```go
package main

import "net/http"

func CreateUser(w http.ResponseWriter, r *http.Request) {
}

func main() {}
```

### Tests
```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func post(body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	CreateUser(rec, httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body)))
	return rec
}

func TestCreateOK(t *testing.T) {
	rec := post(`{"name":"Ada","email":"ada@x.io"}`)
	if rec.Code != 201 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content type %q", ct)
	}
	var got struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got.Name != "Ada" || got.Email != "ada@x.io" || got.ID < 1 {
		t.Errorf("body %s (%v)", rec.Body, err)
	}
	second := post(`{"name":"Bob","email":"bob@x.io"}`)
	var got2 struct{ ID int }
	json.Unmarshal(second.Body.Bytes(), &got2)
	if got2.ID != got.ID+1 {
		t.Errorf("ids: %d then %d", got.ID, got2.ID)
	}
}

func TestCreateErrors(t *testing.T) {
	cases := []struct {
		body string
		code int
		msg  string
	}{
		{`not json`, 400, "invalid json"},
		{`{"name":"A","email":"a@b","extra":1}`, 400, "invalid json"},
		{`{"name":"","email":"a@b.c"}`, 422, "name is required"},
		{`{"name":"A","email":"nope"}`, 422, "email is invalid"},
		{`{"name":"","email":"nope"}`, 422, "name is required"},
	}
	for _, tc := range cases {
		rec := post(tc.body)
		var e struct{ Error string }
		json.Unmarshal(rec.Body.Bytes(), &e)
		if rec.Code != tc.code || e.Error != tc.msg {
			t.Errorf("%s -> %d %q, want %d %q", tc.body, rec.Code, e.Error, tc.code, tc.msg)
		}
	}
}

func TestWrongMethod(t *testing.T) {
	rec := httptest.NewRecorder()
	CreateUser(rec, httptest.NewRequest(http.MethodGet, "/users", nil))
	if rec.Code != 405 {
		t.Errorf("status %d", rec.Code)
	}
}
```

### Hints
- `dec := json.NewDecoder(r.Body); dec.DisallowUnknownFields()`
- A small `writeJSON(w, status, v)` helper keeps handlers tidy
- Use `sync/atomic` for the id counter

### Solution
```go
package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
)

var nextID atomic.Int64

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	switch {
	case in.Name == "":
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "name is required"})
	case !strings.Contains(in.Email, "@"):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "email is invalid"})
	default:
		writeJSON(w, http.StatusCreated, map[string]any{"id": nextID.Add(1), "name": in.Name, "email": in.Email})
	}
}

func main() {}
```

### Explanation
Distinguish malformed input (400) from well-formed but invalid input (422), always set the content type, and keep the id counter atomic because handlers run concurrently.

## fetch-with-retry
title: Fetch With Retry and Backoff
difficulty: advanced
module: networking
lesson: networking/http-clients
skill: http

### Problem
Write `FetchWithRetry(ctx context.Context, client *http.Client, url string, attempts int, backoff time.Duration) ([]byte, error)`.

- Return the body of the first `200` response.
- Retry when the status is 500, 502, 503 or 504, or on transport errors. Wait `backoff` before the first retry and double it each time.
- Do **not** retry other statuses (e.g. 404): return an error containing the status code.
- Stop immediately with `ctx.Err()` if the context is cancelled while waiting.
- After the final attempt fails, return an error that mentions the number of attempts.

### Constraints
- Close every response body
- Respect context cancellation during backoff sleeps

### Starter
```go
package main

import (
	"context"
	"net/http"
	"time"
)

func FetchWithRetry(ctx context.Context, client *http.Client, url string, attempts int, backoff time.Duration) ([]byte, error) {
	return nil, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetriesUntilSuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("hello"))
	}))
	defer srv.Close()
	body, err := FetchWithRetry(context.Background(), srv.Client(), srv.URL, 5, time.Millisecond)
	if err != nil || string(body) != "hello" || calls.Load() != 3 {
		t.Errorf("body=%q err=%v calls=%d", body, err, calls.Load())
	}
}

func TestNoRetryOn404(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.NotFound(w, r)
	}))
	defer srv.Close()
	_, err := FetchWithRetry(context.Background(), srv.Client(), srv.URL, 5, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "404") || calls.Load() != 1 {
		t.Errorf("err=%v calls=%d", err, calls.Load())
	}
}

func TestGivesUp(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(500)
	}))
	defer srv.Close()
	_, err := FetchWithRetry(context.Background(), srv.Client(), srv.URL, 3, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "3") || calls.Load() != 3 {
		t.Errorf("err=%v calls=%d", err, calls.Load())
	}
}

func TestBackoffDoublesAndContextCancels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := FetchWithRetry(ctx, srv.Client(), srv.URL, 10, 30*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("want DeadlineExceeded, got %v", err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Error("did not stop promptly when the context expired")
	}
}
```

### Hints
- Build requests with `http.NewRequestWithContext`
- Sleep with `select { case <-time.After(wait): case <-ctx.Done(): return nil, ctx.Err() }`
- Drain and close the body before retrying

### Solution
```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func retryable(status int) bool {
	return status == 500 || status == 502 || status == 503 || status == 504
}

func FetchWithRetry(ctx context.Context, client *http.Client, url string, attempts int, backoff time.Duration) ([]byte, error) {
	var lastErr error
	wait := backoff
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			wait *= 2
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusOK && readErr == nil:
			return body, nil
		case resp.StatusCode == http.StatusOK:
			lastErr = readErr
		case retryable(resp.StatusCode):
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		default:
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", attempts, lastErr)
}

func main() {}
```

### Explanation
The retry loop distinguishes retryable failures (5xx, transport errors) from permanent ones (4xx), backs off exponentially, and uses `select` so waiting never outlives the context.
