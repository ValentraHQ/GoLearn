# Testing
id: testing
number: 15
track: advanced
paths: beginner, pro
skill: testing
requires: errors
project: fully-tested-rest-api
summary: Unit, table-driven, HTTP and integration tests, benchmarks, race detection and coverage.

## Unit Testing
slug: unit-testing
minutes: 8
objectives: Write and run a Go test; Follow the naming and file conventions; Run tests with the go test command
takeaways: Tests live in *_test.go files as func TestXxx(t *testing.T); t.Errorf marks failure and continues, t.Fatalf stops the test; go test ./... runs everything

### Concept
Go's testing support is built in — no framework required.

```go norun
// slug.go
package slug

func Make(s string) string { /* ... */ }

// slug_test.go
package slug

import "testing"

func TestMake(t *testing.T) {
	got := Make("Hello, World!")
	want := "hello-world"
	if got != want {
		t.Errorf("Make() = %q, want %q", got, want)
	}
}
```
Commands you will use constantly:

```bash
go test ./...            # every package
go test -v ./...         # verbose: show each test
go test -run TestMake    # only matching tests
go test -count=1 ./...   # bypass the test cache
go test -race ./...      # data race detector
go test -cover ./...     # coverage summary
go test -bench=.         # benchmarks
```
Use `t.Errorf` to report and keep going (see all failures at once) and `t.Fatalf` when continuing makes no sense (e.g. setup failed). Failure messages should follow **"got X, want Y"** and name the input.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func Slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

func main() {
	fmt.Println(Slug("Hello, World!"), Slug("  Go   is fun "))
}
```

### Exercise
Implement `isLeap(year int) bool` (divisible by 4, except centuries unless divisible by 400). The program checks it against known cases and prints `ok` only if all pass.
```text expect
ok
```
```go solution
package main

import "fmt"

// BEGIN
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// END

func main() {
	cases := map[int]bool{2000: true, 1900: false, 2024: true, 2023: false, 2100: false, 2400: true}
	for year, want := range cases {
		if got := isLeap(year); got != want {
			fmt.Printf("isLeap(%d) = %v, want %v\n", year, got, want)
			return
		}
	}
	fmt.Println("ok")
}
```

### Check
Q: Which file name pattern does `go test` recognise?
T: mcq
- [ ] test_*.go
- [x] *_test.go
- [ ] *Test.go
- [ ] tests/*.go
E: Files ending in `_test.go` are compiled only for tests.

Q: What is the difference between `t.Errorf` and `t.Fatalf`?
T: mcq
- [ ] None
- [x] Errorf records a failure and continues; Fatalf records it and stops the test immediately
- [ ] Fatalf only prints
- [ ] Errorf panics
E: Use Fatalf when the rest of the test can't meaningfully continue.

## testing Package
slug: testing-package
status: planned

## Test Functions
slug: test-functions
status: planned

## Assertions
slug: assertions
minutes: 7
objectives: Write clear manual assertions with got/want messages; Compare complex values with reflect.DeepEqual or go-cmp; Know when an assertion library helps
takeaways: Go tests use plain if statements and t.Errorf rather than assert macros; Use reflect.DeepEqual (or github.com/google/go-cmp) for slices, maps and structs; A good failure message says what was tested, what happened and what was expected

### Concept
```go norun
if got != want {
	t.Errorf("Add(%d, %d) = %d, want %d", a, b, got, want)
}
if !reflect.DeepEqual(got, want) {
	t.Errorf("got %#v, want %#v", got, want)
}
if !errors.Is(err, ErrNotFound) {
	t.Errorf("err = %v, want ErrNotFound", err)
}
```
`%#v` prints Go syntax, ideal for diffs. For rich diffs use `cmp.Diff(want, got)` from go-cmp. Libraries like testify/`assert` are popular; the standard style works fine and keeps tests dependency-free. Mark helper functions with `t.Helper()` so failures point at the caller's line.

### Example
```go
package main

import (
	"fmt"
	"reflect"
)

func check(name string, got, want any) string {
	if reflect.DeepEqual(got, want) {
		return name + ": ok"
	}
	return fmt.Sprintf("%s: got %#v, want %#v", name, got, want)
}

func main() {
	fmt.Println(check("slice", []int{1, 2}, []int{1, 2}))
	fmt.Println(check("map", map[string]int{"a": 1}, map[string]int{"a": 2}))
}
```

