# Production Go
id: production
number: 22
track: advanced-go
paths: pro
skill: architecture
requires: rest-api
summary: Configuration, logging, shutdown, resilience, security and debugging in production.

## Configuration
slug: configuration
minutes: 8
objectives: Load configuration once into an immutable struct; Validate at startup and fail fast; Separate configuration from secrets
takeaways: Parse env/flags/files into a typed Config at startup and pass it explicitly; Validate everything up front — a bad config should stop the process before serving traffic; Treat secrets separately: never log them and prefer mounted files or a secrets manager

### Concept
```go norun
type Config struct {
	Addr         string
	DatabaseURL  string
	ReadTimeout  time.Duration
	LogLevel     slog.Level
}

func Load(getenv func(string) string) (Config, error) { ... }
```
Principles:

- **Typed** — durations as `time.Duration`, sizes as ints; parse once.
- **Explicit** — pass `Config` (or the parts a component needs) into constructors; no package-level globals or `os.Getenv` sprinkled around.
- **Validated** — return one error listing all problems.
- **Immutable** — don't mutate after start; reload by building a new value.
- **Secrets** — keep out of logs and version control; log a *redacted* summary at startup (which settings, not their values).
- **Defaults** that are safe for production, overridable for development.

### Example
```go
package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Addr    string
	Timeout time.Duration
	Workers int
}

func Load(get func(string) string) (Config, error) {
	c := Config{Addr: ":8080", Timeout: 5 * time.Second, Workers: 4}
	var errs []error
	if v := get("ADDR"); v != "" {
		c.Addr = v
	}
	if v := get("TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("TIMEOUT: %w", err))
		}
		c.Timeout = d
	}
	if v := get("WORKERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			errs = append(errs, errors.New("WORKERS: must be a positive integer"))
		}
		c.Workers = n
	}
	return c, errors.Join(errs...)
}

func main() {
	env := map[string]string{"TIMEOUT": "2s", "WORKERS": "8"}
	c, err := Load(func(k string) string { return env[k] })
	fmt.Printf("%+v %v\n", c, err)
}
```

### Exercise
Add validation to `Load`: `Workers` above 64 is an error (`WORKERS: must be at most 64`). Print the error for `WORKERS=100`.
```text expect
WORKERS: must be at most 64
```
```go solution
package main

import (
	"errors"
	"fmt"
	"strconv"
)

type Config struct{ Workers int }

func Load(get func(string) string) (Config, error) {
	c := Config{Workers: 4}
	var errs []error
	if v := get("WORKERS"); v != "" {
		n, err := strconv.Atoi(v)
		switch {
		case err != nil || n < 1:
			errs = append(errs, errors.New("WORKERS: must be a positive integer"))
		// BEGIN
		case n > 64:
			errs = append(errs, errors.New("WORKERS: must be at most 64"))
		// END
		default:
			c.Workers = n
		}
	}
	return c, errors.Join(errs...)
}

func main() {
	_, err := Load(func(k string) string { return map[string]string{"WORKERS": "100"}[k] })
	fmt.Println(err)
}
```

### Check
Q: When should configuration be validated?
T: mcq
- [ ] The first time each setting is used
- [x] Once at startup, before serving traffic
- [ ] Only in CI
- [ ] Never
E: Fail fast: a misconfigured instance should never receive requests.

Q: Which is safest for a database password in production?
T: mcq
- [ ] A constant in the source code
- [ ] A committed .env file
- [x] A secret injected at runtime (mounted file or secrets manager)
- [ ] A command-line flag visible in process listings
E: Command-line arguments and committed files leak easily.

## Structured Logging
slug: structured-logging
status: planned

## Graceful Shutdown
slug: graceful-shutdown
status: planned

## Resource Management
slug: resource-management
status: planned

## Connection Pools
slug: connection-pools
status: planned

## Timeouts
slug: timeouts
minutes: 7
objectives: Set a timeout on every outbound call and server phase; Budget a total deadline across sub-calls; Choose sensible values
takeaways: Every network call needs a timeout — the defaults are infinite; Give each request an overall deadline and derive sub-call timeouts from the remaining time; Timeouts + retries + circuit breakers together prevent cascading failures

### Concept
An unbounded call is a resource leak waiting to happen: a slow dependency piles up goroutines, connections and memory until your service falls over.

Layers to bound:

