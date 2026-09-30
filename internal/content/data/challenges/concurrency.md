# Concurrency Challenges

## worker-pool-results
title: Bounded Worker Pool
difficulty: intermediate
module: concurrency
lesson: concurrency/worker-pools
skill: concurrency

### Problem
Write `ProcessAll(inputs []int, workers int, f func(int) int) []int`. It applies `f` to every input using **at most `workers` goroutines at the same time** and returns the results **in input order**. `workers < 1` is treated as 1. Empty input returns an empty slice.

### Constraints
- Never run more than `workers` calls to `f` concurrently
- Results must line up with `inputs`
- Must be race-free

### Starter
```go
package main

func ProcessAll(inputs []int, workers int, f func(int) int) []int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrderAndValues(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	got := ProcessAll(in, 3, func(n int) int { return n * n })
	want := []int{1, 4, 9, 16, 25, 36, 49, 64}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBoundedConcurrency(t *testing.T) {
	var cur, max atomic.Int32
	in := make([]int, 20)
	ProcessAll(in, 4, func(n int) int {
		c := cur.Add(1)
		for {
			m := max.Load()
			if c <= m || max.CompareAndSwap(m, c) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		cur.Add(-1)
		return n
	})
	if m := max.Load(); m > 4 {
		t.Errorf("observed %d concurrent calls, limit was 4", m)
	}
	if m := max.Load(); m < 2 {
		t.Errorf("expected some parallelism, max concurrency was %d", m)
	}
}

func TestEdgeCases(t *testing.T) {
	if got := ProcessAll(nil, 3, func(n int) int { return n }); len(got) != 0 {
		t.Errorf("empty: %v", got)
	}
	if got := ProcessAll([]int{2, 3}, 0, func(n int) int { return n + 1 }); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("workers=0: %v", got)
	}
}
```

### Hints
- Feed `(index, value)` pairs through a channel to the workers
- Each worker writes `out[index]`; distinct indexes mean no race
- `wg.Wait()` before returning

### Solution
```go
package main

import "sync"

func ProcessAll(inputs []int, workers int, f func(int) int) []int {
	if workers < 1 {
		workers = 1
	}
	out := make([]int, len(inputs))
	type job struct{ i, n int }
	jobs := make(chan job)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				out[j.i] = f(j.n)
			}
		}()
	}
	for i, n := range inputs {
		jobs <- job{i, n}
	}
	close(jobs)
	wg.Wait()
	return out
}

func main() {}
```

### Explanation
A fixed number of workers reading one channel bounds concurrency. Writing to distinct slice indexes preserves order without locks.

## merge-channels
title: Merge Channels
difficulty: intermediate
module: concurrency
lesson: concurrency/fan-in
skill: concurrency

### Problem
Write `Merge(cs ...<-chan int) <-chan int` that fans several channels into one. The returned channel must be closed exactly once, after **all** inputs are closed and drained. `Merge()` with no inputs returns an already-closed channel.

### Constraints
- Do not lose or duplicate values
- Close the output only after every input is exhausted

### Starter
```go
package main

func Merge(cs ...<-chan int) <-chan int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"sort"
	"testing"
	"time"
)

func gen(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func collect(t *testing.T, ch <-chan int) []int {
	t.Helper()
	var out []int
	timeout := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-timeout:
			t.Fatal("timed out: output channel never closed")
		}
	}
}

func TestMerge(t *testing.T) {
	got := collect(t, Merge(gen(1, 2, 3), gen(10, 20), gen(100)))
	sort.Ints(got)
	want := []int{1, 2, 3, 10, 20, 100}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestMergeEmpty(t *testing.T) {
	if got := collect(t, Merge()); len(got) != 0 {
		t.Errorf("no inputs: %v", got)
	}
	if got := collect(t, Merge(gen(), gen())); len(got) != 0 {
		t.Errorf("empty inputs: %v", got)
	}
}

func TestMergeKeepsPerInputOrder(t *testing.T) {
	got := collect(t, Merge(gen(1, 2, 3, 4, 5), gen(101, 102, 103)))
	last := map[bool]int{}
	for _, v := range got {
		big := v > 100
		if v <= last[big] {
			t.Fatalf("order within one input broken: %v", got)
		}
		last[big] = v
	}
}
```