### Exercise
Write `assertEqual(name string, got, want any) string` returning `"<name> ok"` when equal (DeepEqual) or `"<name> FAIL: got <got>, want <want>"` using `%v` for both.
```text expect
add ok
list FAIL: got [1 2], want [1 3]
```
```go solution
package main

import (
	"fmt"
	"reflect"
)

// BEGIN
func assertEqual(name string, got, want any) string {
	if reflect.DeepEqual(got, want) {
		return name + " ok"
	}
	return fmt.Sprintf("%s FAIL: got %v, want %v", name, got, want)
}

// END

func main() {
	fmt.Println(assertEqual("add", 1+1, 2))
	fmt.Println(assertEqual("list", []int{1, 2}, []int{1, 3}))
}
```

### Check
Q: Why do failure messages usually follow the "got X, want Y" form?
T: mcq
- [ ] The compiler requires it
- [x] It makes it clear what was tested, what happened and what was expected
- [ ] It runs faster
- [ ] It is needed by go vet
E: A readable failure message saves debugging time.

Q: What does `t.Helper()` do inside a helper function?
T: mcq
- [ ] Runs the helper in parallel
- [x] Makes failure locations point to the caller instead of the helper
- [ ] Skips the test
- [ ] Caches results
E: It removes the helper from the reported file:line.

## Table-Driven Tests
slug: table-driven-tests
minutes: 8
objectives: Structure many cases as a table of inputs and expectations; Name cases so failures are easy to find; Loop with t.Run for subtests
takeaways: Table-driven tests put cases in a slice of structs and loop over them; Give every case a name and print it in failures; Adding a case is one line — the idiomatic Go way to test many inputs

### Concept
```go norun
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Abs(tc.in); got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
```
Since Go 1.22 each iteration has its own `tc`, so parallel subtests (`t.Parallel()`) are safe. Include edge cases (empty, zero, negative, boundaries, invalid) — tables make them cheap. Use a `map[string]struct{...}` when you want the case name to be the key.

### Example
```go
package main

import "fmt"

func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func main() {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
	}
	for _, tc := range tests {
		got := Abs(tc.in)
		fmt.Printf("%-9s pass=%v\n", tc.name, got == tc.want)
	}
}
```

### Exercise
Write `clamp(n, lo, hi int) int` and complete the table so all cases pass, printing `all passed`.
```text expect
all passed
```
```go solution
package main

import "fmt"

// BEGIN
func clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

// END

func main() {
	tests := []struct {
		name          string
		n, lo, hi, want int
	}{
		{"inside", 5, 0, 10, 5},
		{"below", -3, 0, 10, 0},
		{"above", 42, 0, 10, 10},
		{"at low", 0, 0, 10, 0},
		{"at high", 10, 0, 10, 10},
	}
	failed := 0
	for _, tc := range tests {
		if got := clamp(tc.n, tc.lo, tc.hi); got != tc.want {
			fmt.Printf("%s: got %d, want %d\n", tc.name, got, tc.want)
			failed++
		}
	}
	if failed == 0 {
		fmt.Println("all passed")
	}
}
```

### Check
Q: Why give each table case a name?
T: mcq
- [ ] Names speed up the tests
- [x] Failures identify exactly which case broke, and subtests can be run individually
- [ ] It is required by t.Run
- [ ] To sort the cases
E: `go test -run TestAbs/negative` runs one case by name.

Q: What is the main advantage of table-driven tests?
T: mcq
- [ ] They avoid loops
- [x] Adding a new case is a single line and logic isn't duplicated
- [ ] They run in parallel automatically
- [ ] They need no test function
E: One loop, many inputs.

## Subtests
slug: subtests
minutes: 7
objectives: Use t.Run to group and name cases; Select subtests with -run patterns; Run independent subtests in parallel
takeaways: t.Run(name, func(t *testing.T)) creates a named subtest; go test -run 'TestX/name' runs matching subtests; t.Parallel() lets independent subtests run concurrently

### Concept
```go norun
func TestUser(t *testing.T) {
	t.Run("valid", func(t *testing.T) { ... })
	t.Run("missing email", func(t *testing.T) { ... })
}
```
Subtest names have spaces replaced by underscores: `TestUser/missing_email`. `-run` takes a `/`-separated list of regular expressions, one per level:

