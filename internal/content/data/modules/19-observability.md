# Observability
id: observability
number: 19
track: devops
paths: pro
skill: observability
requires: rest-api
project: observable-go-service
summary: Metrics, logs, traces and health checks for services you can actually operate.

## Metrics
slug: metrics
minutes: 9
objectives: Distinguish counters, gauges and histograms; Expose metrics in Prometheus text format; Follow the RED method for services
takeaways: Counters only go up, gauges go up and down, histograms bucket observations; Prometheus scrapes a /metrics endpoint in a simple text format; RED: Rate, Errors, Duration for every request-handling service

### Concept
```
Go Application ──▶ Prometheus ──▶ Grafana
   /metrics        (scrapes)      (dashboards, alerts)
```
| Type | Example |
| --- | --- |
| Counter | `http_requests_total` (only increases) |
| Gauge | `in_flight_requests`, queue depth |
| Histogram | `request_duration_seconds` buckets → percentiles |

Instrument with the official client (`github.com/prometheus/client_golang`):

```go norun
var reqs = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total", Help: "Requests by method and status.",
}, []string{"method", "status"})

reqs.WithLabelValues("GET", "200").Inc()
http.Handle("/metrics", promhttp.Handler())
```
**Labels create time series** — never use unbounded values (user IDs, URLs with IDs) as labels or you'll blow up cardinality. Use the **RED** method: **R**ate, **E**rrors, **D**uration for services; **USE** (Utilisation, Saturation, Errors) for resources.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Counter struct {
	mu     sync.Mutex
	values map[string]int
}

func (c *Counter) Inc(label string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		c.values = map[string]int{}
	}
	c.values[label]++
}

func (c *Counter) Expose(name string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# TYPE %s counter\n", name)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s{code=%q} %d\n", name, k, c.values[k])
	}
	return b.String()
}

func main() {
	var c Counter
	c.Inc("200")
	c.Inc("200")
	c.Inc("500")
	fmt.Print(c.Expose("http_requests_total"))
}
```

### Exercise
Add `Gauge` with `Add(delta int)` and `Expose(name string) string` returning the Prometheus text `# TYPE <name> gauge` followed by `<name> <value>`.
```text expect
# TYPE in_flight gauge
in_flight 1
```
```go solution
package main

import (
	"fmt"
	"sync"
)

// BEGIN
type Gauge struct {
	mu sync.Mutex
	v  int
}

func (g *Gauge) Add(delta int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.v += delta
}

func (g *Gauge) Expose(name string) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return fmt.Sprintf("# TYPE %s gauge\n%s %d\n", name, name, g.v)
}

// END

func main() {
	var g Gauge
	g.Add(2)
	g.Add(-1)
	fmt.Print(g.Expose("in_flight"))
}
```

### Check
Q: Which metric type can decrease?
T: mcq
- [ ] Counter
- [x] Gauge
- [ ] Histogram
- [ ] None
E: Gauges represent current values like queue depth; counters only increase.

Q: Using a user ID as a metric label is a good idea.
T: tf
A: false
E: Every distinct label value creates a new time series; unbounded labels cause a cardinality explosion.

## Logs
slug: logs
status: planned

## Traces
slug: traces
status: planned

## Health Checks
slug: health-checks
minutes: 7
objectives: Explain why services expose health endpoints; Separate liveness from readiness; Keep health handlers fast and dependency-aware
takeaways: Health endpoints let orchestrators decide whether to restart or route to a pod; Liveness answers "is the process stuck?"; readiness answers "can it serve traffic now?"; Health handlers must be cheap and must not cascade failures

### Concept
Kubernetes probes call HTTP endpoints:

| Probe | Question | On failure |
| --- | --- | --- |
| **liveness** `/livez` | Is the process alive and not deadlocked? | container restarted |
| **readiness** `/readyz` | Can it handle requests right now? | removed from Service endpoints |
| **startup** | Has slow initialisation finished? | delays other probes |

