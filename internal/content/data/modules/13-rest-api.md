# REST API Development
id: rest-api
number: 13
track: advanced
paths: beginner, pro
skill: rest
requires: networking
project: production-rest-api
summary: Design and build a layered, validated, authenticated and rate-limited REST API.

## REST Architecture
slug: rest-architecture
minutes: 8
objectives: Explain REST's constraints in practical terms; Model resources and URLs as nouns; Structure a service as handler → service → repository
takeaways: REST models resources (nouns) addressed by URLs and manipulated with standard HTTP methods; Stateless requests carry everything the server needs; Layer the code: handler (HTTP) → service (rules) → repository (storage)

### Concept
REST is a style, not a spec. In practice: **resources** (`/users/42`), **representations** (JSON), **standard methods** (GET, POST, PUT, PATCH, DELETE) and **stateless** requests — the server keeps no session between calls.

```
Client → Router → Middleware → Handler → Service → Repository → Database
```
- **Handler** — translates HTTP ⇄ Go values; no business rules.
- **Service** — business logic, unaware of HTTP.
- **Repository** — storage behind an interface.

Each layer depends only on the one below, through interfaces, so you can test the service with a fake repository and swap the database freely.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID   int
	Name string
}

type Repository interface {
	Find(id int) (User, error)
}

type memRepo map[int]User