### Hints
- One forwarding goroutine per input
- A `sync.WaitGroup` plus a closer goroutine closes `out`
- Call `wg.Add(len(cs))` before starting the forwarders

### Solution
```go
package main

import "sync"

func Merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {}
```

### Explanation
The closer goroutine is the heart of fan-in: it waits for every forwarder and then closes the output exactly once, so consumers can simply `range`.

## pipeline-stages
title: Pipeline Stages
difficulty: intermediate
module: concurrency
lesson: concurrency/pipelines
skill: concurrency

### Problem
Implement three composable pipeline stages:

- `Gen(nums ...int) <-chan int` emits the numbers then closes
- `Filter(in <-chan int, keep func(int) bool) <-chan int` forwards values for which `keep` is true
- `Map(in <-chan int, f func(int) int) <-chan int` forwards `f(v)` for each value

Each stage must close its output when its input closes.

### Constraints
- Each stage runs in its own goroutine
- Preserve order

### Starter
```go
package main

func Gen(nums ...int) <-chan int {
	return nil
}

func Filter(in <-chan int, keep func(int) bool) <-chan int {
	return nil
}

func Map(in <-chan int, f func(int) int) <-chan int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
	"time"
)

func drain(t *testing.T, ch <-chan int) []int {
	t.Helper()
	var out []int
	timeout := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-timeout:
			t.Fatal("timed out: channel never closed")
		}
	}
}

func TestGen(t *testing.T) {
	if got := drain(t, Gen(3, 1, 2)); !reflect.DeepEqual(got, []int{3, 1, 2}) {
		t.Errorf("got %v", got)
	}
	if got := drain(t, Gen()); len(got) != 0 {
		t.Errorf("empty Gen: %v", got)
	}
}

func TestFilterMap(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	sq := func(n int) int { return n * n }
	got := drain(t, Map(Filter(Gen(1, 2, 3, 4, 5, 6), even), sq))
	if !reflect.DeepEqual(got, []int{4, 16, 36}) {
		t.Errorf("got %v", got)
	}
}

func TestNothingPasses(t *testing.T) {
	got := drain(t, Filter(Gen(1, 3, 5), func(n int) bool { return n%2 == 0 }))
	if len(got) != 0 {
		t.Errorf("got %v", got)
	}
}
```

### Hints
- Each stage: create `out`, start a goroutine that ranges over the input and sends, `defer close(out)`
- Return `out` immediately so stages run concurrently

### Solution
```go
package main

func Gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func Filter(in <-chan int, keep func(int) bool) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			if keep(v) {
				out <- v
			}
		}
	}()
	return out
}

func Map(in <-chan int, f func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- f(v)
		}
	}()
	return out
}

func main() {}
```

### Explanation
Every stage follows the same contract — read until the input closes, own and close the output — which is what makes stages composable.

## cancellable-worker
title: Cancellable Generator
difficulty: intermediate
module: concurrency
lesson: concurrency/cancellation
skill: concurrency

### Problem
Write `Generate(ctx context.Context) <-chan int` that emits 0, 1, 2, … on the returned channel until `ctx` is cancelled, then **closes** the channel and lets its goroutine exit. Even if the consumer stops reading, cancelling `ctx` must free the goroutine.

### Constraints
- The goroutine must exit after cancellation (no leak)
- The channel must be closed after cancellation

### Starter
```go
package main

import "context"

func Generate(ctx context.Context) <-chan int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestGenerateValuesInOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := Generate(ctx)
	for want := 0; want < 5; want++ {
		select {
		case got := <-ch:
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for a value")
		}
	}
}

func TestGenerateClosesOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := Generate(ctx)
	<-ch
	cancel()
	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("channel was not closed after cancel")
		}
	}
}

func TestNoLeakWhenConsumerStops(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		ch := Generate(ctx)
		<-ch
		cancel() // consumer walks away without draining
	}
	for i := 0; i < 100; i++ {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutines leaked: before=%d after=%d", before, runtime.NumGoroutine())
}
```