Liveness should check **only the process itself** (return 200 if the handler runs). If liveness checks the database, a DB outage restarts *every pod*, making things worse. Readiness may check dependencies (database ping with a short timeout) so traffic stops flowing to an instance that can't serve.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func livez(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") }

func main() {
	rec := httptest.NewRecorder()
	livez(rec, httptest.NewRequest("GET", "/livez", nil))
	fmt.Println(rec.Code, rec.Body.String())
}
```

### Exercise
Write `readyz(check func() error) http.HandlerFunc` returning 200 `ready` when `check()` succeeds and 503 `not ready: <err>` when it fails.
```text expect
200 ready
503 not ready: db down
```
```go solution
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// BEGIN
func readyz(check func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := check(); err != nil {
			http.Error(w, "not ready: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, "ready")
	}
}

// END

func main() {
	for _, check := range []func() error{
		func() error { return nil },
		func() error { return errors.New("db down") },
	} {
		rec := httptest.NewRecorder()
		readyz(check)(rec, httptest.NewRequest("GET", "/readyz", nil))
		fmt.Println(rec.Code, trim(rec.Body.String()))
	}
}

func trim(s string) string {
	for len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}
```

### Check
Q: Why shouldn't a liveness probe check the database?
T: mcq
- [ ] Databases can't be pinged
- [x] A database outage would make every pod fail liveness and restart, worsening the incident
- [ ] It's too slow to matter
- [ ] Liveness must be authenticated
E: Restarting doesn't fix a dependency outage; use readiness to stop routing traffic instead.

Q: Which probe failing removes a pod from a Service's endpoints without restarting it?
T: short
A: readiness
E: Failing readiness stops traffic; failing liveness restarts the container.

## Readiness
slug: readiness
minutes: 6
objectives: Gate readiness on startup work and shutdown; Flip readiness to false before shutting down; Use readiness to shed load safely
takeaways: Report not-ready until caches/connections are warmed; On SIGTERM mark not-ready first so load balancers drain traffic, then stop the server; A ready flag is a good use of atomic.Bool

### Concept
Readiness has a lifecycle:

1. **Starting** — not ready until migrations, cache warm-up and dependency connections finish.
2. **Ready** — receives traffic.
3. **Draining** — on SIGTERM set ready=false, wait a few seconds for endpoints to update, then `srv.Shutdown`.

```go norun
var ready atomic.Bool

func readyz(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprint(w, "ok")
}
```
This avoids the classic gap where a terminating pod still receives requests that fail.

### Example
```go
package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var ready atomic.Bool
	fmt.Println("starting:", ready.Load())
	ready.Store(true)
	fmt.Println("serving:", ready.Load())
	ready.Store(false) // SIGTERM received: drain
	fmt.Println("draining:", ready.Load())
}
```

### Exercise
Write `type Server struct` with `Ready() bool`, `Start()` (marks ready), `Drain()` (marks not ready) using `atomic.Bool`. Print readiness after new, Start and Drain.
```text expect
false true false
```
```go solution
package main

import (
	"fmt"
	"sync/atomic"
)

// BEGIN
type Server struct{ ready atomic.Bool }

func (s *Server) Ready() bool { return s.ready.Load() }
func (s *Server) Start()      { s.ready.Store(true) }
func (s *Server) Drain()      { s.ready.Store(false) }

// END

func main() {
	var s Server
	a := s.Ready()
	s.Start()
	b := s.Ready()
	s.Drain()
	fmt.Println(a, b, s.Ready())
}
```

### Check
Q: What should a service do first when it receives SIGTERM?
T: mcq
- [ ] Exit immediately
- [x] Mark itself not ready so traffic drains, then shut down
- [ ] Ignore it
- [ ] Restart itself
E: Draining avoids sending requests to a pod that's about to stop.

Q: Readiness should report true while the service is still loading its cache.
T: tf
A: false
E: Not-ready until initialisation is finished, so requests aren't served with cold or missing data.

## Liveness
slug: liveness
status: planned

## Correlation IDs
slug: correlation-ids
status: planned

## Request IDs
slug: request-ids
minutes: 8
objectives: Generate a unique ID per request; Propagate it through context and into logs; Return it in a response header
takeaways: A request ID ties together every log line for one request; Accept an incoming X-Request-ID (if trustworthy) or generate one; Store it in the context and add it to every log via a logger

### Concept
When something fails you need to find *all* log lines for that request. Middleware creates the ID, stores it in the context, adds it to the response, and downstream code logs with it:

```go norun
type ctxKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```
Forward the ID on outgoing calls (`req.Header.Set("X-Request-ID", id)`) so it spans services; with tracing it becomes the trace ID. Generate IDs with `crypto/rand` (random hex) or ULIDs.

### Example
```go
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type ctxKey struct{}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func WithID(ctx context.Context, id string) context.Context { return context.WithValue(ctx, ctxKey{}, id) }