func (m memRepo) Find(id int) (User, error) {
	if u, ok := m[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

type Service struct{ repo Repository }

func (s Service) Greeting(id int) (string, error) {
	u, err := s.repo.Find(id)
	if err != nil {
		return "", fmt.Errorf("greeting: %w", err)
	}
	return "hello " + u.Name, nil
}

func main() {
	svc := Service{memRepo{1: {1, "Ada"}}}
	fmt.Println(svc.Greeting(1))
	_, err := svc.Greeting(2)
	fmt.Println(errors.Is(err, ErrNotFound))
}
```

### Exercise
Implement `Repository.Save` on `memRepo` (store the user under its ID) and `Service.Rename(id int, name string) error` that finds the user, changes the name and saves it. Renaming user 1 to `Grace` prints `Grace`.
```text expect
Grace
```
```go solution
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID   int
	Name string
}

type Repository interface {
	Find(id int) (User, error)
	Save(u User) error
}

type memRepo map[int]User

func (m memRepo) Find(id int) (User, error) {
	if u, ok := m[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

// BEGIN
func (m memRepo) Save(u User) error {
	m[u.ID] = u
	return nil
}

// END

type Service struct{ repo Repository }

// BEGIN
func (s Service) Rename(id int, name string) error {
	u, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	u.Name = name
	return s.repo.Save(u)
}

// END

func main() {
	repo := memRepo{1: {1, "Ada"}}
	svc := Service{repo}
	_ = svc.Rename(1, "Grace")
	fmt.Println(repo[1].Name)
}
```

### Check
Q: What does "stateless" mean for a REST API?
T: mcq
- [ ] The server has no database
- [x] Each request contains all information needed to process it; no server-side session is required
- [ ] Responses are never cached
- [ ] Only GET is allowed
E: The server doesn't remember the client between requests, which makes scaling horizontally easy.

Q: Which layer should contain business rules?
T: mcq
- [ ] The HTTP handler
- [x] The service layer
- [ ] The repository
- [ ] The router
E: Handlers translate HTTP; repositories store data; rules live in the service.

## API Design
slug: api-design
status: planned

## HTTP Methods
slug: http-methods
minutes: 6
objectives: Match methods to operations; Distinguish safe and idempotent methods; Choose PUT versus PATCH
takeaways: GET reads, POST creates/acts, PUT replaces, PATCH partially updates, DELETE removes; GET is safe (no side effects); GET, PUT and DELETE are idempotent; POST is neither; Idempotency determines what a client may safely retry

### Concept
| Method | Meaning | Safe | Idempotent |
| --- | --- | --- | --- |
| GET | read a resource | yes | yes |
| POST | create / perform an action | no | no |
| PUT | replace a resource | no | yes |
| PATCH | partial update | no | usually no |
| DELETE | remove | no | yes |

**Safe** = no visible side effects. **Idempotent** = repeating the request has the same effect as sending it once. Clients and proxies may retry idempotent requests automatically; for `POST` use an **Idempotency-Key** header so retries don't double-charge.

### Example
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
		fmt.Println(m)
	}
}
```

### Exercise
Write `idempotent(method string) bool` returning true for GET, HEAD, PUT, DELETE and OPTIONS.
```text expect
true false true false
```
```go solution
package main

import (
	"fmt"
	"net/http"
)

// BEGIN
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

// END

func main() {
	fmt.Println(idempotent("GET"), idempotent("POST"), idempotent("DELETE"), idempotent("PATCH"))
}
```

### Check
Q: Which method replaces a resource entirely and is idempotent?
T: short
A: PUT
E: Sending the same PUT twice leaves the resource in the same state.

Q: Why may a client safely retry a DELETE after a timeout?
T: mcq
- [ ] DELETE is safe
- [x] DELETE is idempotent: deleting twice has the same end result
- [ ] Servers ignore DELETE retries
- [ ] Retries are always safe
E: The resource ends up deleted either way (the second call may return 404, but the state is the same).

## Status Codes
slug: status-codes
minutes: 7
objectives: Choose correct status codes for common outcomes; Map application errors to codes in one place; Avoid returning 200 for failures
takeaways: 2xx success, 3xx redirect, 4xx client error, 5xx server error; 201 Created, 204 No Content, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 409 Conflict, 422 Unprocessable, 429 Too Many Requests; Translate domain errors to statuses in a single function

### Concept
| Code | Use |
| --- | --- |
| 200 OK | successful GET/PUT/PATCH with a body |
| 201 Created | successful POST that created something (+ `Location` header) |
| 204 No Content | success with no body (DELETE) |
| 400 Bad Request | malformed request / invalid JSON |
| 401 Unauthorized | missing or invalid credentials (really "unauthenticated") |
| 403 Forbidden | authenticated but not allowed |
| 404 Not Found | resource doesn't exist |
| 409 Conflict | state conflict (duplicate, version mismatch) |
| 422 Unprocessable Entity | well-formed but semantically invalid |
| 429 Too Many Requests | rate limited (+ `Retry-After`) |
| 500 / 503 | server bug / temporarily unavailable |

Never hide failures inside a `200` with `{"error": ...}`. Keep one `statusFor(err)` mapping so handlers stay consistent.

### Example
```go
package main

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrNotFound = errors.New("not found")

func statusFor(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func main() {
	fmt.Println(statusFor(nil), statusFor(fmt.Errorf("get: %w", ErrNotFound)), statusFor(errors.New("boom")))
}
```

### Exercise
Extend `statusFor` with `ErrInvalid` → 422 and `ErrConflict` → 409 (checked with `errors.Is`).
```text expect
422 409 500
```
```go solution
package main

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
	ErrConflict = errors.New("conflict")
)

// BEGIN
func statusFor(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalid):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// END

func main() {
	fmt.Println(statusFor(ErrInvalid), statusFor(fmt.Errorf("x: %w", ErrConflict)), statusFor(errors.New("boom")))
}
```

### Check
Q: Which status means "authenticated, but not allowed to do this"?
T: mcq
- [ ] 401
- [x] 403
- [ ] 404
- [ ] 400
E: 401 = "who are you?"; 403 = "I know who you are, and no".

Q: Returning `200 OK` with `{"error":"not found"}` is good API design.
T: tf
A: false
E: Clients, caches and monitoring rely on status codes; use 404 (or the right code) for failures.

## Routing
slug: routing
status: planned

## Handlers
slug: handlers
minutes: 8
objectives: Write handlers that return errors and centralise error → response conversion; Decode requests and encode responses consistently; Keep handlers thin
takeaways: An adapter type lets handlers return error and one place writes error responses; Handlers decode input, call the service, encode output — nothing else; Use one JSON error shape for the whole API

### Concept
The standard `http.HandlerFunc` can't return an error, which pushes every handler to repeat error-writing code. Add a small adapter:

```go norun
type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		writeError(w, err) // maps err → status + JSON body once
	}
}
```
Handlers then read: decode → validate → call service → `writeJSON`. Business logic stays in the service; the handler only converts between HTTP and Go values.

### Example
```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

var ErrNotFound = errors.New("not found")

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func main() {
	h := APIHandler(func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("user 7: %w", ErrNotFound)
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/users/7", nil))
	fmt.Print(rec.Code, " ", rec.Body.String())
}
```

### Exercise
Write an `APIHandler` for `GET /ping` that writes `{"pong":true}` (via `json.NewEncoder`) with status 200 and returns nil.
```text expect
200 {"pong":true}
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BEGIN
var ping = APIHandler(func(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]bool{"pong": true})
})

// END

func main() {
	rec := httptest.NewRecorder()
	ping.ServeHTTP(rec, httptest.NewRequest("GET", "/ping", nil))
	fmt.Print(rec.Code, " ", rec.Body.String())
}
```

### Check
Q: What is the main benefit of a handler adapter that returns `error`?
T: mcq
- [ ] Faster requests
- [x] Error-to-response conversion lives in one place
- [ ] It replaces middleware
- [ ] It removes the need for JSON
E: Handlers just `return err`; consistent status codes and error bodies come from one function.

Q: Where should business rules (pricing, permissions logic) live?
T: mcq
- [ ] In handlers
- [x] In services, called by thin handlers
- [ ] In the router
- [ ] In JSON tags
E: Keeping handlers thin makes rules reusable and testable without HTTP.

## Middleware
slug: middleware
status: planned

## Request Validation
slug: request-validation
minutes: 7
objectives: Validate decoded input and report all problems at once; Return field-level errors in a consistent JSON shape; Limit body size and reject unknown fields
takeaways: Validate after decoding, before calling the service; Collect every problem in a map[field]message so clients fix them in one round-trip; Bound the request body with http.MaxBytesReader

### Concept
Never trust input. A validation pass returns **all** the problems:

```go norun
type Errors map[string]string

func (in CreateUser) Validate() Errors {
	errs := Errors{}
	if in.Name == "" { errs["name"] = "is required" }
	if !strings.Contains(in.Email, "@") { errs["email"] = "is invalid" }
	return errs
}
```
Respond `422` with `{"errors": {"name": "is required", ...}}`. Protect the decoder: `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` and `DisallowUnknownFields()`. Validation libraries exist, but a small hand-written `Validate()` method is usually clear and dependency-free.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

type CreateUser struct {
	Name  string
	Email string
	Age   int
}

func (u CreateUser) Validate() map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(u.Name) == "" {
		errs["name"] = "is required"
	}
	if !strings.Contains(u.Email, "@") {
		errs["email"] = "is invalid"
	}
	if u.Age < 0 || u.Age > 150 {
		errs["age"] = "must be between 0 and 150"
	}
	return errs
}

func main() {
	errs := CreateUser{Email: "nope", Age: 200}.Validate()
	keys := make([]string, 0, len(errs))
	for k := range errs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, errs[k])
	}
}
```

### Exercise
Write `Validate(title string, priority int) []string` returning messages `title is required` (blank title, after trimming) and `priority must be 1-5`, in that order; an empty (nil) slice when valid.
```text expect
[title is required priority must be 1-5]
[]
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func Validate(title string, priority int) []string {
	var problems []string
	if strings.TrimSpace(title) == "" {
		problems = append(problems, "title is required")
	}
	if priority < 1 || priority > 5 {
		problems = append(problems, "priority must be 1-5")
	}
	return problems
}