| Layer | Setting |
| --- | --- |
| HTTP server | `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` |
| HTTP client | `Client.Timeout`, `Transport.ResponseHeaderTimeout`, dial/TLS timeouts |
| Database | `context.WithTimeout` on every query, `SetConnMaxLifetime` |
| Whole request | a deadline in `context` that all calls inherit |

Choose values from **observed latency** (a bit above p99), and keep downstream timeouts **shorter** than the caller's — otherwise the caller gives up while work continues.

### Example
```go
package main

import (
	"context"
	"fmt"
	"time"
)

func call(ctx context.Context, name string, need time.Duration) error {
	select {
	case <-time.After(need):
		fmt.Println(name, "ok")
		return nil
	case <-ctx.Done():
		fmt.Println(name, "gave up:", ctx.Err())
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_ = call(ctx, "cache", 10*time.Millisecond)
	_ = call(ctx, "database", 200*time.Millisecond)
}
```

### Exercise
Write `budget(total, used time.Duration, share float64) time.Duration` returning `share` (0–1) of the **remaining** time (`total - used`, at least 0). Print the budget for `total=1s, used=400ms, share=0.5`.
```text expect
300ms
```
```go solution
package main

import (
	"fmt"
	"time"
)

// BEGIN
func budget(total, used time.Duration, share float64) time.Duration {
	remaining := total - used
	if remaining < 0 {
		remaining = 0
	}
	return time.Duration(float64(remaining) * share)
}

// END

func main() {
	fmt.Println(budget(time.Second, 400*time.Millisecond, 0.5))
}
```

### Check
Q: What is the timeout of a zero-value `http.Client`?
T: mcq
- [ ] 30 seconds
- [x] None
- [ ] 5 seconds
- [ ] 1 minute
E: Always set a timeout.

Q: Why should downstream timeouts be shorter than the caller's?
T: mcq
- [ ] It's faster
- [x] So the caller doesn't give up while the downstream work is still running and wasting resources
- [ ] Downstreams can't time out otherwise
- [ ] It avoids TLS
E: Nested timeouts should shrink as you go deeper.

## Retries
slug: retries
minutes: 8
objectives: Retry only transient, idempotent operations; Use exponential backoff with jitter; Cap attempts and respect context cancellation
takeaways: Retry transient failures (timeouts, 502/503/504, connection reset) — never permanent ones (400, 404) or non-idempotent calls without an idempotency key; Exponential backoff with random jitter prevents retry storms; Always cap attempts and stop when the context ends

### Concept
```go norun
func Retry(ctx context.Context, attempts int, base time.Duration, f func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = f(); err == nil || !retryable(err) {
			return err
		}
		sleep := base << i                       // 100ms, 200ms, 400ms…
		sleep += time.Duration(rand.Int64N(int64(sleep) / 2)) // jitter
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}
```
**Jitter** spreads retries so thousands of clients don't hammer a recovering server in lockstep. Honour `Retry-After` headers. Retries multiply load: combine them with **circuit breakers** and **timeouts**, and retry at one layer only (retries at every layer multiply exponentially).

### Example
```go
package main

import (
	"fmt"
	"time"
)

func backoff(attempt int, base time.Duration) time.Duration {
	return base << attempt
}

func main() {
	for i := 0; i < 4; i++ {
		fmt.Print(backoff(i, 100*time.Millisecond), " ")
	}
	fmt.Println()
}
```

### Exercise
Write `backoff(attempt int, base, max time.Duration) time.Duration` returning `base * 2^attempt` capped at `max`. Print attempts 0–5 with base 100ms and cap 1s.
```text expect
100ms 200ms 400ms 800ms 1s 1s
```
```go solution
package main

import (
	"fmt"
	"strings"
	"time"
)

// BEGIN
func backoff(attempt int, base, max time.Duration) time.Duration {
	d := base << attempt
	if d > max || d <= 0 { // d <= 0 guards against overflow
		return max
	}
	return d
}

// END

func main() {
	var parts []string
	for i := 0; i < 6; i++ {
		parts = append(parts, backoff(i, 100*time.Millisecond, time.Second).String())
	}
	fmt.Println(strings.Join(parts, " "))
}
```

### Check
Q: Which failure should NOT be retried?
T: mcq
- [ ] 503 Service Unavailable
- [ ] A connection reset
- [x] 400 Bad Request
- [ ] A timeout on an idempotent GET
E: A malformed request will fail identically every time.

