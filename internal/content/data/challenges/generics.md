# Generics Challenges

## generic-map-filter
title: Generic Slice Toolkit
difficulty: intermediate
module: generics
lesson: generics/generic-functions
skill: generics

### Problem
Write four generic helpers:

- `Map[T, U any](xs []T, f func(T) U) []U`
- `Filter[T any](xs []T, keep func(T) bool) []T`
- `Reduce[T, A any](xs []T, init A, f func(A, T) A) A`
- `GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T` — groups elements by key, preserving input order inside each group

Never return `nil` slices from `Map`/`Filter` (empty results are empty slices). Do not modify the input.

### Constraints
- Use generics; no `any` casts
- Preserve order

### Starter
```go
package main

func Map[T, U any](xs []T, f func(T) U) []U {
	return nil
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	return nil
}

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	return init
}

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, strconv.Itoa)
	if !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Errorf("got %v", got)
	}
	if got := Map([]int(nil), strconv.Itoa); got == nil || len(got) != 0 {
		t.Errorf("empty Map should be a non-nil empty slice, got %#v", got)
	}
}

func TestFilter(t *testing.T) {
	in := []string{"go", "", "rust", ""}
	got := Filter(in, func(s string) bool { return s != "" })
	if !reflect.DeepEqual(got, []string{"go", "rust"}) {
		t.Errorf("got %v", got)
	}
	if got := Filter([]int{1, 3}, func(n int) bool { return n%2 == 0 }); got == nil || len(got) != 0 {
		t.Errorf("no matches should be a non-nil empty slice, got %#v", got)
	}
	if len(in) != 4 {
		t.Error("input modified")
	}
}

func TestReduce(t *testing.T) {
	if got := Reduce([]int{1, 2, 3, 4}, 0, func(a, x int) int { return a + x }); got != 10 {
		t.Errorf("sum = %d", got)
	}
	got := Reduce([]string{"a", "b", "c"}, "", func(a string, x string) string { return a + strings.ToUpper(x) })
	if got != "ABC" {
		t.Errorf("concat = %q", got)
	}
	if got := Reduce([]int(nil), 42, func(a, x int) int { return a + x }); got != 42 {
		t.Errorf("empty reduce = %d", got)
	}
}

func TestGroupBy(t *testing.T) {
	words := []string{"apple", "avocado", "banana", "blueberry", "cherry", "apricot"}
	got := GroupBy(words, func(s string) byte { return s[0] })
	want := map[byte][]string{
		'a': {"apple", "avocado", "apricot"},
		'b': {"banana", "blueberry"},
		'c': {"cherry"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
	if got := GroupBy([]int(nil), func(n int) int { return n }); len(got) != 0 {
		t.Errorf("empty GroupBy: %v", got)
	}
}
```

### Hints
- Preallocate with `make([]U, 0, len(xs))`
- `GroupBy`: `out[k] = append(out[k], x)`

### Solution
```go
package main

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	out := make([]T, 0)
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, x := range xs {
		k := key(x)
		out[k] = append(out[k], x)
	}
	return out
}

func main() {}
```

### Explanation
Four tiny generic functions replace dozens of per-type loops. Type inference figures out `T`, `U`, `A` and `K` from the arguments.

## generic-stack
title: Generic Stack and Queue
difficulty: intermediate
module: generics
lesson: generics/generic-data-structures
skill: generics

### Problem
Implement two generic containers with useful zero values:

- `Stack[T]`: `Push(T)`, `Pop() (T, bool)`, `Peek() (T, bool)`, `Len() int`
- `Queue[T]`: `Enqueue(T)`, `Dequeue() (T, bool)`, `Len() int` (FIFO)

`Pop`, `Peek` and `Dequeue` return the zero value and `false` when empty.

### Constraints
- Zero values must be usable (`var s Stack[int]`)
- Works with any element type

### Starter
```go
package main

type Stack[T any] struct {
}

func (s *Stack[T]) Push(x T)        {}
func (s *Stack[T]) Pop() (T, bool)  { var z T; return z, false }
func (s *Stack[T]) Peek() (T, bool) { var z T; return z, false }
func (s *Stack[T]) Len() int        { return 0 }

type Queue[T any] struct {
}

func (q *Queue[T]) Enqueue(x T)        {}
func (q *Queue[T]) Dequeue() (T, bool) { var z T; return z, false }
func (q *Queue[T]) Len() int           { return 0 }

func main() {}
```