// END

func main() {
	fmt.Println(Validate("  ", 9))
	fmt.Println(Validate("Ship it", 3))
}
```

### Check
Q: Why report all validation errors at once?
T: mcq
- [ ] It is faster
- [x] The client can fix everything in one round-trip
- [ ] It hides implementation details
- [ ] JSON requires it
E: Returning only the first error makes users fix problems one at a time.

Q: Which HTTP status suits well-formed JSON that fails business validation?
T: short
A: 422
E: 422 Unprocessable Entity (many APIs also use 400 — be consistent).

## Response Formatting
slug: response-formatting
status: planned

## Pagination
slug: pagination
minutes: 7
challenges: paginate
objectives: Implement offset pagination with page and page_size; Return metadata (total, pages); Know when cursor pagination is better
takeaways: Always paginate list endpoints with a maximum page size; Offset pagination (page/page_size) is simple but slows and drifts on large, changing data; Cursor (keyset) pagination scales and stays stable

### Concept
Unbounded lists eventually take your service down. Accept `page` and `page_size` (default 20, max 100), and return metadata:

```json
{"data": [...], "meta": {"page": 2, "page_size": 20, "total": 143, "pages": 8}}
```
SQL: `LIMIT :size OFFSET :(page-1)*size`. Offset gets slower on deep pages and shifts if rows are inserted while paging. **Keyset/cursor** pagination (`WHERE id > :last_id ORDER BY id LIMIT :n`, returning `next_cursor`) is constant-time and stable — prefer it for feeds and big tables.

### Example
```go
package main