### Hints
- Loop with `select { case <-ctx.Done(): return; case out <- i: }`
- `defer close(out)` at the top of the goroutine

### Solution
```go
package main

import "context"

func Generate(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case out <- i:
			}
		}
	}()
	return out
}

func main() {}
```

### Explanation
Sending inside a `select` with `ctx.Done()` means a blocked send can always be abandoned, so a consumer that walks away can't leak the goroutine.

## graceful-stop
title: Pool With Graceful Stop
difficulty: advanced
module: concurrency
lesson: concurrency/graceful-shutdown
skill: concurrency

### Problem
Implement a job pool:

- `NewPool(workers, queue int) *Pool` — starts `workers` goroutines reading from a queue of size `queue`
- `Submit(job func()) error` — enqueues a job; returns `ErrClosed` once the pool is stopping
- `Stop(ctx context.Context) error` — stops accepting jobs, lets workers **finish every job already accepted**, and returns `nil` when they are done, or `ctx.Err()` if `ctx` ends first

`Stop` may be called more than once. `ErrClosed` is provided.

### Constraints
- Every job for which `Submit` returned nil must run before `Stop` returns nil
- `Submit` after `Stop` returns `ErrClosed`
- Must be race-free (Submit and Stop can run concurrently)

### Starter
```go
package main

import (
	"context"
	"errors"
)

var ErrClosed = errors.New("pool closed")

type Pool struct {
}

func NewPool(workers, queue int) *Pool {
	return &Pool{}
}

func (p *Pool) Submit(job func()) error {
	return nil
}

func (p *Pool) Stop(ctx context.Context) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllAcceptedJobsRun(t *testing.T) {
	p := NewPool(3, 10)
	var done atomic.Int32
	for i := 0; i < 10; i++ {
		if err := p.Submit(func() {
			time.Sleep(2 * time.Millisecond)
			done.Add(1)
		}); err != nil {
			t.Fatalf("submit %d: %v", i, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if got := done.Load(); got != 10 {
		t.Errorf("ran %d of 10 accepted jobs", got)
	}
}

func TestSubmitAfterStop(t *testing.T) {
	p := NewPool(1, 1)
	ctx := context.Background()
	if err := p.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(func() {}); !errors.Is(err, ErrClosed) {
		t.Errorf("want ErrClosed, got %v", err)
	}
	if err := p.Stop(ctx); err != nil {
		t.Errorf("second Stop: %v", err)
	}
}

func TestStopHonoursContext(t *testing.T) {
	p := NewPool(1, 1)
	release := make(chan struct{})
	_ = p.Submit(func() { <-release })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := p.Stop(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("want DeadlineExceeded, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("Stop ignored the context deadline")
	}
	close(release)
}

func TestConcurrentSubmitAndStop(t *testing.T) {
	p := NewPool(4, 100)
	var ran, accepted atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			if err := p.Submit(func() { ran.Add(1) }); err == nil {
				accepted.Add(1)
			}
		}
	}()
	time.Sleep(time.Millisecond)
	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-done
	if ran.Load() != accepted.Load() {
		t.Errorf("accepted %d jobs but ran %d", accepted.Load(), ran.Load())
	}
}
```

### Hints
- A mutex-protected `closed` flag makes `Submit` and `Stop` race-free
- Close the queue channel in `Stop` (once, via `sync.Once` or under the mutex)
- Wait for workers with a `WaitGroup` inside a goroutine that signals a `done` channel; `select` on it and `ctx.Done()`

### Solution
```go
package main

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("pool closed")

type Pool struct {
	mu     sync.Mutex
	closed bool
	jobs   chan func()
	wg     sync.WaitGroup
}

func NewPool(workers, queue int) *Pool {
	p := &Pool{jobs: make(chan func(), queue)}
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				job()
			}
		}()
	}
	return p
}

func (p *Pool) Submit(job func()) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrClosed
	}
	p.jobs <- job // holds the lock while the queue is full: Stop waits its turn
	return nil
}

func (p *Pool) Stop(ctx context.Context) error {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {}
```

### Explanation
The mutex guarantees a job is either accepted before the queue closes or rejected with `ErrClosed` — never sent on a closed channel. Waiting on the `WaitGroup` in a helper goroutine lets `Stop` also honour the caller's deadline.