### Tests
```go
package main

import "testing"

func TestStack(t *testing.T) {
	var s Stack[string]
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty stack should fail")
	}
	s.Push("a")
	s.Push("b")
	if v, ok := s.Peek(); !ok || v != "b" || s.Len() != 2 {
		t.Errorf("Peek: %q %v len=%d", v, ok, s.Len())
	}
	if v, _ := s.Pop(); v != "b" {
		t.Errorf("Pop = %q", v)
	}
	if v, _ := s.Pop(); v != "a" {
		t.Errorf("Pop = %q", v)
	}
	if s.Len() != 0 {
		t.Error("stack should be empty")
	}
}

func TestStackOtherTypes(t *testing.T) {
	var s Stack[struct{ N int }]
	s.Push(struct{ N int }{7})
	if v, ok := s.Pop(); !ok || v.N != 7 {
		t.Errorf("%+v %v", v, ok)
	}
	var e Stack[*int]
	if v, ok := e.Pop(); ok || v != nil {
		t.Errorf("empty pointer stack: %v %v", v, ok)
	}
}

func TestQueueFIFO(t *testing.T) {
	var q Queue[int]
	if _, ok := q.Dequeue(); ok {
		t.Error("Dequeue on empty queue should fail")
	}
	for i := 1; i <= 5; i++ {
		q.Enqueue(i)
	}
	if q.Len() != 5 {
		t.Errorf("Len = %d", q.Len())
	}
	for want := 1; want <= 5; want++ {
		if got, ok := q.Dequeue(); !ok || got != want {
			t.Errorf("Dequeue = %d %v, want %d", got, ok, want)
		}
	}
	q.Enqueue(9)
	if got, _ := q.Dequeue(); got != 9 {
		t.Errorf("after reuse: %d", got)
	}
}

func TestQueueLargeIsEfficient(t *testing.T) {
	var q Queue[int]
	for i := 0; i < 200000; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < 200000; i++ {
		if got, _ := q.Dequeue(); got != i {
			t.Fatalf("got %d want %d", got, i)
		}
	}
}
```

### Hints
- A slice is enough for each; for the queue keep a `head` index instead of re-slicing from the front each time (amortised O(1))
- `var zero T` gives the empty return value

### Solution
```go
package main

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(x T) { s.items = append(s.items, x) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	x := s.items[len(s.items)-1]
	s.items[len(s.items)-1] = zero // let the GC reclaim it
	s.items = s.items[:len(s.items)-1]
	return x, true
}

func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int { return len(s.items) }

type Queue[T any] struct {
	items []T
	head  int
}

func (q *Queue[T]) Enqueue(x T) { q.items = append(q.items, x) }

func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.head >= len(q.items) {
		return zero, false
	}
	x := q.items[q.head]
	q.items[q.head] = zero
	q.head++
	if q.head > 32 && q.head*2 >= len(q.items) { // compact occasionally
		q.items = append([]T(nil), q.items[q.head:]...)
		q.head = 0
	}
	return x, true
}

func (q *Queue[T]) Len() int { return len(q.items) - q.head }

func main() {}
```

### Explanation
The stack uses the slice end; the queue tracks a head index and compacts occasionally so dequeue stays amortised O(1). Clearing removed slots lets the garbage collector free large elements.

## generic-lru-cache
title: Generic LRU Cache
difficulty: expert
module: generics
lesson: generics/generic-data-structures
skill: generics

### Problem
Implement a thread-safe **least-recently-used** cache: `LRU[K comparable, V any]`.

- `NewLRU[K comparable, V any](capacity int) *LRU[K, V]` (capacity < 1 is treated as 1)
- `Get(key K) (V, bool)` — a hit makes the key the most recently used
- `Put(key K, value V)` — inserts or updates and marks as most recently used; when full, evicts the least recently used entry
- `Len() int`
- `Keys() []K` — keys from **most** to **least** recently used

`Get` and `Put` must run in O(1) (use a map plus a doubly linked list).