import "fmt"

func bounds(total, page, size int) (start, end int) {
	start = (page - 1) * size
	if start > total {
		start = total
	}
	end = start + size
	if end > total {
		end = total
	}
	return
}

func main() {
	items := []int{1, 2, 3, 4, 5, 6, 7}
	s, e := bounds(len(items), 3, 3)
	fmt.Println(items[s:e])
}
```

### Exercise
Write `pages(total, size int) int` returning the number of pages (ceiling division; 0 items → 0 pages; `size < 1` treated as 1).
```text expect
3 0 1 4
```
```go solution
package main

import "fmt"

// BEGIN
func pages(total, size int) int {
	if size < 1 {
		size = 1
	}
	return (total + size - 1) / size
}

// END

func main() {
	fmt.Println(pages(10, 4), pages(0, 10), pages(5, 5), pages(7, 2))
}
```

### Check
Q: Why cap `page_size` on the server?
T: mcq
- [ ] To make URLs shorter
- [x] To stop clients from requesting huge pages that strain the service
- [ ] Because SQL requires it
- [ ] To sort results
E: Without a maximum, one request can load the whole table.

Q: Which pagination style stays stable when rows are inserted while a client is paging?
T: mcq
- [ ] Offset/page number
- [x] Cursor (keyset)
- [ ] Random
- [ ] Neither
E: Cursors anchor to a specific row, so new rows don't shift the window.

## Filtering
slug: filtering
status: planned

## Sorting
slug: sorting
status: planned

## Authentication
slug: authentication
status: planned

## Authorization
slug: authorization
status: planned

## JWT
slug: jwt
minutes: 9
objectives: Describe the structure of a JWT; Sign and verify an HS256 token; List the mistakes that make JWT auth insecure
takeaways: A JWT is base64url(header).base64url(payload).signature; HS256 signs with an HMAC-SHA256 over header.payload using a shared secret; Always verify the signature, the algorithm and the expiry — never trust an unverified token

### Concept
A **JSON Web Token** has three base64url parts: header (`{"alg":"HS256","typ":"JWT"}`), payload (claims like `sub`, `exp`, `iss`) and signature. Signature = `HMAC-SHA256(secret, header + "." + payload)`.

The token is **signed, not encrypted** — anyone can read the claims, so never put secrets in them.

Verification checklist:
1. Recompute the signature and compare with `hmac.Equal` (constant time).
2. Pin the expected algorithm — reject `none` and unexpected `alg`.
3. Check `exp` (and `nbf`, `iss`, `aud`).
4. Use short lifetimes and rotate keys; store the secret outside source control.

In production use a maintained library (e.g. `github.com/golang-jwt/jwt/v5`); this lesson builds one by hand so you understand what it does.

### Example
```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

var enc = base64.RawURLEncoding

func sign(payload, secret string) string {
	head := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := enc.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(head + "." + body))
	return head + "." + body + "." + enc.EncodeToString(mac.Sum(nil))
}

func main() {
	tok := sign(`{"sub":"ada"}`, "s3cret")
	fmt.Println(strings.Count(tok, "."), len(strings.Split(tok, ".")[2]) > 0)
}
```

### Exercise
Write `verify(token, secret string) (payload string, ok bool)` that recomputes the HMAC and compares it in constant time; on success return the decoded payload JSON.
```text expect
{"sub":"ada"} true
 false
 false