## fix-goroutine-leak
title: Fix the Goroutine Leak
difficulty: intermediate
module: concurrency
lesson: concurrency/goroutine-leaks
skill: concurrency

### Problem
`FirstResult` runs every function concurrently and returns the first result. In the starter, the losers block forever on an unbuffered channel, leaking goroutines. Fix it so that **all goroutines finish** once their functions return, while `FirstResult` still returns as soon as the first result is ready. With no functions it returns `""` immediately.

### Constraints
- Return as soon as the first result arrives
- No goroutine may be left blocked after every function has returned

### Starter
```go
package main

func FirstResult(fs ...func() string) string {
	ch := make(chan string)
	for _, f := range fs {
		go func() { ch <- f() }()
	}
	return <-ch
}

func main() {}
```

### Tests
```go
package main

import (
	"runtime"
	"testing"
	"time"
)

func sleeper(name string, d time.Duration) func() string {
	return func() string {
		time.Sleep(d)
		return name
	}
}

func TestReturnsFirst(t *testing.T) {
	start := time.Now()
	got := FirstResult(sleeper("slow", 150*time.Millisecond), sleeper("fast", time.Millisecond), sleeper("mid", 50*time.Millisecond))
	if got != "fast" {
		t.Errorf("got %q, want fast", got)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Errorf("did not return as soon as the first result was ready (%v)", time.Since(start))
	}
}

func TestNoFunctions(t *testing.T) {
	done := make(chan string, 1)
	go func() { done <- FirstResult() }()
	select {
	case got := <-done:
		if got != "" {
			t.Errorf("got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("FirstResult() with no functions blocked")
	}
}

func TestNoLeak(t *testing.T) {
	time.Sleep(20 * time.Millisecond)
	before := runtime.NumGoroutine()
	for i := 0; i < 5; i++ {
		FirstResult(sleeper("a", 30*time.Millisecond), sleeper("b", time.Millisecond), sleeper("c", 20*time.Millisecond))
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutines leaked: before=%d after=%d", before, runtime.NumGoroutine())
}
```

### Hints
- Give the channel a buffer as large as the number of functions
- Handle `len(fs) == 0` before receiving

### Solution
```go
package main

func FirstResult(fs ...func() string) string {
	if len(fs) == 0 {
		return ""
	}
	ch := make(chan string, len(fs)) // every send succeeds, so no goroutine can block
	for _, f := range fs {
		go func() { ch <- f() }()
	}
	return <-ch
}

func main() {}
```

### Explanation
Buffering the channel to the number of senders guarantees each send completes, so the goroutines that lost the race simply finish. It is the simplest leak-proof fix when the number of senders is known.

## compute-once-cache
title: Compute-Once Cache
difficulty: advanced
module: concurrency
lesson: concurrency/race-conditions
skill: concurrency

### Problem
Implement `Cache` with `GetOrCompute(key string, compute func() string) string`. If the value is cached return it. If several goroutines ask for the **same missing key at the same time**, `compute` must run **exactly once** for that key and all callers get its result. Different keys must be computable concurrently (a slow `compute` for one key must not block another key). The zero value must be usable.

### Constraints
- `compute` runs at most once per key
- Different keys must not block each other
- Must be race-free

### Starter
```go
package main

type Cache struct {
	data map[string]string
}

func (c *Cache) GetOrCompute(key string, compute func() string) string {
	if c.data == nil {
		c.data = map[string]string{}
	}
	if v, ok := c.data[key]; ok {
		return v
	}
	v := compute()
	c.data[key] = v
	return v
}

func main() {}
```

