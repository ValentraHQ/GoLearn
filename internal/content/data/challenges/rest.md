# REST API Challenges

## paginate
title: Paginate a Slice
difficulty: beginner
module: rest-api
lesson: rest-api/pagination
skill: rest

### Problem
Write a generic `Paginate[T any](items []T, page, size int) (Page[T], error)` returning one page of `items` plus metadata.

`Page[T]` has `Items []T`, `Page int`, `Size int`, `Total int` and `Pages int` (total number of pages, 0 when there are no items).

- `page` starts at 1; `page < 1` or `size < 1` → error
- `size` larger than 100 is capped at 100 (and reported as `Size: 100`)
- A page beyond the last returns empty `Items` (not an error) with correct metadata

### Constraints
- Use generics
- Do not mutate the input slice
- `Items` must be an empty (non-nil) slice when there is nothing to show

### Starter
```go
package main

type Page[T any] struct {
	Items []T
	Page  int
	Size  int
	Total int
	Pages int
}

func Paginate[T any](items []T, page, size int) (Page[T], error) {
	return Page[T]{}, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestPaginate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7}
	tests := []struct {
		page, size int
		want       Page[int]
	}{
		{1, 3, Page[int]{[]int{1, 2, 3}, 1, 3, 7, 3}},
		{3, 3, Page[int]{[]int{7}, 3, 3, 7, 3}},
		{4, 3, Page[int]{[]int{}, 4, 3, 7, 3}},
		{1, 100, Page[int]{items, 1, 100, 7, 1}},
		{1, 500, Page[int]{items, 1, 100, 7, 1}},
	}
	for _, tc := range tests {
		got, err := Paginate(items, tc.page, tc.size)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Paginate(page=%d,size=%d) = %+v, %v\nwant %+v", tc.page, tc.size, got, err, tc.want)
		}
	}
}

func TestPaginateEmptyAndErrors(t *testing.T) {
	got, err := Paginate[string](nil, 1, 10)
	if err != nil || got.Total != 0 || got.Pages != 0 || got.Items == nil || len(got.Items) != 0 {
		t.Errorf("empty: %+v %v", got, err)
	}
	for _, bad := range [][2]int{{0, 10}, {-1, 10}, {1, 0}, {1, -5}} {
		if _, err := Paginate([]int{1}, bad[0], bad[1]); err == nil {
			t.Errorf("page=%d size=%d should error", bad[0], bad[1])
		}
	}
}

func TestPaginateDoesNotAlias(t *testing.T) {
	items := []int{1, 2, 3}
	got, _ := Paginate(items, 1, 2)
	got.Items[0] = 99
	if items[0] != 1 {
		t.Error("returned page aliases the input slice")
	}
}
```

### Hints
- `start := (page-1)*size` and clamp `start`/`end` to `len(items)`
- `pages := (total + size - 1) / size`
- Copy the window into a new slice with `append([]T{}, items[start:end]...)`

### Solution
```go
package main

import "errors"

type Page[T any] struct {
	Items []T
	Page  int
	Size  int
	Total int
	Pages int
}

func Paginate[T any](items []T, page, size int) (Page[T], error) {
	if page < 1 || size < 1 {
		return Page[T]{}, errors.New("page and size must be at least 1")
	}
	if size > 100 {
		size = 100
	}
	total := len(items)
	start := min((page-1)*size, total)
	end := min(start+size, total)
	return Page[T]{
		Items: append([]T{}, items[start:end]...),
		Page:  page,
		Size:  size,
		Total: total,
		Pages: (total + size - 1) / size,
	}, nil
}

func main() {}
```

### Explanation
Clamping `start` and `end` to the length makes out-of-range pages naturally empty, and copying the window prevents callers from mutating the source data.

## token-bucket
title: Token Bucket Limiter
difficulty: advanced
module: rest-api
lesson: rest-api/rate-limiting
skill: rest

### Problem
Implement a token-bucket rate limiter with an **injectable clock**:

- `NewLimiter(rate float64, burst int, now func() time.Time) *Limiter` — refills `rate` tokens per second up to `burst`; the bucket starts full
- `Allow(key string) bool` — takes a token for that client key if available

Each key has its own bucket. Limiters must be safe for concurrent use.

### Constraints
- Use the injected `now` function, never `time.Now` directly
- Buckets are independent per key
- Race-free

### Starter
```go
package main

import "time"

type Limiter struct {
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{}
}

func (l *Limiter) Allow(key string) bool {
	return false
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

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func TestBurstThenRefill(t *testing.T) {
	c := &clock{time.Unix(1000, 0)}
	l := NewLimiter(2, 3, c.now) // 2 tokens/s, burst 3
	for i := 0; i < 3; i++ {
		if !l.Allow("a") {
			t.Fatalf("request %d within burst was denied", i)
		}
	}
	if l.Allow("a") {
		t.Error("4th immediate request should be denied")
	}
	c.t = c.t.Add(500 * time.Millisecond) // +1 token
	if !l.Allow("a") {
		t.Error("token should have refilled after 500ms")
	}
	if l.Allow("a") {
		t.Error("only one token should have been added")
	}
	c.t = c.t.Add(time.Hour) // refill is capped at burst
	n := 0
	for l.Allow("a") {
		n++
		if n > 10 {
			break
		}
	}
	if n != 3 {
		t.Errorf("after a long idle period got %d tokens, want burst=3", n)
	}
}

func TestKeysAreIndependent(t *testing.T) {
	c := &clock{time.Unix(0, 0)}
	l := NewLimiter(1, 1, c.now)
	if !l.Allow("a") || l.Allow("a") {
		t.Fatal("key a")
	}
	if !l.Allow("b") {
		t.Error("key b should have its own bucket")
	}
}

func TestConcurrentUse(t *testing.T) {
	c := &clock{time.Unix(0, 0)}
	l := NewLimiter(0.0001, 50, c.now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("shared") {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 50 {
		t.Errorf("allowed %d requests, want exactly 50", allowed.Load())
	}
}
```

### Hints
- Store `tokens` and `last` per key in a map guarded by a mutex
- On each call: `tokens = min(burst, tokens + elapsed.Seconds()*rate)`; then spend one if `tokens >= 1`
- Create a full bucket the first time a key is seen

### Solution
```go
package main

import (
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type Limiter struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	now     func() time.Time
	buckets map[string]*bucket
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: rate, burst: float64(burst), now: now, buckets: map[string]*bucket{}}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	t := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: t}
		l.buckets[key] = b
	}
	b.tokens = min(l.burst, b.tokens+t.Sub(b.last).Seconds()*l.rate)
	b.last = t
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func main() {}
```

### Explanation
Lazily refilling on each call avoids background goroutines: the elapsed time since the last call converts into tokens, capped at the burst size. Injecting the clock makes the behaviour deterministic to test.