```bash
go test -run 'TestUser/valid'
go test -run '/missing'         # any top-level test, subtest matching missing
```
`t.Parallel()` inside a subtest pauses it until the parent's serial part finishes, then runs it alongside siblings. Share setup in the parent and cleanup with `t.Cleanup`.

### Example
```go
package main

import (
	"fmt"
	"regexp"
	"strings"
)

// matches mimics how -run patterns select tests: one regexp per "/" level.
func matches(pattern, fullName string) bool {
	pats := strings.Split(pattern, "/")
	names := strings.Split(fullName, "/")
	for i, p := range pats {
		if i >= len(names) || !regexp.MustCompile(p).MatchString(names[i]) {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println(matches("TestUser/valid", "TestUser/valid"), matches("TestUser/valid", "TestUser/invalid_email"))
}
```

### Exercise
Write `subtestName(s string) string` replacing spaces with underscores as `t.Run` does (`"missing email"` → `"missing_email"`).
```text expect
missing_email empty
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func subtestName(s string) string {
	if s == "" {
		return "empty"
	}
	return strings.ReplaceAll(s, " ", "_")
}

// END

func main() {
	fmt.Println(subtestName("missing email"), subtestName(""))
}
```

### Check
Q: Which command runs only the subtest `valid` of `TestUser`?
T: mcq
- [ ] go test -run valid
- [x] go test -run 'TestUser/valid'
- [ ] go test -sub TestUser.valid
- [ ] go test TestUser:valid
E: `-run` accepts slash-separated patterns for nested subtests.

Q: What does `t.Parallel()` inside a subtest do?
T: mcq
- [ ] Makes the whole test binary parallel
- [x] Lets that subtest run concurrently with other parallel subtests
- [ ] Skips the subtest
- [ ] Repeats it
E: Parallel subtests resume together after the parent function's sequential part completes.

## Test Fixtures
slug: test-fixtures
minutes: 7
objectives: Use t.TempDir and t.Cleanup for isolated state; Store input and golden files in testdata/; Set up shared state with TestMain sparingly
takeaways: t.TempDir gives an isolated directory that is removed automatically; t.Cleanup registers teardown that runs in LIFO order; The testdata/ directory is ignored by the build and holds fixtures and golden files

### Concept
```go norun
func TestSave(t *testing.T) {
	dir := t.TempDir()                       // removed after the test
	path := filepath.Join(dir, "out.json")
	t.Cleanup(func() { /* extra teardown */ })
	...
}

want, _ := os.ReadFile("testdata/report.golden")
```
- **`t.TempDir()`** — private scratch directory per test.
- **`t.Cleanup(f)`** — runs `f` when the test finishes, **last registered first** (like `defer`).
- **`testdata/`** — Go tooling skips this directory; keep sample inputs and *golden files* (expected output) here.
- **`TestMain(m *testing.M)`** — package-level setup/teardown (start a container, seed a database). Use sparingly — per-test fixtures isolate better.
- `t.Setenv("KEY", "v")` sets an environment variable for the test only.

### Example
```go
package main

import (
	"fmt"
)

type Cleanups struct{ fns []func() }

func (c *Cleanups) Add(f func()) { c.fns = append(c.fns, f) }

func (c *Cleanups) Run() {
	for i := len(c.fns) - 1; i >= 0; i-- {
		c.fns[i]()
	}
}

func main() {
	var c Cleanups
	c.Add(func() { fmt.Println("close database") })
	c.Add(func() { fmt.Println("remove temp dir") })
	c.Run()
}
```

### Exercise
Implement `Cleanups.Run` so cleanups run in **reverse** order of registration (LIFO), as `t.Cleanup` does.
```text expect
third
second
first
```
```go solution
package main

import "fmt"

type Cleanups struct{ fns []func() }

func (c *Cleanups) Add(f func()) { c.fns = append(c.fns, f) }

// BEGIN
func (c *Cleanups) Run() {
	for i := len(c.fns) - 1; i >= 0; i-- {
		c.fns[i]()
	}
}

// END

func main() {
	var c Cleanups
	for _, name := range []string{"first", "second", "third"} {
		c.Add(func() { fmt.Println(name) })
	}
	c.Run()
}
```