Q: Why add jitter to backoff?
T: mcq
- [ ] To make waits longer
- [x] To spread out retries so clients don't all retry at the same instant
- [ ] It's required by HTTP
- [ ] To avoid timeouts
E: Synchronised retries can overwhelm a recovering service (thundering herd).

## Resilience
slug: resilience
minutes: 9
objectives: Explain the circuit breaker pattern; Implement a simple breaker with closed, open and half-open states; Combine bulkheads and load shedding
takeaways: A circuit breaker stops calling a failing dependency for a while so it can recover and your callers fail fast; States: closed (normal) → open (fail fast) → half-open (probe) → closed; Also use bulkheads (bounded concurrency), timeouts and graceful degradation

### Concept
```
      failures ≥ N                 cool-down elapsed
CLOSED ───────────▶ OPEN ───────────────────────▶ HALF-OPEN
   ▲                                                 │  │
   └───────────── probe succeeds ────────────────────┘  │
                       OPEN ◀─────── probe fails ───────┘
```
While **open**, calls fail immediately (`ErrOpen`) without touching the struggling service, giving it room to recover and freeing your goroutines. After a cool-down the breaker lets one probe request through (**half-open**).

Other resilience tools: **bulkheads** (limit concurrent calls per dependency with a semaphore), **load shedding** (reject early with 503 when overloaded), **fallbacks** (serve cached/default data), **hedged requests**. Libraries: `sony/gobreaker`, `failsafe-go`.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrOpen = errors.New("circuit open")

type Breaker struct {
	failures, threshold int
}

func (b *Breaker) Call(f func() error) error {
	if b.failures >= b.threshold {
		return ErrOpen
	}
	if err := f(); err != nil {
		b.failures++
		return err
	}
	b.failures = 0
	return nil
}

func main() {
	b := &Breaker{threshold: 2}
	fail := func() error { return errors.New("boom") }
	fmt.Println(b.Call(fail), b.Call(fail), b.Call(fail))
}
```

### Exercise
Extend `Breaker` with a cool-down: inject `now func() time.Time` and `cooldown`; once open for at least `cooldown`, allow one call through (a success closes it, a failure re-opens it with a new open time). Print the results of the sequence shown.
```text expect
boom boom circuit open circuit open <nil>
```
```go solution
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrOpen = errors.New("circuit open")

type Breaker struct {
	failures, threshold int
	openedAt            time.Time
	cooldown            time.Duration
	now                 func() time.Time
}

// BEGIN
func (b *Breaker) Call(f func() error) error {
	if b.failures >= b.threshold {
		if b.now().Sub(b.openedAt) < b.cooldown {
			return ErrOpen
		}
		// half-open: let one probe through
	}
	if err := f(); err != nil {
		b.failures++
		if b.failures >= b.threshold {
			b.openedAt = b.now()
		}
		return err
	}
	b.failures = 0
	return nil
}

// END

func main() {
	clock := time.Unix(0, 0)
	b := &Breaker{threshold: 2, cooldown: time.Minute, now: func() time.Time { return clock }}
	fail := func() error { return errors.New("boom") }
	ok := func() error { return nil }
	var out []string
	rec := func(err error) {
		if err == nil {
			out = append(out, "<nil>")
		} else {
			out = append(out, err.Error())
		}
	}
	rec(b.Call(fail))
	rec(b.Call(fail))
	rec(b.Call(ok)) // open: rejected
	clock = clock.Add(30 * time.Second)
	rec(b.Call(ok)) // still open
	clock = clock.Add(31 * time.Second)
	rec(b.Call(ok)) // cool-down over: probe succeeds
	fmt.Println(strings.Join(out, " "))
}
```

### Check
Q: What does an open circuit breaker do with new calls?
T: mcq
- [ ] Retries them forever
- [x] Fails them immediately without calling the dependency
- [ ] Queues them
- [ ] Sends them to a random server
E: Fast failure protects both the caller and the struggling dependency.

Q: What is a bulkhead?
T: mcq
- [ ] A type of retry
- [x] A limit on concurrent use of a resource so one slow dependency can't consume all capacity
- [ ] A TLS setting
- [ ] A load balancer
E: Named after ship compartments: contain failures locally.

## Security
slug: security
status: planned

## Dependency Management
slug: dependency-management
status: planned

## Production Debugging
slug: production-debugging
status: planned