```
```go solution
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

var enc = base64.RawURLEncoding

func mac(signingInput, secret string) []byte {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(signingInput))
	return m.Sum(nil)
}

func sign(payload, secret string) string {
	head := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := enc.EncodeToString([]byte(payload))
	return head + "." + body + "." + enc.EncodeToString(mac(head+"."+body, secret))
}

// BEGIN
func verify(token, secret string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}
	got, err := enc.DecodeString(parts[2])
	if err != nil {
		return "", false
	}
	if !hmac.Equal(got, mac(parts[0]+"."+parts[1], secret)) {
		return "", false
	}
	payload, err := enc.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	return string(payload), true
}

// END

func main() {
	tok := sign(`{"sub":"ada"}`, "s3cret")
	fmt.Println(verify(tok, "s3cret"))
	fmt.Println(verify(tok, "wrong"))
	parts := strings.Split(tok, ".")
	parts[1] = enc.EncodeToString([]byte(`{"sub":"admin"}`))
	fmt.Println(verify(strings.Join(parts, "."), "s3cret"))
}
```

### Check
Q: Is the payload of a signed JWT encrypted?
T: mcq
- [ ] Yes, with the secret
- [x] No — it is only base64url-encoded; anyone can read it
- [ ] Only for HS256
- [ ] Only the sub claim
E: JWS tokens are signed for integrity, not confidentiality.

Q: Which comparison should be used for the signature?
T: mcq
- [ ] ==
- [ ] strings.EqualFold
- [x] hmac.Equal (constant time)
- [ ] bytes.Contains
E: Constant-time comparison avoids timing side channels.

## RBAC
slug: rbac
status: planned

## API Versioning
slug: api-versioning
status: planned

## Rate Limiting
slug: rate-limiting
minutes: 8
challenges: token-bucket
objectives: Explain token-bucket limiting; Implement a limiter with an injectable clock; Return 429 with Retry-After
takeaways: Rate limiting protects your service from overload and abuse; A token bucket allows short bursts up to capacity while enforcing an average rate; Return 429 Too Many Requests with a Retry-After header

### Concept
A **token bucket** holds up to `burst` tokens and refills at `rate` tokens/second. Each request takes one token; with none left the request is rejected (`429`).

```go norun
tokens = min(burst, tokens + elapsed*rate)
if tokens >= 1 { tokens--; allow } else { reject }
```
Key per client (API key, user ID, or IP — remember proxies: only trust `X-Forwarded-For` from your own load balancer). In-process limiters don't coordinate across replicas; for a global limit use a shared store (Redis) or the gateway. `golang.org/x/time/rate` is the standard, well-tested implementation.

### Example
```go
package main

import (
	"fmt"
	"time"
)

type Bucket struct {
	tokens, burst, rate float64
	last                time.Time
}

func (b *Bucket) Allow(now time.Time) bool {
	b.tokens = min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func main() {
	t0 := time.Unix(0, 0)
	b := &Bucket{tokens: 2, burst: 2, rate: 1, last: t0}
	fmt.Println(b.Allow(t0), b.Allow(t0), b.Allow(t0), b.Allow(t0.Add(time.Second)))
}
```

### Exercise
Write `retryAfter(tokens, rate float64) int` returning the seconds (rounded up) until one token is available: 0 when `tokens >= 1`.
```text expect
0 2 1
```
```go solution
package main

import (
	"fmt"
	"math"
)

// BEGIN
func retryAfter(tokens, rate float64) int {
	if tokens >= 1 {
		return 0
	}
	return int(math.Ceil((1 - tokens) / rate))
}

// END

func main() {
	fmt.Println(retryAfter(1.5, 1), retryAfter(0, 0.5), retryAfter(0.4, 1))
}
```

### Check
Q: Which status code signals rate limiting?
T: short
A: 429
E: `429 Too Many Requests`, ideally with a `Retry-After` header.

Q: An in-process rate limiter enforces a global limit across many server replicas.
T: tf
A: false
E: Each replica counts independently; use a shared store or the gateway for a global limit.