### Check
Q: In what order do `t.Cleanup` functions run?
T: mcq
- [ ] Registration order
- [x] Reverse registration order (last registered first)
- [ ] Random
- [ ] Alphabetical
E: Like `defer`, cleanup is LIFO so resources are released in the reverse of how they were acquired.

Q: What is special about a directory named `testdata`?
T: mcq
- [ ] It's compiled first
- [x] The go tool ignores it, so it's the conventional place for fixtures
- [ ] It runs automatically
- [ ] It's excluded from git
E: Go tooling skips `testdata` when building packages.

## Mocks
slug: mocks
minutes: 8
objectives: Use hand-written fakes and spies instead of heavy mocking frameworks; Verify interactions when they matter; Decide between fake, stub and spy
takeaways: A stub returns canned data, a spy records calls, a fake is a working lightweight implementation; Prefer fakes/stubs and assert outcomes over asserting call sequences; Generated mocks (gomock, mockery) exist but are rarely needed for small interfaces

### Concept
Because Go interfaces are implicit, a test double is just a struct with the right methods.

```go norun
type spyMailer struct {
	sent  []string
	err   error // stub: make Send fail on demand
}

func (m *spyMailer) Send(to, body string) error {
	m.sent = append(m.sent, to)
	return m.err
}
```
- **Stub** — returns canned values.
- **Spy** — records how it was called.
- **Fake** — a simplified working implementation (in-memory repository).
- **Mock** — pre-programmed with expectations that fail the test if unmet.

Assert on **observable outcomes** (state, return values) before asserting *how* something was called; over-specified interaction tests break on harmless refactors.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

type Mailer interface{ Send(to string) error }

type spy struct {
	sent []string
	err  error
}

func (s *spy) Send(to string) error {
	s.sent = append(s.sent, to)
	return s.err
}

func Welcome(m Mailer, users []string) (sent int, err error) {
	for _, u := range users {
		if e := m.Send(u); e != nil {
			return sent, e
		}
		sent++
	}
	return sent, nil
}

func main() {
	s := &spy{}
	n, err := Welcome(s, []string{"a", "b"})
	fmt.Println(n, err, s.sent)
	s2 := &spy{err: errors.New("smtp down")}
	fmt.Println(Welcome(s2, []string{"a"}))
}
```

### Exercise
Write a spy `callCounter` implementing `Greeter` (`Greet(name string) string`) that returns `hi <name>` and counts calls. After greeting two names print the count.
```text expect
2 hi ada
```
```go solution
package main

import "fmt"

type Greeter interface{ Greet(name string) string }

// BEGIN
type callCounter struct{ calls int }

func (c *callCounter) Greet(name string) string {
	c.calls++
	return "hi " + name
}

// END

func use(g Greeter) string {
	g.Greet("bob")
	return g.Greet("ada")
}

func main() {
	c := &callCounter{}
	last := use(c)
	fmt.Println(c.calls, last)
}
```

### Check
Q: What is the difference between a stub and a spy?
T: mcq
- [ ] None
- [x] A stub returns canned answers; a spy also records how it was used
- [ ] A spy is generated by tools
- [ ] A stub never fails
E: Spies let tests assert on interactions; stubs only control inputs.

Q: Over-specifying exact call sequences in tests makes them brittle.
T: tf
A: true
E: Prefer asserting outcomes; interaction checks should be limited to behaviour that matters.

## Interfaces for Testing
slug: interfaces-for-testing
minutes: 7
objectives: Introduce seams for time, randomness and I/O; Inject a clock instead of calling time.Now; Keep production code free of test-only hooks
takeaways: Anything non-deterministic (time, randomness, network, filesystem) needs a seam; Inject a clock (func() time.Time) or interface so tests control it; Small consumer-defined interfaces keep test doubles trivial

### Concept
Code that calls `time.Now()` or `rand.Int()` directly can't be tested deterministically. Inject the dependency:

```go norun
type Service struct {
	now func() time.Time // production: time.Now
}
func (s Service) Expired(t Token) bool { return s.now().After(t.Expiry) }
```
In tests pass a fixed function: `Service{now: func() time.Time { return fixed }}`. The same idea applies to the filesystem (`fs.FS`), HTTP (`http.RoundTripper`, `httptest.Server`), randomness (`*rand.Rand` with a seed) and ID generators.

### Example
```go
package main