### Tests
```go
package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestComputeOncePerKey(t *testing.T) {
	var c Cache
	var calls atomic.Int32
	var wg sync.WaitGroup
	results := make([]string, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = c.GetOrCompute("k", func() string {
				calls.Add(1)
				time.Sleep(30 * time.Millisecond)
				return "value"
			})
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Errorf("compute ran %d times, want 1", n)
	}
	for i, r := range results {
		if r != "value" {
			t.Errorf("caller %d got %q", i, r)
		}
	}
	if got := c.GetOrCompute("k", func() string { return "other" }); got != "value" {
		t.Errorf("cached value lost: %q", got)
	}
}

func TestDifferentKeysRunInParallel(t *testing.T) {
	var c Cache
	start := time.Now()
	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c", "d"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.GetOrCompute(k, func() string {
				time.Sleep(60 * time.Millisecond)
				return k
			})
		}()
	}
	wg.Wait()
	if d := time.Since(start); d > 180*time.Millisecond {
		t.Errorf("keys were computed serially (%v)", d)
	}
}
```

### Hints
- Guard the map with a mutex, but don't hold it while computing
- Store a per-key entry containing a `sync.Once` (or a done channel) so waiters block only on that key
- Insert the entry under the lock, then run `entry.once.Do(...)` outside it

### Solution
```go
package main

import "sync"

type entry struct {
	once sync.Once
	val  string
}

type Cache struct {
	mu   sync.Mutex
	data map[string]*entry
}

func (c *Cache) GetOrCompute(key string, compute func() string) string {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[string]*entry{}
	}
	e, ok := c.data[key]
	if !ok {
		e = &entry{}
		c.data[key] = e
	}
	c.mu.Unlock()

	e.once.Do(func() { e.val = compute() })
	return e.val
}

func main() {}
```

### Explanation
The map lock is held only to find or insert the entry. The per-key `sync.Once` serialises computation for that key alone: the first caller runs `compute`, everyone else waits on the same `Once` and then reads the shared result — a miniature version of `singleflight`.

## first-error-cancels
title: First Error Cancels the Rest
difficulty: advanced
module: concurrency
lesson: concurrency/cancellation
skill: concurrency

### Problem
Write `RunAll(ctx context.Context, tasks ...func(ctx context.Context) error) error` that runs all tasks concurrently. When one returns an error, cancel the context passed to the others. Wait for **all** tasks to return, then return the **first** error (or `nil`). If the parent `ctx` is cancelled, tasks see it as well.

### Constraints
- All tasks must have returned when `RunAll` returns
- Return the first error observed, not later `context.Canceled` errors from cancelled tasks
- Must be race-free

### Starter
```go
package main

import "context"

func RunAll(ctx context.Context, tasks ...func(ctx context.Context) error) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllSucceed(t *testing.T) {
	var n atomic.Int32
	task := func(ctx context.Context) error { n.Add(1); return nil }
	if err := RunAll(context.Background(), task, task, task); err != nil || n.Load() != 3 {
		t.Errorf("err=%v ran=%d", err, n.Load())
	}
	if err := RunAll(context.Background()); err != nil {
		t.Errorf("no tasks: %v", err)
	}
}

func TestFirstErrorCancelsOthers(t *testing.T) {
	boom := errors.New("boom")
	var cancelled atomic.Int32
	waiter := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			cancelled.Add(1)
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	}
	failing := func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return boom
	}
	start := time.Now()
	err := RunAll(context.Background(), waiter, failing, waiter)
	if !errors.Is(err, boom) {
		t.Errorf("want boom, got %v", err)
	}
	if cancelled.Load() != 2 {
		t.Errorf("%d waiters saw cancellation, want 2", cancelled.Load())
	}
	if time.Since(start) > time.Second {
		t.Error("RunAll did not cancel the other tasks promptly")
	}
}

func TestParentCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	task := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	if err := RunAll(ctx, task, task); !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}
```

### Hints
- `ctx, cancel := context.WithCancel(ctx)`; call `cancel()` on the first error
- Record the first error under a `sync.Once` or mutex so later `Canceled` errors don't overwrite it
- `wg.Wait()` before returning

### Solution
```go
package main

import (
	"context"
	"sync"
)

func RunAll(ctx context.Context, tasks ...func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	for _, task := range tasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task(ctx); err != nil {
				once.Do(func() {
					first = err
					cancel()
				})
			}
		}()
	}
	wg.Wait()
	return first
}

func main() {}
```

### Explanation
`sync.Once` records only the first error and triggers cancellation once. This is a simplified version of `errgroup.Group` from `golang.org/x/sync`.