### Constraints
- O(1) Get and Put
- Safe for concurrent use
- Updating an existing key must not grow the cache

### Starter
```go
package main

type LRU[K comparable, V any] struct {
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	var zero V
	return zero, false
}

func (c *LRU[K, V]) Put(key K, value V) {}

func (c *LRU[K, V]) Len() int { return 0 }

func (c *LRU[K, V]) Keys() []K { return nil }

func main() {}
```

### Tests
```go
package main

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestEvictsLeastRecentlyUsed(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 { // a is now most recent
		t.Fatalf("Get(a) = %d %v", v, ok)
	}
	c.Put("c", 3) // evicts b
	if _, ok := c.Get("b"); ok {
		t.Error("b should have been evicted")
	}
	if !reflect.DeepEqual(c.Keys(), []string{"c", "a"}) {
		t.Errorf("Keys = %v", c.Keys())
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d", c.Len())
	}
}

func TestUpdateDoesNotGrow(t *testing.T) {
	c := NewLRU[int, string](2)
	c.Put(1, "x")
	c.Put(2, "y")
	c.Put(1, "z") // update + refresh
	c.Put(3, "w") // evicts 2
	if v, ok := c.Get(1); !ok || v != "z" {
		t.Errorf("Get(1) = %q %v", v, ok)
	}
	if _, ok := c.Get(2); ok {
		t.Error("2 should be evicted")
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d", c.Len())
	}
}

func TestCapacityClamp(t *testing.T) {
	c := NewLRU[int, int](0)
	c.Put(1, 1)
	c.Put(2, 2)
	if c.Len() != 1 {
		t.Errorf("capacity 0 should behave as 1, Len = %d", c.Len())
	}
	if _, ok := c.Get(2); !ok {
		t.Error("most recent entry should remain")
	}
}

func TestMissReturnsZero(t *testing.T) {
	c := NewLRU[string, *int](3)
	if v, ok := c.Get("nope"); ok || v != nil {
		t.Errorf("miss: %v %v", v, ok)
	}
}

func TestLargeSequenceIsFast(t *testing.T) {
	c := NewLRU[int, int](1000)
	for i := 0; i < 300000; i++ {
		c.Put(i, i)
		c.Get(i - 500)
	}
	if c.Len() != 1000 {
		t.Errorf("Len = %d", c.Len())
	}
}

func TestConcurrent(t *testing.T) {
	c := NewLRU[string, int](50)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				k := fmt.Sprintf("k%d", (g*7+i)%80)
				c.Put(k, i)
				c.Get(k)
				c.Keys()
			}
		}()
	}
	wg.Wait()
	if c.Len() > 50 {
		t.Errorf("Len %d exceeds capacity", c.Len())
	}
}
```

### Hints
- `container/list` gives a doubly linked list; store `*list.Element` values in the map
- Put entries at the front on access; evict `list.Back()`
- Store key and value in each list element (a small struct) so eviction can delete the map entry

### Solution
```go
package main

import (
	"container/list"
	"sync"
)

type entry[K comparable, V any] struct {
	key K
	val V
}

type LRU[K comparable, V any] struct {
	mu    sync.Mutex
	cap   int
	ll    *list.List
	items map[K]*list.Element
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity < 1 {
		capacity = 1
	}
	return &LRU[K, V]{cap: capacity, ll: list.New(), items: make(map[K]*list.Element)}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		return el.Value.(*entry[K, V]).val, true
	}
	var zero V
	return zero, false
}

func (c *LRU[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		el.Value.(*entry[K, V]).val = value
		c.ll.MoveToFront(el)
		return
	}
	c.items[key] = c.ll.PushFront(&entry[K, V]{key, value})
	if c.ll.Len() > c.cap {
		oldest := c.ll.Back()
		c.ll.Remove(oldest)
		delete(c.items, oldest.Value.(*entry[K, V]).key)
	}
}

func (c *LRU[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *LRU[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]K, 0, c.ll.Len())
	for el := c.ll.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*entry[K, V]).key)
	}
	return keys
}

func main() {}
```

### Explanation
The map gives O(1) lookup of a list element; the linked list keeps recency order so moving to the front and evicting from the back are O(1) too. A mutex guards both structures, since a `Get` also mutates the order.