import (
	"fmt"
	"time"
)

type Token struct{ Expiry time.Time }

type Service struct{ now func() time.Time }

func (s Service) Expired(t Token) bool { return s.now().After(t.Expiry) }

func main() {
	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s := Service{now: func() time.Time { return fixed }}
	fmt.Println(s.Expired(Token{fixed.Add(time.Hour)}), s.Expired(Token{fixed.Add(-time.Hour)}))
}
```

### Exercise
Write `Backoff(attempt int, base time.Duration, jitter func() float64) time.Duration` returning `base * 2^attempt` scaled by `1 + jitter()` (jitter in [0,1)). Inject a fixed jitter of `0.5` and print the result for attempt 2 with a 100ms base.
```text expect
600ms
```
```go solution
package main

import (
	"fmt"
	"time"
)

// BEGIN
func Backoff(attempt int, base time.Duration, jitter func() float64) time.Duration {
	d := base << attempt
	return time.Duration(float64(d) * (1 + jitter()))
}

// END

func main() {
	fmt.Println(Backoff(2, 100*time.Millisecond, func() float64 { return 0.5 }))
}
```

### Check
Q: How do you make code that needs "the current time" testable?
T: mcq
- [ ] Sleep in the test
- [x] Inject a clock function or interface
- [ ] Set the system clock
- [ ] Avoid testing it
E: A fixed clock makes time-dependent logic deterministic.

Q: Should production code contain flags like `isTest` to help tests?
T: tf
A: false
E: Use injection and interfaces instead of test-only branches in production code.

## Integration Tests
slug: integration-tests
status: planned

## HTTP Tests
slug: http-tests
minutes: 8
objectives: Test handlers with httptest.NewRecorder; Test clients against httptest.NewServer; Assert status, headers and JSON bodies
takeaways: httptest.NewRecorder captures a handler's response without a network; httptest.NewServer starts a real local server for client and end-to-end tests; Test the mux (router) to cover routing plus handlers together

### Concept
```go norun
req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
rec := httptest.NewRecorder()
mux.ServeHTTP(rec, req)

res := rec.Result()
if res.StatusCode != http.StatusOK { ... }
```
Test through the **router** so pattern matching and middleware are exercised. Decode JSON responses into structs rather than comparing raw strings (field order and whitespace shouldn't matter). For code that calls out over HTTP use `httptest.NewServer` (a real listener on localhost) or a custom `http.RoundTripper` to fake responses.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": r.PathValue("id")})
	})
	return mux
}

func main() {
	tests := []struct {
		method, path string
		want         int
	}{
		{"GET", "/users/42", 200},
		{"POST", "/users/42", 405},
		{"GET", "/nope", 404},
	}
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		newMux().ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		fmt.Println(tc.method, tc.path, rec.Code == tc.want)
	}
}
```

### Exercise
Write a table test loop that requests `/health` and `/missing` against the mux and prints `<path> <status>` for each.
```text expect
/health 200
/missing 404
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	// BEGIN
	for _, path := range []string{"/health", "/missing"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		fmt.Println(path, rec.Code)
	}
	// END
}
```

### Check
Q: Which type captures a handler's output in a test without a network?
T: short
A: httptest.ResponseRecorder
E: `httptest.NewRecorder()` returns one; call `rec.Result()` for the `*http.Response`.

Q: Why decode JSON responses into structs instead of comparing strings?
T: mcq
- [ ] Strings can't be compared
- [x] Whitespace and key order in JSON don't matter semantically
- [ ] It's faster
- [ ] Recorders return structs
E: Semantic comparison avoids brittle tests.

## Benchmarks
slug: benchmarks
minutes: 8
objectives: Write a benchmark with b.N; Read ns/op and allocs/op; Avoid common benchmarking mistakes
takeaways: A benchmark is func BenchmarkXxx(b *testing.B) looping b.N times; Run with go test -bench=. -benchmem to see ns/op and allocations; Guard against the compiler optimising away results and use b.ResetTimer after setup