func ID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func main() {
	ctx := WithID(context.Background(), newID())
	fmt.Println(len(ID(ctx)), ID(context.Background()) == "")
}
```

### Exercise
Write middleware `RequestID(next http.Handler) http.Handler` that reuses an incoming `X-Request-ID` (or uses `"generated"`) and sets it on the response header. Print the header for a request with and without the incoming header.
```text expect
abc-123 generated
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// BEGIN
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "generated"
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// END

func main() {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Header().Get("X-Request-ID"), rec2.Header().Get("X-Request-ID"))
}
```

### Check
Q: Why forward the request ID on outgoing HTTP calls?
T: mcq
- [ ] It's required by HTTP
- [x] So logs from downstream services can be correlated with the original request
- [ ] To speed up requests
- [ ] To avoid retries
E: A shared ID lets you follow one request across services.

Q: Which package should generate a random request ID?
T: mcq
- [ ] math/rand
- [x] crypto/rand (or a UUID/ULID library)
- [ ] time
- [ ] os
E: Any unique, unpredictable-enough ID works; crypto/rand is a safe default.

## Structured Logging
slug: structured-logging
minutes: 8
objectives: Emit JSON logs with log/slog; Attach request-scoped attributes; Choose levels and avoid logging secrets
takeaways: Structured logs are key/value events, typically JSON, that platforms can index; Use slog.With to attach context (request_id, user_id) once; Log errors once with context; never log secrets or personal data

### Concept
```
Application ──▶ Structured Logs ──▶ Log Platform (Loki, ELK, CloudWatch)
```
```go norun
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("order placed", "order_id", 42, "total_cents", 1999, "user_id", uid)
// {"time":"…","level":"INFO","msg":"order placed","order_id":42,"total_cents":1999,"user_id":"u-1"}
```
- Stable **message**, variable **attributes** (`"order placed"`, not `"order 42 placed"`).
- Add per-request context once: `l := logger.With("request_id", id)`.
- Levels: `Debug` (noisy), `Info` (business events), `Warn` (recoverable), `Error` (needs attention).
- Log to **stdout** in containers; the platform collects it.
- **Never** log passwords, tokens, full card numbers; implement `slog.LogValuer` to redact types.

### Example
```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	opts := &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("service", "billing")
	l.Info("order placed", "order_id", 42, "total_cents", 1999)
}
```

### Exercise
Define `type Secret string` implementing `slog.LogValuer` so logging it prints `[REDACTED]`. Log `"login"` with a `token` attribute holding a `Secret`.
```text expect
{"level":"INFO","msg":"login","token":"[REDACTED]"}
```
```go solution
package main

import (
	"log/slog"
	"os"
)

// BEGIN
type Secret string

func (Secret) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }

// END

func main() {
	opts := &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	l.Info("login", "token", Secret("abc123"))
}
```

### Check
Q: Why keep the log message constant and put variable data in attributes?
T: mcq
- [ ] It prints faster
- [x] Log tools can group identical messages and filter/aggregate on the fields
- [ ] JSON forbids variables in messages
- [ ] It saves disk space only
E: Structured attributes are indexable and queryable.

Q: Which is safe to log?
T: mcq
- [ ] A user's password
- [ ] An API token
- [x] An order ID
- [ ] A full credit card number
E: Never log secrets or sensitive personal data.