### Concept
```go norun
func BenchmarkJoin(b *testing.B) {
	parts := strings.Split(strings.Repeat("x,", 100), ",")
	b.ResetTimer()                       // exclude setup
	for i := 0; i < b.N; i++ {
		_ = strings.Join(parts, ",")
	}
}
```
```bash
go test -bench=. -benchmem
go test -bench=Join -count=5 | benchstat   # compare statistically
```
Since Go 1.24 you can write `for b.Loop() { ... }`, which handles the timer and keeps the result alive for you. Never trust a single noisy run; compare with `benchstat`. Benchmarks are for *comparing* implementations on the same machine, not absolute truth.

`testing.Benchmark(f)` lets you run a benchmark from ordinary code — used below.

### Example
```go
package main

import (
	"fmt"
	"strings"
	"testing"
)

func concat(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "x"
	}
	return s
}

func builder(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte('x')
	}
	return b.String()
}

func main() {
	slow := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			concat(200)
		}
	})
	fast := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			builder(200)
		}
	})
	fmt.Println("builder allocates less:", fast.AllocsPerOp() < slow.AllocsPerOp())
}
```

### Exercise
Benchmark appending 1000 ints to a slice with and without preallocation. Print whether the preallocated version performs fewer allocations per operation.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"testing"
)

func growing(n int) []int {
	var out []int
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	return out
}

func prealloc(n int) []int {
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	return out
}

func main() {
	// BEGIN
	a := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			growing(1000)
		}
	})
	p := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			prealloc(1000)
		}
	})
	fmt.Println(p.AllocsPerOp() < a.AllocsPerOp())
	// END
}
```

### Check
Q: What does `go test -bench=. -benchmem` add to the output?
T: mcq
- [ ] Coverage
- [x] Allocation statistics (B/op, allocs/op)
- [ ] Race detection
- [ ] Verbose logging
E: `-benchmem` reports bytes and allocations per operation.

Q: Why call `b.ResetTimer()` after expensive setup?
T: mcq
- [ ] To restart the loop
- [x] So setup time isn't included in the measurement
- [ ] To skip warm-up
- [ ] It's required to compile
E: Only the code inside the loop should be timed.

## Race Detection
slug: race-detection
status: planned

## Test Coverage
slug: test-coverage
minutes: 7
objectives: Measure coverage with go test -cover and -coverprofile; Read a coverage profile; Use coverage as a guide, not a goal
takeaways: go test -cover reports the percentage of statements executed; go test -coverprofile=c.out && go tool cover -html=c.out shows which lines were missed; 100% coverage doesn't mean correct — focus on meaningful behaviour and edge cases

### Concept
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out        # per-function percentages
go tool cover -html=coverage.out        # browsable HTML report
```
A profile line looks like `pkg/file.go:12.30,14.2 1 1` = `file:startLine.col,endLine.col numStatements count`. Coverage shows what code ran, **not** whether assertions were meaningful. Use it to find *untested* branches (error paths!), and set a sensible CI threshold (e.g. 80% for core packages) rather than chasing 100%.

### Example
```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// percent parses a Go coverage-profile line and returns statements and hits.
func parse(line string) (stmts, count int) {
	fields := strings.Fields(line)
	stmts, _ = strconv.Atoi(fields[1])
	count, _ = strconv.Atoi(fields[2])
	return
}

func main() {
	fmt.Println(parse("pkg/file.go:12.30,14.2 3 1"))
}
```

### Exercise
Write `coverage(profile []string) float64` returning the percentage of statements with a hit count > 0 (`covered/total*100`), given profile lines in the format above.
```text expect
62.5
```
```go solution
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// BEGIN
func coverage(profile []string) float64 {
	total, covered := 0, 0
	for _, line := range profile {
		f := strings.Fields(line)
		if len(f) != 3 {
			continue
		}
		stmts, _ := strconv.Atoi(f[1])
		count, _ := strconv.Atoi(f[2])
		total += stmts
		if count > 0 {
			covered += stmts
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

// END

func main() {
	fmt.Println(coverage([]string{
		"a.go:1.1,3.2 5 1",
		"a.go:4.1,6.2 3 0",
	}))
}
```

### Check
Q: Which command produces a browsable HTML coverage report from a profile?
T: mcq
- [ ] go test -html
- [x] go tool cover -html=coverage.out
- [ ] go cover report
- [ ] go vet -cover
E: First create the profile with `go test -coverprofile=coverage.out`, then render it.

Q: 100% statement coverage guarantees the code is correct.
T: tf
A: false
E: Coverage shows code was executed, not that assertions checked the right things.
